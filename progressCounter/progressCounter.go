package progressCounter

import (
	"os"
	"path/filepath"
)

func CountFiles(path string) (int, error) {
	count := 0
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() == ".DS_Store" {
		} else {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
