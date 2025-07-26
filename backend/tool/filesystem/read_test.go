package filesystem

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/furisto/construct/backend/tool/base"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
)

func TestReadFile(t *testing.T) {
	t.Parallel()

	setup := &base.ToolTestSetup[*ReadFileInput, *ReadFileResult]{
		Call: func(ctx context.Context, services *base.ToolTestServices, input *ReadFileInput) (*ReadFileResult, error) {
			return ReadFile(services.FS, input)
		},
		CmpOptions: []cmp.Option{
			cmpopts.IgnoreFields(base.ToolError{}, "Suggestions"),
		},
	}

	setup.RunToolTests(t, []base.ToolTestScenario[*ReadFileInput, *ReadFileResult]{
		{
			Name:           "successful read of text file",
			TestInput:      &ReadFileInput{Path: "/workspace/test.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/test.txt",
					Content: "1: Hello, World!\n2: This is a test file.",
				},
			},
		},
		{
			Name:           "successful read of empty file",
			TestInput:      &ReadFileInput{Path: "/workspace/empty.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/empty.txt",
					Content: "",
				},
			},
		},
		{
			Name:           "successful read of JSON file",
			TestInput:      &ReadFileInput{Path: "/workspace/config.json"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/config.json",
					Content: `1: {"name": "test", "version": "1.0.0"}`,
				},
			},
		},
		{
			Name:           "successful read of code file",
			TestInput:      &ReadFileInput{Path: "/workspace/src/main.go"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/src/main.go",
					Content: `1: package main
2: 
3: import "fmt"
4: 
5: func main() {
6: 	fmt.Println("Hello, World!")
7: }`,
				},
			},
		},
		{
			Name:           "successful read of binary file",
			TestInput:      &ReadFileInput{Path: "/workspace/binary.bin"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/binary.bin",
					Content: "1: " + string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF}),
				},
			},
		},
		{
			Name:           "file not found",
			TestInput:      &ReadFileInput{Path: "/workspace/nonexistent.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewError(base.FileNotFound, "path", "/workspace/nonexistent.txt"),
			},
		},
		{
			Name:           "file not found in nested directory",
			TestInput:      &ReadFileInput{Path: "/workspace/deep/nested/missing.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewError(base.FileNotFound, "path", "/workspace/deep/nested/missing.txt"),
			},
		},
		{
			Name:           "directory instead of file",
			TestInput:      &ReadFileInput{Path: "/workspace/src"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewError(base.PathIsDirectory, "path", "/workspace/src"),
			},
		},
		{
			Name:           "file with special characters in name",
			TestInput:      &ReadFileInput{Path: "/workspace/special-file_with@symbols.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/special-file_with@symbols.txt",
					Content: "1: File with special characters in name",
				},
			},
		},
		{
			Name:           "file with unicode content",
			TestInput:      &ReadFileInput{Path: "/workspace/unicode.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/unicode.txt",
					Content: "1: Hello 世界! 🌍 Здравствуй мир! ¡Hola mundo!",
				},
			},
		},
		{
			Name:           "large file content",
			TestInput:      &ReadFileInput{Path: "/workspace/large.txt"},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/large.txt",
					Content: strings.TrimRight(generateLargeContent(true), "\n"),
				},
			},
		},
	})
}

func seedTestFilesystem(ctx context.Context, fs afero.Fs) {
	fs.MkdirAll("/workspace/src", 0755)
	fs.MkdirAll("/workspace/deep/nested", 0755)

	afero.WriteFile(fs, "/workspace/test.txt", []byte("Hello, World!\nThis is a test file."), 0644)
	afero.WriteFile(fs, "/workspace/empty.txt", []byte(""), 0644)
	afero.WriteFile(fs, "/workspace/config.json", []byte(`{"name": "test", "version": "1.0.0"}`), 0644)
	afero.WriteFile(fs, "/workspace/src/main.go", []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`), 0644)
	afero.WriteFile(fs, "/workspace/binary.bin", []byte{0x00, 0x01, 0x02, 0x03, 0xFF}, 0644)
	afero.WriteFile(fs, "/workspace/special-file_with@symbols.txt", []byte("File with special characters in name"), 0644)
	afero.WriteFile(fs, "/workspace/unicode.txt", []byte("Hello 世界! 🌍 Здравствуй мир! ¡Hola mundo!"), 0644)
	afero.WriteFile(fs, "/workspace/large.txt", []byte(generateLargeContent(false)), 0644)
}

func generateLargeContent(lineNumbers bool) string {
	content := ""
	for i := 0; i < 1000; i++ {
		if lineNumbers {
			content += strconv.Itoa(i+1) + ": "
		}
		content += "This is line " + string(rune(i+'0')) + " of a large file for testing purposes.\n"
	}
	return content
}
