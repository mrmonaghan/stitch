package stitch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func getYamlFileNamesFromDir(dir string) ([]string, error) {

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return []string{}, fmt.Errorf("unable to get absolute filepath for dir '%s': %w", dir, err)
	}

	files := os.DirFS(absDir)

	var s []string
	err = fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			if ext == ".yaml" || ext == ".yml" {
				s = append(s, fmt.Sprintf("%s/%s", absDir, path))
			}
		}
		return nil
	})

	if err != nil {
		return s, err
	}

	return s, nil
}
