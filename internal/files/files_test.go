package files

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

var testFiles = []*File{
	{
		Name:     "test-file-1.yaml",
		Contents: []byte(`test contents 1`),
	},
	{
		Name:     "test-file-2.yaml",
		Contents: []byte(`test contents 2`),
	},
	{
		Name:     "test-file-3.yml",
		Contents: []byte(`test contents 3`),
	},
	{
		Name:     "test-file-4.yml",
		Contents: []byte(`test contents 4`),
	},
}

func setupTestDir(dir string, subdirs ...string) error {
	for _, subdir := range subdirs {
		if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, subdir), fs.ModeDir); err != nil {
			return err
		}
	}
	return nil
}

func setupTestFS(files []*File, mode fs.FileMode) fstest.MapFS {

	m := make(map[string]*fstest.MapFile)

	for _, file := range files {
		mf := &fstest.MapFile{
			Data:    file.Contents,
			Mode:    mode,
			ModTime: time.Now(),
		}

		m[file.Name] = mf
	}

	return m
}

func TestReadDirYaml(t *testing.T) {

	tests := []struct {
		name    string
		fs      fs.FS
		want    []*File
		wantErr bool
	}{
		{
			name:    "happy path",
			fs:      setupTestFS(testFiles, fs.ModePerm),
			want:    testFiles,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadDirYaml(tt.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadDirYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadDirYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveDirInput(t *testing.T) {
	type args struct {
		in  string
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				in:  "./templates",
				dir: "templates",
			},
		},
		{
			name: "dir missing",
			args: args{
				in:  "./templates",
				dir: "not-there",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dir := t.TempDir()
			if err := setupTestDir(dir, tt.args.dir); err != nil {
				t.Error("ResolveDirInput() error: unable to set up test dir")
			}

			want := fmt.Sprintf("%s/%s", dir, tt.args.dir)

			err := os.Chdir(dir)
			if err != nil {
				t.Errorf("ResolveDirInput() error: unable to enter tempDir %s", dir)
			}

			got, err := ResolveDirInput(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveDirInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != want {
				t.Errorf("ResolveDirInput() = %v, want %v", got, want)
			}
		})
	}
}
