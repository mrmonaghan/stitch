package stitch

import (
	"io/fs"
	"testing/fstest"
	"time"
)

type testFile struct {
	Name     string
	Contents []byte
	Mode     fs.FileMode
}

func (t *testFile) MapFile() *fstest.MapFile {
	return &fstest.MapFile{
		ModTime: time.Now(),
		Data:    t.Contents,
		Mode:    t.Mode,
	}
}

func newMapFS(files []testFile) fstest.MapFS {
	m := make(map[string]*fstest.MapFile)

	for _, file := range files {
		m[file.Name] = file.MapFile()
	}

	return m
}
