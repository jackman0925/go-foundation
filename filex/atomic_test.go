package filex

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAtomicWrite(t *testing.T) {
	for _, opts := range []AtomicWriteOptions{{}, {Mode: 0o600, NoSync: true}, {CreateDir: true}} {
		dir := t.TempDir()
		dest := filepath.Join(dir, "file")
		if opts.CreateDir {
			dest = filepath.Join(dir, "nested", "file")
		}
		for _, value := range []string{"first", "replacement", ""} {
			if err := WriteAtomic(dest, []byte(value), opts); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(dest)
			if err != nil || string(got) != value {
				t.Fatalf("%q %v", got, err)
			}
		}
		info, err := os.Stat(dest)
		if err != nil {
			t.Fatal(err)
		}
		mode := opts.Mode
		if mode == 0 {
			mode = 0o644
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != mode {
			t.Fatal(info.Mode())
		}
		entries, err := os.ReadDir(filepath.Dir(dest))
		if err != nil || len(entries) != 1 {
			t.Fatalf("temp leak: %v %v", entries, err)
		}
	}
}

type brokenReader struct{ err error }

func (r brokenReader) Read(p []byte) (int, error) { copy(p, "partial"); return 7, r.err }

func TestAtomicReadFailure(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "file")
	if err := os.WriteFile(dest, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("read failed")
	n, err := CopyAtomic(dest, brokenReader{sentinel}, AtomicWriteOptions{})
	if n != 7 || !errors.Is(err, sentinel) {
		t.Fatalf("%d %v", n, err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "old" {
		t.Fatalf("%q %v", got, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatal("temporary file leaked")
	}
}

func TestAtomicInvalidAndMissingDirectory(t *testing.T) {
	if err := WriteAtomic("", nil, AtomicWriteOptions{}); err == nil {
		t.Fatal("empty destination")
	}
	if _, err := CopyAtomic("unused", nil, AtomicWriteOptions{}); err == nil {
		t.Fatal("nil reader")
	}
	if err := WriteAtomic("unused", nil, AtomicWriteOptions{Mode: fs.ModeDir}); err == nil {
		t.Fatal("bad mode")
	}
	if err := WriteAtomic(filepath.Join(t.TempDir(), "missing", "file"), nil, AtomicWriteOptions{}); err == nil {
		t.Fatal("missing parent")
	}
	dir := t.TempDir()
	if err := WriteAtomic(dir, []byte("new"), AtomicWriteOptions{}); err == nil {
		t.Fatal("replaced directory")
	}
	entries, err := os.ReadDir(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".atomic-") {
			t.Fatal("temp leaked")
		}
	}
}

type failingAtomicFile struct {
	File   *os.File
	stage  string
	err    error
	synced bool
}

func (f *failingAtomicFile) Name() string { return f.File.Name() }

func (f *failingAtomicFile) Write(p []byte) (int, error) {
	if f.stage == "write" {
		return 0, f.err
	}
	return f.File.Write(p)
}
func (f *failingAtomicFile) Chmod(m fs.FileMode) error {
	if f.stage == "chmod" {
		return f.err
	}
	return f.File.Chmod(m)
}
func (f *failingAtomicFile) Sync() error {
	f.synced = true
	if f.stage == "sync" {
		return f.err
	}
	return f.File.Sync()
}
func (f *failingAtomicFile) Close() error {
	err := f.File.Close()
	if f.stage == "close" {
		return f.err
	}
	return err
}

func TestAtomicInjectedFailures(t *testing.T) {
	for _, stage := range []string{"mkdir", "create", "write", "chmod", "sync", "close", "rename"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			dest := filepath.Join(dir, "file")
			if err := os.WriteFile(dest, []byte("old"), 0o600); err != nil {
				t.Fatal(err)
			}
			sentinel := errors.New(stage)
			ops := atomicOps{
				mkdir: func(p string, m fs.FileMode) error {
					if stage == "mkdir" {
						return sentinel
					}
					return os.MkdirAll(p, m)
				},
				create: func(p string) (atomicFile, error) {
					if stage == "create" {
						return nil, sentinel
					}
					f, err := os.CreateTemp(p, ".atomic-*")
					if err != nil {
						return nil, err
					}
					return &failingAtomicFile{File: f, stage: stage, err: sentinel}, nil
				},
				rename: func(a, b string) error {
					if stage == "rename" {
						return sentinel
					}
					return os.Rename(a, b)
				},
				remove: func(p string) error { _ = os.Remove(p); return errors.New("cleanup failure") },
			}
			_, err := copyAtomic(dest, io.LimitReader(strings.NewReader("new"), 3), AtomicWriteOptions{CreateDir: true}, ops)
			if !errors.Is(err, sentinel) {
				t.Fatalf("lost original error: %v", err)
			}
			got, err := os.ReadFile(dest)
			if err != nil || string(got) != "old" {
				t.Fatalf("old file changed: %q %v", got, err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatal("temp leaked")
			}
		})
	}
}

func TestAtomicSyncPolicy(t *testing.T) {
	for _, noSync := range []bool{false, true} {
		dir := t.TempDir()
		var temp *failingAtomicFile
		ops := atomicOps{
			create: func(p string) (atomicFile, error) {
				f, err := os.CreateTemp(p, ".atomic-*")
				if err != nil {
					return nil, err
				}
				temp = &failingAtomicFile{File: f}
				return temp, nil
			},
			mkdir: os.MkdirAll, rename: os.Rename, remove: os.Remove,
		}
		n, err := copyAtomic(filepath.Join(dir, "file"), strings.NewReader("new"), AtomicWriteOptions{NoSync: noSync}, ops)
		if err != nil || n != 3 {
			t.Fatalf("%d %v", n, err)
		}
		if temp.synced == noSync {
			t.Fatalf("NoSync=%v, synced=%v", noSync, temp.synced)
		}
	}
}
