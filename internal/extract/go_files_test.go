package extract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// создание пустых файлов в каталоге
func makeTree(t *testing.T, paths []string) string {
	root := t.TempDir()
	for _, p := range paths {
		full := filepath.Join(root, filepath.FromSlash(p))
		err := os.MkdirAll(filepath.Dir(full), 0o755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(full, []byte("package p\n"), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func checkGoFiles(t *testing.T, files []string, includeTests bool, excludes []string, expected []string) {
	root := makeTree(t, files)
	actual, err := GoFiles(root, includeTests, excludes)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(actual, ";") != strings.Join(expected, ";") {
		t.Errorf("ожидалось %v, получено %v", expected, actual)
	}
}

func TestGoFiles1(t *testing.T) {
	checkGoFiles(t, []string{"b.go", "a.go", "sub/c.go"}, false, nil, []string{"a.go", "b.go", "sub/c.go"})
}

func TestGoFiles2(t *testing.T) {
	checkGoFiles(t, []string{"a.go", "readme.md", "data.yaml"}, false, nil, []string{"a.go"})
}

func TestGoFiles3(t *testing.T) {
	checkGoFiles(t, []string{"a.go", "a_test.go"}, false, nil, []string{"a.go"})
}

func TestGoFiles4(t *testing.T) {
	checkGoFiles(t, []string{"a.go", "a_test.go"}, true, nil, []string{"a.go", "a_test.go"})
}

func TestGoFiles5(t *testing.T) {
	checkGoFiles(t, []string{"a.go", "vendor/v.go", "testdata/t.go", ".git/g.go", ".hidden/h.go", "_priv/p.go"},
		false, nil, []string{"a.go"})
}

func TestGoFiles6(t *testing.T) {
	checkGoFiles(t, []string{"a.go", "gen/x.go"}, false, []string{"gen/*.go"}, []string{"a.go"})
}

func TestGoFilesSingleFile(t *testing.T) {
	checkGoFiles(t, []string{"single.go"}, true, nil, []string{"single.go"})
}

func TestGoFilesMissing(t *testing.T) {
	_, err := GoFiles(filepath.Join(t.TempDir(), "no-dir"), false, nil)
	if err == nil {
		t.Fatal("ожидалась ошибка для несуществующего каталога")
	}
}
