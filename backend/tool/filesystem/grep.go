package filesystem

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"

	"github.com/furisto/construct/backend/tool/base"
)

type GrepInput struct {
	Query          string `json:"query"`
	Path           string `json:"path"`
	IncludePattern string `json:"include_pattern"`
	ExcludePattern string `json:"exclude_pattern"`
	CaseSensitive  bool   `json:"case_sensitive"`
	MaxResults     int    `json:"max_results"`
}

type GrepMatch struct {
	FilePath    string        `json:"file_path"`
	LineNumber  int           `json:"line_number"`
	LineContent string        `json:"line_content"`
	Context     []ContextLine `json:"context"`
}

type ContextLine struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

type GrepResult struct {
	Matches       []GrepMatch `json:"matches"`
	TotalMatches  int         `json:"total_matches"`
	SearchedFiles int         `json:"searched_files"`
}

func Grep(input *GrepInput) (*GrepResult, error) {
	if input.Query == "" || input.Path == "" {
		return nil, base.NewCustomError("query and path are required", []string{
			"Provide both a search query and a path to search in",
		})
	}

	if input.MaxResults == 0 {
		input.MaxResults = 50
	}

	if isRipgrepAvailable() {
		return performRipgrep(input)
	}

	return performRegularGrep(input)
}

func isRipgrepAvailable() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

func performRipgrep(input *GrepInput) (*GrepResult, error) {
	args := []string{
		"--json",
		"--line-number",
		"--with-filename",
		"--context", "2",
	}

	if !input.CaseSensitive {
		args = append(args, "--ignore-case")
	}

	if input.IncludePattern != "" {
		args = append(args, "--glob", input.IncludePattern)
	}

	if input.ExcludePattern != "" {
		args = append(args, "--glob", "!"+input.ExcludePattern)
	}

	args = append(args, input.Query, input.Path)

	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No matches found, return empty result
			return &GrepResult{
				Matches:       []GrepMatch{},
				TotalMatches:  0,
				SearchedFiles: 0,
			}, nil
		}
		return nil, base.NewCustomError("ripgrep error", []string{
			"Check that your regex pattern is valid",
			"Verify that the search path exists and is accessible",
		}, "error", err)
	}

	return parseRipgrepOutput(string(output), input.MaxResults)
}

func performRegularGrep(input *GrepInput) (*GrepResult, error) {
	args := []string{
		"-r",
		"-n",
		"-H",
		"-C", "2",
	}

	if !input.CaseSensitive {
		args = append(args, "-i")
	}

	if input.IncludePattern != "" {
		args = append(args, "--include="+input.IncludePattern)
	}

	if input.ExcludePattern != "" {
		args = append(args, "--exclude="+input.ExcludePattern)
	}

	args = append(args, input.Query, input.Path)

	cmd := exec.Command("grep", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No matches found, return empty result
			return &GrepResult{
				Matches:       []GrepMatch{},
				TotalMatches:  0,
				SearchedFiles: 0,
			}, nil
		}
		return nil, base.NewCustomError("grep error", []string{
			"Check that your regex pattern is valid",
			"Verify that the search path exists and is accessible",
		}, "error", err)
	}

	return parseGrepOutput(string(output), input.MaxResults)
}

func parseRipgrepOutput(output string, maxResults int) (*GrepResult, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	matches := []GrepMatch{}
	searchedFiles := make(map[string]bool)

	type rgMatch struct {
		Type string `json:"type"`
		Data struct {
			Path struct {
				Text string `json:"text"`
			} `json:"path"`
			LineNumber int `json:"line_number"`
			Lines      struct {
				Text string `json:"text"`
			} `json:"lines"`
		} `json:"data"`
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		var match rgMatch
		if err := json.Unmarshal([]byte(line), &match); err != nil {
			continue
		}

		if match.Type == "match" {
			if len(matches) >= maxResults {
				break
			}

			filePath := match.Data.Path.Text
			searchedFiles[filePath] = true

			grepMatch := GrepMatch{
				FilePath:    filePath,
				LineNumber:  match.Data.LineNumber,
				LineContent: match.Data.Lines.Text,
				Context:     []ContextLine{},
			}

			matches = append(matches, grepMatch)
		}
	}

	return &GrepResult{
		Matches:       matches,
		TotalMatches:  len(matches),
		SearchedFiles: len(searchedFiles),
	}, nil
}

func parseGrepOutput(output string, maxResults int) (*GrepResult, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	matches := []GrepMatch{}
	searchedFiles := make(map[string]bool)

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if len(matches) >= maxResults {
			break
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		lineNumStr := parts[1]
		content := parts[2]

		lineNum, err := strconv.Atoi(lineNumStr)
		if err != nil {
			continue
		}

		searchedFiles[filePath] = true

		grepMatch := GrepMatch{
			FilePath:    filePath,
			LineNumber:  lineNum,
			LineContent: content,
			Context:     []ContextLine{},
		}

		matches = append(matches, grepMatch)
	}

	return &GrepResult{
		Matches:       matches,
		TotalMatches:  len(matches),
		SearchedFiles: len(searchedFiles),
	}, nil
}
