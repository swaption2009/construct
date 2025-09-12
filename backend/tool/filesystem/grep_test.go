package filesystem

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestParseRipgrepOutput(t *testing.T) {
	tests := []struct {
		name           string
		maxResults     int
		context        int
		expectedResult *GrepResult
		// Truncation tests
	}{
		{
			name:       "no_truncation_under_limit",
			maxResults: 5,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file1.go",
						Value:    ":5:match 1",
					},
					{
						FilePath: "/path/file2.go",
						Value:    ":10:match 2",
					},
				},
				TotalMatches:     2,
				TruncatedMatches: 0,
				SearchedFiles:    2,
			},
		},
		{
			name:       "exact_limit_no_truncation",
			maxResults: 3,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file1.go",
						Value:    ":5:match 1",
					},
					{
						FilePath: "/path/file2.go",
						Value:    ":10:match 2",
					},
					{
						FilePath: "/path/file3.go",
						Value:    ":15:match 3",
					},
				},
				TotalMatches:     3,
				TruncatedMatches: 0,
				SearchedFiles:    3,
			},
		},
		{
			name:       "over_limit_with_truncation",
			maxResults: 3,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file1.go",
						Value:    ":5:match 1",
					},
					{
						FilePath: "/path/file2.go",
						Value:    ":10:match 2",
					},
					{
						FilePath: "/path/file3.go",
						Value:    ":15:match 3",
					},
				},
				TotalMatches:     5,
				TruncatedMatches: 2,
				SearchedFiles:    5,
			},
		},
		// File grouping tests
		{
			name:       "missing_begin_marker",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches:          []GrepMatch{},
				TotalMatches:     0,
				TruncatedMatches: 0,
				SearchedFiles:    0,
			},
		},
		{
			name:       "missing_end_marker",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches:          []GrepMatch{},
				TotalMatches:     0,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
		{
			name:       "empty_file_groups",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file2.go",
						Value:    ":5:actual match",
					},
				},
				TotalMatches:     1,
				TruncatedMatches: 0,
				SearchedFiles:    3,
			},
		},
		// Basic functionality tests
		{
			name:       "empty_input",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches:          []GrepMatch{},
				TotalMatches:     0,
				TruncatedMatches: 0,
				SearchedFiles:    0,
			},
		},
		{
			name:       "single_file_single_match_with_context",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file.go",
						Value:    "-1-package main:2:func main() {-3-    fmt.Println(\"hello\")",
					},
				},
				TotalMatches:     1,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
		{
			name:       "single_file_multiple_matches",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file.go",
						Value:    "-1-package main:2:func main() {-3-    fmt.Println(\"hello\")-4--8-    // another section:9:    fmt.Printf(\"world\")-10-}",
					},
				},
				TotalMatches:     1,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
		{
			name:       "multiple_files_with_matches",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file1.go",
						Value:    ":5:match in file1",
					},
					{
						FilePath: "/path/file2.go",
						Value:    ":10:match in file2",
					},
				},
				TotalMatches:     2,
				TruncatedMatches: 0,
				SearchedFiles:    2,
			},
		},
		{
			name:       "no_matches_only_context",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches:          []GrepMatch{},
				TotalMatches:     0,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
		// JSON parsing tests
		{
			name:       "invalid_json_lines",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file.go",
						Value:    ":2:valid match",
					},
				},
				TotalMatches:     1,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
		{
			name:       "unknown_entry_types",
			maxResults: 50,
			context:    2,
			expectedResult: &GrepResult{
				Matches: []GrepMatch{
					{
						FilePath: "/path/file.go",
						Value:    ":2:valid match-3-valid context",
					},
				},
				TotalMatches:     1,
				TruncatedMatches: 0,
				SearchedFiles:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load snapshot file using test name
			fs := afero.NewOsFs()
			snapshotPath := filepath.Join("snapshots", "grep", tt.name+".json")
			inputData, err := afero.ReadFile(fs, snapshotPath)
			if err != nil {
				t.Fatalf("Failed to read snapshot file %s: %v", snapshotPath, err)
			}

			result, err := parseRipgrepOutput(string(inputData), tt.maxResults, tt.context)
			if err != nil {
				t.Fatalf("parseRipgrepOutput() error = %v", err)
			}

			if diff := cmp.Diff(tt.expectedResult, result); diff != "" {
				t.Errorf("parseRipgrepOutput() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
