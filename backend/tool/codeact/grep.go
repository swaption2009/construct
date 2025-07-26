package codeact

import (
	"fmt"

	"github.com/furisto/construct/backend/tool/filesystem"
	"github.com/grafana/sobek"
)

var grepDescription = `
## Description
The grep tool performs fast text-based regex searches to find exact pattern matches within files or directories. It leverages efficient searching algorithms to quickly scan through your codebase and locate specific patterns.

## Parameters
- **query** (*string*, required): The regex pattern to search for. Must be a valid regex pattern; special characters must be escaped appropriately.
- **path** (*string*, required): Absolute path to the directory or file to search within. Forward slashes (/) work on all platforms.
- **include_pattern** (*string*, optional): Glob pattern for files to include in the search (e.g., "*.js" for JavaScript files only). Allows focusing your search on specific file types.
- **exclude_pattern** (*string*, optional): Glob pattern for files to exclude from the search. Useful for ignoring build artifacts, dependencies, or other irrelevant files.
- **case_sensitive** (*boolean*, optional): Whether the search should be case sensitive. Defaults to false.
- **max_results** (*number*, optional): Maximum number of results to return. Defaults to 50 to prevent overwhelming output.

## Expected Output
Returns an object containing the search results:
%[1]s
{
  "matches": [
    {
      "file_path": "/path/to/file.js",
      "line_number": 42,
      "line_content": "const searchPattern = /regex/;",
      "context": [
        { "line_number": 40, "content": "// Previous lines for context" },
        { "line_number": 41, "content": "// More context" },
        { "line_number": 42, "content": "const searchPattern = /regex/;" },
        { "line_number": 43, "content": "// Context after match" }
      ]
    },
    // Additional matches...
  ],
  "total_matches": 3,
  "searched_files": 125
}
%[1]s

## CRITICAL REQUIREMENTS
- **Precise Pattern Specification**: Your regex pattern must be properly escaped for accurate matching.
  %[1]s
  // To search for "user.login()", escape special characters:
  grep({
    query: "user\\.login\\(\\)",
    path: "/workspace/src"
  })
  %[1]s
- **Search Path Verification**: Always use absolute paths starting with "/" for consistent results.
- **Scope Management**: Use include/exclude patterns to control search scope and improve performance:
  %[1]s
  // Only search JavaScript files, exclude tests
  grep({
    query: "function init",
    path: "/workspace/project",
    include_pattern: "*.js",
    exclude_pattern: "**/__tests__/**"
  })
  %[1]s
- **Performance Considerations**:
  - Narrow your search scope with specific paths and patterns for faster results
  - Be specific with your regex to avoid excessive matches
  - Use reasonable max_results limits for large codebases
- **Complex Pattern Handling**: For complex patterns, test iteratively:
  %[1]s
  // First search broadly
  grep({
    query: "api\\.connect",
    path: "/workspace/src"
  })
  
  // Then refine with more specific pattern
  grep({
    query: "api\\.connect\\(['\"]production['\"]\\)",
    path: "/workspace/src/services"
  })
  %[1]s

## When to use
- **Finding Symbol Definitions**: When you need to locate specific function, class, or variable definitions.
- **Code Pattern Discovery**: When identifying patterns across multiple files (error handling, API calls, etc.).
- **API Usage Exploration**: When discovering how specific APIs or functions are used throughout the codebase.
- **Error Text Location**: When tracking down where specific error messages are defined or thrown.
- **Dependency Identification**: When finding all imports or requires of a specific module.
- **Configuration Search**: When locating specific configuration patterns across multiple files.

## Usage Examples

### Finding Function Definitions
%[1]s
grep({
  query: "function\\s+getUserData\\s*\\(",
  path: "/workspace/src",
  include_pattern: "*.js",
  exclude_pattern: "**/node_modules/**"
})
%[1]s
`

func NewGrepTool() Tool {
	return NewOnDemandTool(
		"grep",
		fmt.Sprintf(grepDescription, "```"),
		grepInput,
		grepHandler,
	)
}

func grepInput(session *Session, args []sobek.Value) (any, error) {
	if len(args) < 1 {
		return nil, nil
	}

	inputObj := args[0].ToObject(session.VM)
	if inputObj == nil {
		return nil, nil
	}

	input := &filesystem.GrepInput{}
	if query := inputObj.Get("query"); query != nil {
		input.Query = query.String()
	}
	if path := inputObj.Get("path"); path != nil {
		input.Path = path.String()
	}
	if includePattern := inputObj.Get("include_pattern"); includePattern != nil {
		input.IncludePattern = includePattern.String()
	}
	if excludePattern := inputObj.Get("exclude_pattern"); excludePattern != nil {
		input.ExcludePattern = excludePattern.String()
	}
	if caseSensitive := inputObj.Get("case_sensitive"); caseSensitive != nil {
		input.CaseSensitive = caseSensitive.ToBoolean()
	}
	if maxResults := inputObj.Get("max_results"); maxResults != nil {
		input.MaxResults = int(maxResults.ToInteger())
	}

	if input.MaxResults == 0 {
		input.MaxResults = 50
	}

	return input, nil
}

func grepHandler(session *Session) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		rawInput, err := grepInput(session, call.Arguments)
		if err != nil {
			session.Throw(err)
		}
		input := rawInput.(*filesystem.GrepInput)

		result, err := filesystem.Grep(input)
		if err != nil {
			session.Throw(err)
		}

		SetValue(session, "result", result)
		return session.VM.ToValue(result)
	}
}
