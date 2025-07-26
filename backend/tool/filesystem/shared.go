package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/furisto/construct/backend/tool/base"
)

// ValidateAbsolutePath checks if the given path is absolute
func ValidateAbsolutePath(path string) error {
	if path == "" {
		return base.ValidationError{Field: "path", Message: "path is required"}
	}
	if !filepath.IsAbs(path) {
		return base.ErrPathNotAbsolute
	}
	return nil
}

// EnsureParentDirs creates parent directories if they don't exist
func EnsureParentDirs(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

// CheckPathExists checks if a path exists and returns info about it
func CheckPathExists(path string) (exists bool, isDir bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		if os.IsPermission(err) {
			return false, false, base.ErrPermissionDenied
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

// WalkDirectoryTree walks a directory tree and returns a string representation
func WalkDirectoryTree(rootPath string) (string, error) {
	result := rootPath + "\n"
	err := walkDirRecursive(rootPath, &result, 1, 3, "  ")
	if err != nil {
		return "", err
	}
	return result, nil
}

// walkDirRecursive is a helper function for WalkDirectoryTree
func walkDirRecursive(path string, result *string, currentLevel, maxLevel int, indent string) error {
	if currentLevel > maxLevel {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			*result += indent + "|_ " + entry.Name() + "\n"
			err := walkDirRecursive(filepath.Join(path, entry.Name()), result, currentLevel+1, maxLevel, indent+"        ")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// SanitizePath cleans and normalizes a file path
func SanitizePath(path string) string {
	// Clean the path to remove any . or .. elements
	cleaned := filepath.Clean(path)

	// Convert to forward slashes for consistency
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")

	return cleaned
}
