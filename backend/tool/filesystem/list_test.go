package filesystem

import (
	"context"
	"testing"

	"github.com/furisto/construct/backend/tool/base"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
)

func TestListFiles(t *testing.T) {
	t.Parallel()

	setup := &base.ToolTestSetup[*ListFilesInput, *ListFilesResult]{
		Call: func(ctx context.Context, services *base.ToolTestServices, input *ListFilesInput) (*ListFilesResult, error) {
			return ListFiles(services.FS, input)
		},
		CmpOptions: []cmp.Option{
			cmpopts.IgnoreFields(base.ToolError{}, "Suggestions"),
			cmpopts.SortSlices(func(a, b DirectoryEntry) bool {
				return a.Name < b.Name
			}),
		},
	}

	setup.RunToolTests(t, []base.ToolTestScenario[*ListFilesInput, *ListFilesResult]{
		{
			Name: "path is not absolute",
			TestInput: &ListFilesInput{
				Path:      "relative/path",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Error: base.NewError(base.PathIsNotAbsolute, "path", "relative/path"),
			},
		},
		{
			Name: "directory does not exist",
			TestInput: &ListFilesInput{
				Path:      "/nonexistent",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Error: base.NewError(base.DirectoryNotFound, "path", "/nonexistent"),
			},
		},
		{
			Name: "path is not a directory",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				afero.WriteFile(fs, "/file.txt", []byte("content"), 0644)
			},
			TestInput: &ListFilesInput{
				Path:      "/file.txt",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Error: base.NewError(base.PathIsNotDirectory, "path", "/file.txt"),
			},
		},
		{
			Name: "empty directory non-recursive",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/empty", 0755)
			},
			TestInput: &ListFilesInput{
				Path:      "/empty",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path:    "/empty",
					Entries: []DirectoryEntry{},
				},
			},
		},
		{
			Name: "empty directory recursive",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/empty", 0755)
			},
			TestInput: &ListFilesInput{
				Path:      "/empty",
				Recursive: true,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path:    "/empty",
					Entries: []DirectoryEntry{},
				},
			},
		},
		{
			Name: "directory with files non-recursive",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/workspace", 0755)
				afero.WriteFile(fs, "/workspace/file1.txt", []byte("content1"), 0644)
				afero.WriteFile(fs, "/workspace/file2.js", []byte("console.log('hello')"), 0644)
				fs.MkdirAll("/workspace/subdir", 0755)
			},
			TestInput: &ListFilesInput{
				Path:      "/workspace",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path: "/workspace",
					Entries: []DirectoryEntry{
						{Name: "/workspace/file1.txt", Type: "f", Size: 1},
						{Name: "/workspace/file2.js", Type: "f", Size: 1},
						{Name: "/workspace/subdir", Type: "d", Size: 0},
					},
				},
			},
		},
		{
			Name: "directory with files recursive",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/workspace/src/components", 0755)
				fs.MkdirAll("/workspace/docs", 0755)

				afero.WriteFile(fs, "/workspace/README.md", []byte("# Project"), 0644)
				afero.WriteFile(fs, "/workspace/src/main.js", []byte("console.log('main')"), 0644)
				afero.WriteFile(fs, "/workspace/src/components/Button.js", []byte("export default function Button() {}"), 0644)
				afero.WriteFile(fs, "/workspace/docs/guide.md", []byte("## Getting Started"), 0644)
			},
			TestInput: &ListFilesInput{
				Path:      "/workspace",
				Recursive: true,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path: "/workspace",
					Entries: []DirectoryEntry{
						{Name: "/workspace/README.md", Type: "f", Size: 1},
						{Name: "/workspace/docs", Type: "d", Size: 0},
						{Name: "/workspace/docs/guide.md", Type: "f", Size: 1},
						{Name: "/workspace/src", Type: "d", Size: 0},
						{Name: "/workspace/src/components", Type: "d", Size: 0},
						{Name: "/workspace/src/components/Button.js", Type: "f", Size: 1},
						{Name: "/workspace/src/main.js", Type: "f", Size: 1},
					},
				},
			},
		},
		{
			Name: "large file size calculation",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/test", 0755)
				// Create a file that's exactly 2048 bytes (2 KB)
				content := make([]byte, 2048)
				for i := range content {
					content[i] = 'a'
				}
				afero.WriteFile(fs, "/test/large.txt", content, 0644)

				// Create a file that's 1500 bytes (should round up to 2 KB)
				content2 := make([]byte, 1500)
				for i := range content2 {
					content2[i] = 'b'
				}
				afero.WriteFile(fs, "/test/medium.txt", content2, 0644)
			},
			TestInput: &ListFilesInput{
				Path:      "/test",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path: "/test",
					Entries: []DirectoryEntry{
						{Name: "/test/large.txt", Type: "f", Size: 2},
						{Name: "/test/medium.txt", Type: "f", Size: 2},
					},
				},
			},
		},
		{
			Name: "nested directory structure recursive",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				fs.MkdirAll("/project/src/utils/helpers", 0755)
				fs.MkdirAll("/project/tests/unit", 0755)
				fs.MkdirAll("/project/tests/integration", 0755)

				afero.WriteFile(fs, "/project/package.json", []byte(`{"name": "test"}`), 0644)
				afero.WriteFile(fs, "/project/src/index.js", []byte("// main file"), 0644)
				afero.WriteFile(fs, "/project/src/utils/helpers/format.js", []byte("export const format = () => {}"), 0644)
				afero.WriteFile(fs, "/project/tests/unit/test1.js", []byte("test('unit test')"), 0644)
				afero.WriteFile(fs, "/project/tests/integration/test2.js", []byte("test('integration test')"), 0644)
			},
			TestInput: &ListFilesInput{
				Path:      "/project",
				Recursive: true,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path: "/project",
					Entries: []DirectoryEntry{
						{Name: "/project/package.json", Type: "f", Size: 1},
						{Name: "/project/src", Type: "d", Size: 0},
						{Name: "/project/src/index.js", Type: "f", Size: 1},
						{Name: "/project/src/utils", Type: "d", Size: 0},
						{Name: "/project/src/utils/helpers", Type: "d", Size: 0},
						{Name: "/project/src/utils/helpers/format.js", Type: "f", Size: 1},
						{Name: "/project/tests", Type: "d", Size: 0},
						{Name: "/project/tests/integration", Type: "d", Size: 0},
						{Name: "/project/tests/integration/test2.js", Type: "f", Size: 1},
						{Name: "/project/tests/unit", Type: "d", Size: 0},
						{Name: "/project/tests/unit/test1.js", Type: "f", Size: 1},
					},
				},
			},
		},
		{
			Name: "root directory with content",
			SeedFilesystem: func(ctx context.Context, fs afero.Fs) {
				afero.WriteFile(fs, "/root.txt", []byte("root file"), 0644)
				fs.MkdirAll("/bin", 0755)
				afero.WriteFile(fs, "/bin/executable", []byte("#!/bin/bash\necho hello"), 0755)
			},
			TestInput: &ListFilesInput{
				Path:      "/",
				Recursive: false,
			},
			Expected: base.ToolTestExpectation[*ListFilesResult]{
				Result: &ListFilesResult{
					Path: "/",
					Entries: []DirectoryEntry{
						{Name: "/bin", Type: "d", Size: 0},
						{Name: "/root.txt", Type: "f", Size: 1},
					},
				},
			},
		},
	})
}
