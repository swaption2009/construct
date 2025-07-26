package codeact

import (
	"fmt"

	"github.com/furisto/construct/backend/tool/base"
	"github.com/furisto/construct/backend/tool/filesystem"
	"github.com/grafana/sobek"
)


const findFileDescription = `
## Description
Finds files matching a glob pattern using ripgrep for optimal performance when available, falling back to filesystem walking with doublestar. This tool is designed for discovering files by name patterns rather than content, making it ideal for locating specific files, exploring project structure, or finding files of certain types.

## Parameters
- **pattern** (string, required): Glob pattern to match files against (e.g., "*.js", "**/*.go", "test*.py"). Supports standard glob patterns including wildcards (* and ?) and recursive patterns (**).
- **path** (string, required): Absolute path to the directory to search within. Forward slashes (/) work on all platforms.
- **exclude_pattern** (string, optional): Glob pattern for files to exclude from results. Useful for ignoring build artifacts, dependencies, or other irrelevant files.
- **max_results** (number, optional): Maximum number of results to return. Defaults to 50 to prevent overwhelming output.

## Expected Output
Returns an object containing the matching file paths:
%[1]s
{
  "files": [
    "/path/to/matching/file1.js",
    "/path/to/matching/file2.js",
    "/path/to/nested/dir/file3.js"
  ],
  "total_files": 3,
  "truncated_count": 0
}
%[1]s

**Details**
- **files**: Array of absolute file paths that matched the glob pattern
- **total_files**: Total number of files that matched the pattern and are included in the results
- **truncated_count**: Number of additional matching files that were found but excluded from results due to max_results limit. 0 indicates no truncation occurred.

## IMPORTANT USAGE NOTES
- **Pattern Specificity**: Be as specific as possible with your patterns to get relevant results
  %[1]s
  // Good: Find all React components
  find_file({
    pattern: "**/*Component.jsx",
    path: "/workspace/src"
  })
  
  // Better: Find components in specific directory
  find_file({
    pattern: "*Component.jsx", 
    path: "/workspace/src/components"
  })
  %[1]s
- **Performance Considerations**: Use specific paths and patterns for faster results
- **Exclude Patterns**: Use exclude patterns to filter out unwanted files
  %[1]s
  find_file({
    pattern: "*.js",
    path: "/workspace/project",
    exclude_pattern: "**/node_modules/**"
  })
  %[1]s
- **Path Format**: Always use absolute paths starting with "/"

## When to use
- **File Discovery**: When you need to find files by name patterns across a project
- **Project Exploration**: When exploring unfamiliar codebases to understand structure
- **Type-specific Searches**: When looking for all files of a certain type (e.g., all .json config files)
- **Template/Component Finding**: When locating specific templates, components, or modules
- **Build Artifact Location**: When finding generated files or build outputs

## Usage Examples

### Find all JavaScript files
%[1]s
find_file({
  pattern: "**/*.js",
  path: "/workspace/project/src",
  exclude_pattern: "**/__tests__/**"
})
%[1]s

### Find configuration files
%[1]s
find_file({
  pattern: "*.{json,yaml,yml}",
  path: "/workspace/project",
  max_results: 50
})
%[1]s

### Find test files
%[1]s
find_file({
  pattern: "**/*test.go",
  path: "/workspace/go-project"
})
%[1]s
`

func NewFindFileTool() Tool {
	return NewOnDemandTool(
		base.ToolNameFindFile,
		fmt.Sprintf(findFileDescription, "```"),
		findFileInput,
		findFileHandler,
	)
}

func findFileInput(session *Session, args []sobek.Value) (any, error) {
	if len(args) < 1 {
		return nil, nil
	}

	inputObj := args[0].ToObject(session.VM)
	if inputObj == nil {
		return nil, nil
	}

	input := &filesystem.FindFileInput{}
	if pattern := inputObj.Get("pattern"); pattern != nil {
		input.Pattern = pattern.String()
	}
	if path := inputObj.Get("path"); path != nil {
		input.Path = path.String()
	}
	if excludePattern := inputObj.Get("exclude_pattern"); excludePattern != nil {
		input.ExcludePattern = excludePattern.String()
	}
	if maxResults := inputObj.Get("max_results"); maxResults != nil {
		input.MaxResults = int(maxResults.ToInteger())
	}

	if input.MaxResults == 0 {
		input.MaxResults = 50
	}

	return input, nil
}

func findFileHandler(session *Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		rawInput, err := findFileInput(session, call.Arguments)
		if err != nil {
			session.Throw(err)
		}
		input := rawInput.(*filesystem.FindFileInput)

		result, err := filesystem.FindFile(session.FS, input)
		if err != nil {
			session.Throw(err)
		}

		SetValue(session, "result", result)
		return session.VM.ToValue(result)
	}
}