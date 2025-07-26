package codeact

import (
	"fmt"

	"github.com/grafana/sobek"

	"github.com/furisto/construct/backend/tool/filesystem"
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
  "content": "1: The complete content of the file as a string"
}
%[1]s

If the file doesn't exist or cannot be read, it will throw an exception describing the issue. The content will be returned with line numbers prefixed to each line. 
The line numbers are not part of the actual file content, they are just for you to understand the file structure.

## IMPORTANT USAGE NOTES
- **Check file extensions**: Ensure you're reading appropriate file types; this tool is best suited for text files
- **Process binary files carefully**: Binary files may return unreadable content; consider specialized tools for these cases
- **Path format**: Always use absolute paths starting with "/". For example: /workspace/project/package.json"

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

func NewReadFileTool() Tool {
	return NewOnDemandTool(
		"read_file",
		fmt.Sprintf(readFileDescription, "```", "`"),
		readFileInput,
		readFileHandler,
	)
}

func readFileInput(session *Session, args []sobek.Value) (any, error) {
	if len(args) < 1 {
		return nil, nil
	}

	return &filesystem.ReadFileInput{
		Path: args[0].String(),
	}, nil
}

func readFileHandler(session *Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		rawInput, err := readFileInput(session, call.Arguments)
		if err != nil {
			session.Throw(err)
		}
		input := rawInput.(*filesystem.ReadFileInput)

		result, err := filesystem.ReadFile(session.FS, input)
		if err != nil {
			session.Throw(err)
		}

		SetValue(session, "result", result)
		return session.VM.ToValue(result)
	}
}
