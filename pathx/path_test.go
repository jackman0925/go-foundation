package pathx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeURLPath(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", "/"}, {"/", "/"}, {"a//./b", "/a/b"}, {"中文/%20文件", "/中文/ 文件"},
		{"a\\b", "/a/b"}, {"/%252e%252e/file", "/%2e%2e/file"}, {"/a+b", "/a+b"},
	} {
		got, err := NormalizeURLPath(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("%q: %q %v", tc.in, got, err)
		}
	}
	for _, raw := range []string{"%", "%GG", "%00", "%ff", "C:/file", "/NUL.txt", "/bad?"} {
		if _, err := NormalizeURLPath(raw); !errors.Is(err, ErrInvalidPath) {
			t.Errorf("%q: %v", raw, err)
		}
	}
	for _, raw := range []string{"../x", "/a/../b", "/%2e%2e/x", "/a\\..\\x", "/%2e%2e%5cx"} {
		if _, err := NormalizeURLPath(raw); !errors.Is(err, ErrPathEscape) {
			t.Errorf("%q: %v", raw, err)
		}
	}
}

func TestRelativePaths(t *testing.T) {
	for _, raw := range []string{"", "/", ".", "..", "../x", "a/../b", "..\\x", "/abs", "C:x", "C:/x", "\\\\host\\share", "\\?\\C:\\x", "a\x00b", "a:b", "a.", "a ", "CON", "nul.txt", "COM1.txt", "LPT²", "AUX", "CONIN$", "bad*", "bad\x7f", "\xff"} {
		if _, err := SanitizeRelativePath(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, tc := range []struct{ in, want string }{
		{"中文/ 文件.txt", "中文/ 文件.txt"}, {"a\\b//./c", "a/b/c"},
		{"%2e%2e/file", "%2e%2e/file"}, {"COM12.txt", "COM12.txt"}, {"a%file", "a%file"},
	} {
		got, err := SanitizeRelativePath(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("%q: %q %v", tc.in, got, err)
		}
	}
}

func TestResolveUnderRoot(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"", "."} {
		got, err := ResolveUnderRoot(root, rel)
		if err != nil || got != root {
			t.Fatalf("%q %v", got, err)
		}
	}
	got, err := ResolveUnderRoot(root, "子目录/file")
	if err != nil || got != filepath.Join(root, "子目录", "file") {
		t.Fatalf("%q %v", got, err)
	}
	for _, rel := range []string{"../" + filepath.Base(root) + "-old/file", "/absolute", "C:\\x"} {
		if _, err := ResolveUnderRoot(root, rel); err == nil {
			t.Fatal(rel)
		}
	}
	for _, root := range []string{"", "a\x00b", "\xff"} {
		if _, err := ResolveUnderRoot(root, "file"); !errors.Is(err, ErrInvalidPath) {
			t.Fatal(err)
		}
	}
}

func TestResolveIsLexicalNotSymlinkProtection(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	got, err := ResolveUnderRoot(root, "link")
	if err != nil {
		t.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(outside)
	if err != nil || real != want {
		t.Fatalf("%s %s %v", real, want, err)
	}
}
