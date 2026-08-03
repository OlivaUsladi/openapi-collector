package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Функция сбора файлов в директории
// Dозвращает список Go-файлов внутри root
// excludes - glob-маски относительных путей, которые нужно пропустить
func GoFiles(root string, includeTests bool, excludes []string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !strings.HasSuffix(root, ".go") {
			return nil, fmt.Errorf("файл %q не является go файлом", root)
		}
		if strings.HasSuffix(root, "_test.go") && !includeTests {
			return nil, nil
		}
		return []string{filepath.Base(root)}, nil
	}
	var files []string
	err1 := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err2 := filepath.Rel(root, path)
		if err2 != nil {
			return err2
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if rel == "." {
				return nil
			}
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "testdata" ||
				strings.HasPrefix(name, ".") ||
				strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") && !includeTests {
			return nil
		}
		for _, p := range excludes {
			ok, err3 := filepath.Match(p, rel)
			if err3 != nil {
				return err3
			}
			if ok {
				return nil
			}
		}
		files = append(files, rel)
		return nil
	})
	if err1 != nil {
		return nil, err1
	}
	sort.Strings(files)
	return files, nil
}
