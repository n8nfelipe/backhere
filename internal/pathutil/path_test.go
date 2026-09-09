package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNormalizePathRelativeAndAbsolute(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	rel, err := NormalizePath(".")
	if err != nil {
		t.Fatal(err)
	}
	abs, err := NormalizePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rel != abs {
		t.Fatalf("relative = %q, absolute = %q", rel, abs)
	}
}

func TestNormalizePathResolvesSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may require elevated privileges")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	a, err := NormalizePath(target)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NormalizePath(link)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("target = %q, link = %q", a, b)
	}
}
