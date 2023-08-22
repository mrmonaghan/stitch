package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type File struct {
	Name     string
	Contents []byte
}

func ReadDirYaml(fsys fs.FS) ([]*File, error) {

	var files []*File

	yaml, err := fs.Glob(fsys, "*.yaml")
	if err != nil {
		return files, err
	}

	yml, err := fs.Glob(fsys, "*.yml")
	if err != nil {
		return files, err
	}

	fileNames := append(yaml, yml...)

	for _, fileName := range fileNames {

		b, err := fs.ReadFile(fsys, fileName)
		if err != nil {
			return files, err
		}

		f := &File{
			Name:     fileName,
			Contents: b,
		}

		files = append(files, f)
	}

	return files, nil
}

func ResolveDirInput(in string) (string, error) {
	abs, err := filepath.Abs(in)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return "", fmt.Errorf("directory '%s' does not exist", abs)
	}

	return abs, nil
}
