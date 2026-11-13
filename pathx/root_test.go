package pathx

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRootFileLifecycle(t *testing.T) {
	rootDir := t.TempDir()
	root, err := OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if err := root.MkdirAll("documents/2026", 0o750); err != nil {
		t.Fatal(err)
	}
	file, err := root.OpenFile("documents/2026/report.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("original"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := root.Rename("documents/2026/report.txt", "documents/2026/final.txt"); err != nil {
		t.Fatal(err)
	}
	read, err := root.Open("documents/2026/final.txt")
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(read)
	if closeErr := read.Close(); err == nil {
		err = closeErr
	}
	if err != nil || string(content) != "original" {
		t.Fatalf("content=%q err=%v", content, err)
	}
	if err := root.Remove("documents/2026/final.txt"); err != nil {
		t.Fatal(err)
	}
	if err := root.Remove("documents/2026"); err != nil {
		t.Fatal(err)
	}
	if err := root.Remove("documents"); err != nil {
		t.Fatal(err)
	}
}

func TestRootRejectsSymbolicLinkEscape(t *testing.T) {
	rootDir, outsideDir := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(rootDir, "outside")
	if err := os.Symlink(outsideDir, link); err != nil {
		t.Skipf("symbolic links unavailable: %v", err)
	}
	root, err := OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	for _, operation := range []func() error{
		func() error { _, err := root.Open("outside/secret.txt"); return err },
		func() error { _, err := root.OpenFile("outside/new.txt", os.O_CREATE|os.O_WRONLY, 0o600); return err },
		func() error { return root.MkdirAll("outside/new-dir", 0o700) },
		func() error { return root.Remove("outside/secret.txt") },
		func() error { return root.Rename("outside/secret.txt", "renamed.txt") },
	} {
		if err := operation(); !errors.Is(err, ErrPathEscape) {
			t.Fatalf("operation accepted symbolic link escape: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "secret.txt")); err != nil {
		t.Fatalf("outside content changed: %v", err)
	}
}

func TestRootRejectsLinkedRootAndFinalLink(t *testing.T) {
	parent, outsideDir := t.TempDir(), t.TempDir()
	rootDir := filepath.Join(parent, "share")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rootAlias := filepath.Join(parent, "share-link")
	if err := os.Symlink(rootDir, rootAlias); err != nil {
		t.Skipf("symbolic links unavailable: %v", err)
	}
	if _, err := OpenRoot(rootAlias); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("linked root accepted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outsideDir, "secret.txt"), filepath.Join(rootDir, "final-link")); err != nil {
		t.Fatal(err)
	}
	root, err := OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := root.Open("final-link"); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("final link accepted: %v", err)
	}
}

func TestRootKeepsOpenedDirectoryBoundary(t *testing.T) {
	parent, outside := t.TempDir(), t.TempDir()
	rootPath := filepath.Join(parent, "share")
	if err := os.Mkdir(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	oldPath := filepath.Join(parent, "share-old")
	if err := os.Rename(rootPath, oldPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, rootPath); err != nil {
		t.Skipf("symbolic links unavailable: %v", err)
	}
	file, err := root.OpenFile("written-through-handle.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(oldPath, "written-through-handle.txt")); err != nil {
		t.Fatalf("opened root lost its boundary: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "written-through-handle.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("wrote through replaced root path: %v", err)
	}
}

func TestRootRejectsInvalidUsage(t *testing.T) {
	if _, err := OpenRoot(""); !errors.Is(err, ErrInvalidPath) {
		t.Fatal(err)
	}
	root, err := OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []func() error{
		func() error { _, err := root.Open("../escape"); return err },
		func() error { _, err := root.OpenFile("file", os.O_APPEND, 0o600); return err },
		func() error { return root.MkdirAll("new", 0) },
		func() error { return root.Rename("source", "../target") },
	} {
		if err := operation(); !errors.Is(err, ErrInvalidPath) && !errors.Is(err, ErrPathEscape) {
			t.Fatalf("unexpected validation error: %v", err)
		}
	}
	if err := root.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := root.Open("file"); !errors.Is(err, ErrRootClosed) {
		t.Fatal(err)
	}
}
