package atomicfile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// записывает данные в файл, но только если они изменились
func WriteFile(path string, data []byte) error {
	folder := filepath.Dir(path)
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return fmt.Errorf("не создалась папка %q: %w", folder, err)
	}

	oldData, err := os.ReadFile(path)
	if err == nil && bytes.Equal(oldData, data) {
		return nil
	}

	tempFile, err := os.CreateTemp(folder, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("не создался временный файл: %w", err)
	}
	tempName := tempFile.Name()

	_, err = tempFile.Write(data)
	if err != nil {
		tempFile.Close()
		os.Remove(tempName)
		return fmt.Errorf("не удалось записать во временный файл: %w", err)
	}

	err = tempFile.Close()
	if err != nil {
		os.Remove(tempName)
		return fmt.Errorf("не удалось закрыть временный файл: %w", err)
	}

	err = os.Rename(tempName, path)
	if err != nil {
		os.Remove(tempName)
		return fmt.Errorf("не удалось заменить файл %q: %w", path, err)
	}

	return nil
}
