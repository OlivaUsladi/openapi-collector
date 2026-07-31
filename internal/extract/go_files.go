package extract

import (
	"os"
	"path/filepath"
	"strings"
)

// Функция сбора файлов в директории
func GoFiles(root string, includeTests bool) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{root}, nil
	}
	var files []string
	err1 := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") && !includeTests {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err1 != nil {
		return nil, err
	}
	return files, nil
}
