package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "spec.yaml")
	err := WriteFile(path, []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "data" {
		t.Fatalf("файл не записан: %v, %q", err, got)
	}
}

func TestWriteFile2(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	err := WriteFile(path, []byte("same"))
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	err = WriteFile(path, []byte("same"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Error("файл не должен перезаписываться при одинаковом содержимом")
	}
}

func TestWriteFile3(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	err := WriteFile(path, []byte("old"))
	if err != nil {
		t.Fatal(err)
	}
	err = WriteFile(path, []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "new" {
		t.Fatalf("ожидалось new, получено %q", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("временные файлы должны удаляться, в каталоге: %v", entries)
	}
}
