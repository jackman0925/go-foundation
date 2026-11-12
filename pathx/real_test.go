package pathx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeRealTestFile(t *testing.T, name string) {
	t.Helper()
	if err := os.WriteFile(name, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func makeTestLink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symbolic links unavailable: %v", err)
	}
}

func TestResolveRealOrdinaryAndErrors(t *testing.T) {
	root := t.TempDir()
	writeRealTestFile(t, filepath.Join(root, "文件.txt"))
	for _, rel := range []string{"", ".", "文件.txt"} {
		got, err := ResolveRealUnderRoot(root, rel)
		want, wantErr := filepath.EvalSymlinks(filepath.Join(root, rel))
		if err != nil || wantErr != nil || got != want {
			t.Fatalf("%q: %q %v, want %q", rel, got, err, want)
		}
	}
	for _, tc := range []struct {
		root, rel string
		want      error
	}{
		{"", "file", ErrInvalidPath},
		{root, "../escape", ErrPathEscape},
		{root, "/absolute", ErrInvalidPath},
		{root, "missing", os.ErrNotExist},
		{filepath.Join(root, "missing"), "", os.ErrNotExist},
		{filepath.Join(root, "文件.txt"), "", ErrInvalidPath},
	} {
		got, err := ResolveRealUnderRoot(tc.root, tc.rel)
		if got != "" || !errors.Is(err, tc.want) {
			t.Fatalf("%q %q: %q %v", tc.root, tc.rel, got, err)
		}
	}
}

func TestResolveRealLinks(t *testing.T) {
	parent := t.TempDir()
	root, outside := filepath.Join(parent, "share"), filepath.Join(parent, "share-old")
	for _, dir := range []string{root, outside, filepath.Join(root, "sub")} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeRealTestFile(t, filepath.Join(root, "sub", "file"))
	writeRealTestFile(t, filepath.Join(outside, "file"))
	makeTestLink(t, "sub", filepath.Join(root, "inside"))
	makeTestLink(t, "inside", filepath.Join(root, "chain"))
	makeTestLink(t, outside, filepath.Join(root, "outside"))
	makeTestLink(t, filepath.Join(outside, "file"), filepath.Join(root, "outside-file"))
	makeTestLink(t, "missing", filepath.Join(root, "dangling"))
	makeTestLink(t, "loop", filepath.Join(root, "loop"))
	for _, rel := range []string{"inside/file", "chain/file"} {
		got, err := ResolveRealUnderRoot(root, rel)
		want, _ := filepath.EvalSymlinks(filepath.Join(root, "sub", "file"))
		if err != nil || got != want {
			t.Fatalf("%s: %s %v", rel, got, err)
		}
	}
	for _, rel := range []string{"outside", "outside/file", "outside-file"} {
		got, err := ResolveRealUnderRoot(root, rel)
		if got != "" || !errors.Is(err, ErrPathEscape) {
			t.Fatalf("%s: %s %v", rel, got, err)
		}
	}
	if _, err := ResolveRealUnderRoot(root, "dangling"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := ResolveRealUnderRoot(root, "loop"); err == nil {
		t.Fatal("accepted loop")
	}
	alias := filepath.Join(parent, "alias")
	makeTestLink(t, root, alias)
	got, err := ResolveRealUnderRoot(alias, "sub/file")
	want, _ := filepath.EvalSymlinks(filepath.Join(root, "sub", "file"))
	if err != nil || got != want {
		t.Fatalf("root alias: %s %v", got, err)
	}
}

func TestResolveRealDoesNotGuaranteeLaterAccess(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	dir := filepath.Join(root, "sub")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveRealUnderRoot(root, "sub")
	if err != nil {
		t.Fatal(err)
	}
	// 验证返回路径不是目录句柄：后续替换仍会改变其指向。
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	makeTestLink(t, outside, dir)
	actual, err := filepath.EvalSymlinks(got)
	want, _ := filepath.EvalSymlinks(outside)
	if err != nil || actual != want {
		t.Fatalf("%s %v", actual, err)
	}
}
