package take

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath      = errors.New("invalid path specified")
	ErrPermissionDenied = errors.New("permission denied")
	ErrGitCloneFailed   = errors.New("git clone failed")
	ErrInvalidURL       = errors.New("invalid URL format")
)

// Options represents configuration options for the take command
type Options struct {
	// Path is the target directory or URL
	Path string
	// GitCloneDepth for shallow clones, 0 means full clone
	GitCloneDepth int
	// Force will overwrite existing directory
	Force bool
}

// Result represents the outcome of a take operation
type Result struct {
	// FinalPath is the absolute path of the created/target directory
	FinalPath string
	// WasCreated indicates if a new directory was created
	WasCreated bool
	// WasCloned indicates if a git repository was cloned
	WasCloned bool
	// Error if any occurred
	Error error
}

// Take creates a directory and/or clones a repository and changes to it
func Take(opts Options) Result {
	result := Result{}

	// Handle git URLs
	if isGitURL(opts.Path) {
		// TODO: Implement git clone logic
		return Result{Error: errors.New("git clone not implemented")}
	}

	// Expand path
	expandedPath, err := expandPath(opts.Path)
	if err != nil {
		return Result{Error: err}
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(expandedPath, 0755); err != nil {
		if os.IsPermission(err) {
			return Result{Error: ErrPermissionDenied}
		}
		return Result{Error: err}
	}

	// Get absolute path
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return Result{Error: err}
	}

	result.FinalPath = absPath
	result.WasCreated = true

	return result
}

// isGitURL checks if the given string is a git URL
func isGitURL(path string) bool {
	if strings.HasSuffix(path, ".git") {
		return true
	}

	u, err := url.Parse(path)
	if err != nil {
		return false
	}

	// Check for SSH git URL format (git@github.com:user/repo.git)
	if strings.HasPrefix(path, "git@") {
		return true
	}

	// Check for HTTPS git URL
	return u.Scheme == "https" && strings.Contains(u.Host, "github.com")
}

// expandPath expands the ~ to the user's home directory
func expandPath(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidPath
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	return filepath.Clean(path), nil
}
