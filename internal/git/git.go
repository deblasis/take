package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidURL  = errors.New("invalid git URL")
	ErrCloneFailed = errors.New("git clone failed")
)

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL       string
	TargetDir string
	Depth     int
}

// Clone clones a git repository
func Clone(opts CloneOptions) error {
	if !IsValidURL(opts.URL) {
		return ErrInvalidURL
	}

	args := []string{"clone"}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	// Add target directory if specified
	if opts.TargetDir != "" {
		args = append(args, opts.URL, opts.TargetDir)
	} else {
		args = append(args, opts.URL)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCloneFailed, string(output))
	}

	return nil
}

// IsValidURL checks if the given string is a valid git URL
func IsValidURL(url string) bool {
	// Check for SSH format (git@host:user/repo.git)
	if strings.HasPrefix(url, "git@") && strings.Contains(url, ":") {
		return strings.HasSuffix(url, ".git")
	}

	// Check for HTTPS format
	if strings.HasPrefix(url, "https://") {
		return strings.HasSuffix(url, ".git")
	}

	return false
}

// GetRepoName extracts the repository name from a git URL
func GetRepoName(url string) string {
	// Remove .git suffix if present
	name := strings.TrimSuffix(url, ".git")

	// Handle SSH URLs (git@github.com:user/repo)
	if strings.HasPrefix(name, "git@") {
		parts := strings.Split(name, ":")
		if len(parts) == 2 {
			name = parts[1]
		}
	}

	// Handle HTTPS URLs (https://github.com/user/repo)
	name = filepath.Base(name)

	return name
}

// IsGitInstalled checks if git is installed and available
func IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsGitRepo checks if the given directory is a git repository
func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
