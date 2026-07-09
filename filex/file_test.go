package filex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExistsIsFileAndIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !Exists(file) {
		t.Fatal("expected file to exist")
	}
	if !IsFile(file) {
		t.Fatal("expected file path")
	}
	if IsDir(file) {
		t.Fatal("expected file not directory")
	}
	if !IsDir(dir) {
		t.Fatal("expected directory path")
	}
	if Exists(filepath.Join(dir, "missing.txt")) {
		t.Fatal("expected missing path to not exist")
	}
}

func TestEnsureDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b")

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir returned error: %v", err)
	}
	if !IsDir(dir) {
		t.Fatal("expected directory to be created")
	}
}

func TestReadTextAndWriteText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "demo.txt")

	if err := WriteText(path, "你好", 0o600); err != nil {
		t.Fatalf("WriteText returned error: %v", err)
	}

	text, err := ReadText(path)
	if err != nil {
		t.Fatalf("ReadText returned error: %v", err)
	}
	if text != "你好" {
		t.Fatalf("expected text, got %q", text)
	}
}

func TestReadTextReturnsErrorForMissingFile(t *testing.T) {
	if _, err := ReadText(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestWriteTextReturnsErrorWhenParentIsFile(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "parent")
	if err := os.WriteFile(parent, []byte("not dir"), 0o600); err != nil {
		t.Fatalf("write parent file: %v", err)
	}

	if err := WriteText(filepath.Join(parent, "demo.txt"), "hello", 0o600); err == nil {
		t.Fatal("expected parent file error")
	}
}

func TestCopyFileCopiesContentAndMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "nested", "dst.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o640); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile returned error: %v", err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected content: %q", content)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("expected mode 0640, got %o", info.Mode().Perm())
	}
}

func TestCopyFileRejectsDirectorySource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "dst.txt")
	if err := CopyFile(t.TempDir(), dst); err == nil {
		t.Fatal("expected directory source error")
	}
}

func TestCopyFileReturnsErrorForMissingSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "dst.txt")
	if err := CopyFile(filepath.Join(t.TempDir(), "missing.txt"), dst); err == nil {
		t.Fatal("expected missing source error")
	}
}

func TestFileSizeExtAndBaseName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "demo.tar.gz")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	size, err := FileSize(path)
	if err != nil {
		t.Fatalf("FileSize returned error: %v", err)
	}
	if size != 5 {
		t.Fatalf("expected size 5, got %d", size)
	}
	if Ext(path) != ".gz" {
		t.Fatalf("expected .gz, got %q", Ext(path))
	}
	if BaseName(path) != "demo.tar.gz" {
		t.Fatalf("unexpected base name: %q", BaseName(path))
	}
}

func TestFileSizeRejectsDirectory(t *testing.T) {
	if _, err := FileSize(t.TempDir()); err == nil {
		t.Fatal("expected directory error")
	}
}
