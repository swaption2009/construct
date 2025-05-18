package tool

import (
	"fmt"
	"os"

	"github.com/grafana/sobek"
	"github.com/spf13/afero"

	"github.com/furisto/construct/backend/tool/codeact"
)

const readFileDescription = `
## Description
Reads and returns the complete contents of a file at the specified absolute path. This tool is essential for examining existing files when you need to understand, analyze, or extract information from them. The file content is returned as a string, making it suitable for text files such as code, configuration files, documentation, and structured data.

## Parameters
- **path** (string, required): Absolute path to the file you want to read (e.g., "/workspace/project/src/app.js"). Forward slashes (/) work on all platforms.

## Expected Output
Returns an object containing the file content as a string:
%[1]s
{
  "path": "The absolute path of the file",
  "content": "The complete content of the file as a string"
}
%[1]s

If the file doesn't exist or cannot be read, it will throw an exception describing the issue.

## IMPORTANT USAGE NOTES
- **Check file extensions**: Ensure you're reading appropriate file types; this tool is best suited for text files
- **Process binary files carefully**: Binary files may return unreadable content; consider specialized tools for these cases
"- **Path format**: Always use absolute paths starting with "/". For example: /workspace/project/package.json"

## When to use
- **Code analysis**: When you need to understand existing code structure, imports, or implementations
- **Configuration inspection**: To examine settings in config files like JSON, YAML, or .env files
- **Content extraction**: To retrieve data from text files for processing or analysis
- **Before modifications**: Read a file first to understand its structure before making changes
- **Documentation review**: To analyze README files, specifications, or documentation
- **Data gathering**: When collecting information stored in logs, CSVs, or other structured data files

## Usage Examples

### Analyzing source code
%[1]s
try {
  const sourceCode = read_file("/workspace/project/src/components/Button.jsx");
  // Count React hooks in component
  const hooksCount = sourceCode.content.match(/use[A-Z]\w+\(/g) || [];
  print(%[2]sThis component uses ${hooksCount.length} React hooks%[2]s);
} catch (error) {
  print("Error reading file:", error);
}
%[1]s

### Reading and processing structured data
%[1]s
try {
  const csvData = read_file("/workspace/project/data/users.csv");
  const rows = csvData.content.split('\n').map(row => row.split(','));
  const headers = rows.shift();
  print(%[2]sFound ${rows.length} user records with fields: ${headers.join(', ')}%[2]s);
} catch (error) {
  print("Error reading file:", error);
}
%[1]s
`

type ReadFileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func NewReadFileTool() codeact.Tool {
	return codeact.NewOnDemandTool(
		"read_file",
		fmt.Sprintf(readFileDescription, "```", "`"),
		readFileHandler,
	)
}

func readFileHandler(session *codeact.Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		path := call.Argument(0).String()

		result, err := readFile(session.FS, path)
		if err != nil {
			session.Throw(err)
		}

		return session.VM.ToValue(result)
	}
}

func readFile(fsys afero.Fs, path string) (*ReadFileResult, error) {
	if _, err := fsys.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, codeact.NewError(codeact.FileNotFound, "path", path)
		}
		if os.IsPermission(err) {
			return nil, codeact.NewError(codeact.PermissionDenied, "path", path)
		}
		return nil, codeact.NewError(codeact.CannotStatFile, "path", path)
	}

	content, err := afero.ReadFile(fsys, path)
	if err != nil {
		return nil, codeact.NewCustomError("error reading file", []string{
			"Verify that you have the permission to read the file",
		}, "path", path, "error", err)
	}

	return &ReadFileResult{
		Path:    path,
		Content: string(content),
	}, nil

}
