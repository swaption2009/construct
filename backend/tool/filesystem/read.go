package filesystem

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/furisto/construct/backend/tool/base"
)

type ReadFileInput struct {
	Path string `json:"path"`
}

type ReadFileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func ReadFile(fsys afero.Fs, input *ReadFileInput) (*ReadFileResult, error) {
	if input.Path == "" {
		return nil, base.NewCustomError("path is required", []string{
			"Please provide a valid path to the file you want to read",
		})
	}

	if !filepath.IsAbs(input.Path) {
		return nil, base.NewCustomError("path must be absolute", []string{
			"Please provide a valid absolute path to the file you want to read",
		})
	}
	path := input.Path

	stat, err := fsys.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, base.NewError(base.FileNotFound, "path", path)
		}
		if os.IsPermission(err) {
			return nil, base.NewError(base.PermissionDenied, "path", path)
		}
		return nil, base.NewError(base.CannotStatFile, "path", path)
	}

	if stat.IsDir() {
		return nil, base.NewError(base.PathIsDirectory, "path", path)
	}

	file, err := fsys.Open(path)
	if err != nil {
		return nil, base.NewCustomError("error reading file", []string{
			"Verify that you have the permission to read the file",
		}, "path", path, "error", err)
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		if lineNumber > 1 {
			builder.WriteByte('\n')
		}
		builder.WriteString(strconv.Itoa(lineNumber))
		builder.WriteString(": ")
		builder.WriteString(line)
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, base.NewCustomError("error reading file", []string{
			"Verify that you have the permission to read the file",
		}, "path", path, "error", err)
	}

	return &ReadFileResult{
		Path:    path,
		Content: builder.String(),
	}, nil
}
