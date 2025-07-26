package filesystem

import (
	"path/filepath"

	"github.com/furisto/construct/backend/tool/base"
	"github.com/spf13/afero"
)

type CreateFileInput struct {
	Path    string
	Content string
}

type CreateFileResult struct {
	Overwritten bool `json:"overwritten"`
}

func CreateFile(fsys afero.Fs, input *CreateFileInput) (*CreateFileResult, error) {
	if !filepath.IsAbs(input.Path) {
		return nil, base.NewError(base.PathIsNotAbsolute, "path", input.Path)
	}
	path := input.Path

	var existed bool
	if stat, err := fsys.Stat(path); err == nil {
		if stat.IsDir() {
			return nil, base.NewError(base.PathIsDirectory, "path", path)
		}
		existed = true
	}

	err := fsys.MkdirAll(filepath.Dir(path), 0644)
	if err != nil {
		return nil, base.NewCustomError("could not create the parent directory", []string{
			"Verify that you have the permissions to create the parent directories",
			"Create the missing parent directories manually",
		}, "path", path, "error", err)
	}

	err = afero.WriteFile(fsys, path, []byte(input.Content), 0644)
	if err != nil {
		return nil, base.NewCustomError("error writing file", []string{
			"Ensure that you have the permission to write to the file",
		}, "path", path, "error", err)
	}

	return &CreateFileResult{Overwritten: existed}, nil
}
