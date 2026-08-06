package extract

import (
	"fmt"
	"openapi-collector/internal/model"
	"os"
	"path/filepath"
)

// собирает все @openapi-фрагменты из файла или каталога source
func JoinOpenApi(source string, includeTests bool, excludes []string) ([]model.Fragment, []error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, []error{fmt.Errorf("ошибка в os.Stat %q: %w", source, err)}
	}

	var dir string
	if !info.IsDir() {
		dir = filepath.Dir(source)
	} else {
		dir = source
	}

	files, err := GoFiles(source, includeTests, excludes)
	if err != nil {
		return nil, []error{err}
	}

	var fr []model.Fragment
	var errs []error
	for _, rel := range files {
		diskPath := filepath.Join(dir, filepath.FromSlash(rel))
		src, err := os.ReadFile(diskPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("чтение файла %s: %w", diskPath, err))
			continue
		}

		fileFragments, err := FindApiComment(rel, diskPath, src)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fr = append(fr, fileFragments...)
	}
	return fr, errs
}
