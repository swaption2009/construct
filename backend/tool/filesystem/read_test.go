package filesystem

import (
	"context"
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
					Content: "Hello, World!\nThis is a test file.",
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
					Content: `{"name": "test", "version": "1.0.0"}`,
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
					Content: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
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
					Content: string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF}),
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
					Content: "File with special characters in name",
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
					Content: "Hello 世界! 🌍 Здравствуй мир! ¡Hola mundo!",
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
					Content: strings.TrimRight(generateLargeContent(), "\n"),
				},
			},
		},

		// New test cases for line range functionality
		{
			Name: "read specific line range - middle of file",
			TestInput: &ReadFileInput{
				Path:      "/workspace/src/main.go",
				StartLine: intPtr(3),
				EndLine:   intPtr(5),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/src/main.go",
					Content: `// skipped 2 lines
import "fmt"

func main() {
// 2 lines remaining`,
				},
			},
		},
		{
			Name: "read from beginning with end line",
			TestInput: &ReadFileInput{
				Path:    "/workspace/test.txt",
				EndLine: intPtr(1),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/test.txt",
					Content: "Hello, World!\n// 1 lines remaining",
				},
			},
		},
		{
			Name: "read from start line to end of file",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(2),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/test.txt",
					Content: `// skipped 1 lines
This is a test file.`,
				},
			},
		},
		{
			Name: "read single line",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(1),
				EndLine:   intPtr(1),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/test.txt",
					Content: `Hello, World!
// 1 lines remaining`,
				},
			},
		},
		{
			Name: "read single line in middle",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(2),
				EndLine:   intPtr(2),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/test.txt",
					Content: `// skipped 1 lines
This is a test file.
// 0 lines remaining`,
				},
			},
		},
		{
			Name: "read range beyond file length",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(1),
				EndLine:   intPtr(10),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path: "/workspace/test.txt",
					Content: `Hello, World!
This is a test file.
// 0 lines remaining`,
				},
			},
		},
		{
			Name: "start line beyond file length",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(10),
				EndLine:   intPtr(15),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/test.txt",
					Content: "",
				},
			},
		},
		{
			Name: "read empty file with range",
			TestInput: &ReadFileInput{
				Path:      "/workspace/empty.txt",
				StartLine: intPtr(1),
				EndLine:   intPtr(5),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Result: &ReadFileResult{
					Path:    "/workspace/empty.txt",
					Content: "",
				},
			},
		},
		// Error cases for line range validation
		{
			Name: "negative start line",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(-1),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewCustomError("start_line must be positive", []string{
					"Please provide a start_line value of 1 or greater",
				}),
			},
		},
		{
			Name: "negative end line",
			TestInput: &ReadFileInput{
				Path:    "/workspace/test.txt",
				EndLine: intPtr(-1),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewCustomError("end_line must be positive", []string{
					"Please provide an end_line value of 1 or greater",
				}),
			},
		},
		{
			Name: "start line greater than end line",
			TestInput: &ReadFileInput{
				Path:      "/workspace/test.txt",
				StartLine: intPtr(5),
				EndLine:   intPtr(3),
			},
			SeedFilesystem: seedTestFilesystem,
			Expected: base.ToolTestExpectation[*ReadFileResult]{
				Error: base.NewCustomError("start_line must be less than or equal to end_line", []string{
					"Please ensure start_line <= end_line",
				}),
			},
		},
	})
}

func intPtr(i int) *int {
	return &i
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
	afero.WriteFile(fs, "/workspace/large.txt", []byte(generateLargeContent()), 0644)
}

func generateLargeContent() string {
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is line " + string(rune(i+'0')) + " of a large file for testing purposes.\n"
	}
	return content
}
