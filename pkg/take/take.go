package take

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"archive/zip"

	"github.com/deblasis/take/internal/git"
)

var (
	ErrInvalidPath      = errors.New("invalid path specified")
	ErrPermissionDenied = errors.New("permission denied")
	ErrGitCloneFailed   = errors.New("git clone failed")
	ErrInvalidURL       = errors.New("invalid URL format")
	ErrDownloadFailed   = errors.New("failed to download file")
	ErrExtractionFailed = errors.New("failed to extract archive")
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
	// WasDownloaded indicates if a file was downloaded
	WasDownloaded bool
	// Error if any occurred
	Error error
}

// urlPatterns defines regex patterns for different URL types
var urlPatterns = struct {
	git     *regexp.Regexp
	tarball *regexp.Regexp
	zip     *regexp.Regexp
}{
	git:     regexp.MustCompile(`^([A-Za-z0-9]+@|https?|git|ssh|ftps?|rsync).*\.git/?$`),
	tarball: regexp.MustCompile(`^(https?|ftp).*\.(tar\.(gz|bz2|xz)|tgz)$`),
	zip:     regexp.MustCompile(`^(https?|ftp).*\.(zip)$`),
}

// Take executes the take command with the given options
func Take(opts Options) Result {
	// Validate input
	if opts.Path == "" {
		return Result{Error: ErrInvalidPath}
	}

	// Handle URLs and git repos
	if strings.Contains(opts.Path, "://") || strings.Contains(opts.Path, "@") || git.IsGitRepo(opts.Path) {
		switch {
		case git.IsGitRepo(opts.Path) || urlPatterns.git.MatchString(opts.Path):
			return handleGitURL(opts)
		case urlPatterns.tarball.MatchString(opts.Path):
			return handleTarballURL(opts)
		case urlPatterns.zip.MatchString(opts.Path):
			return handleZipURL(opts)
		default:
			return Result{Error: ErrInvalidURL}
		}
	}

	// Handle local directory
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

	return Result{
		FinalPath:  absPath,
		WasCreated: true,
	}
}

// handleDirectory creates a local directory
func handleDirectory(opts Options) Result {
	path := opts.Path
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return Result{Error: err}
		}
		path = filepath.Join(home, path[1:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Result{Error: err}
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		if os.IsPermission(err) {
			return Result{Error: ErrPermissionDenied}
		}
		return Result{Error: err}
	}

	return Result{
		FinalPath:  absPath,
		WasCreated: true,
	}
}

// handleGitURL handles git repository cloning
func handleGitURL(opts Options) Result {
	targetDir := git.GetRepoName(opts.Path)
	if targetDir == "" {
		targetDir = filepath.Base(opts.Path)
	}

	err := git.Clone(git.CloneOptions{
		URL:       opts.Path,
		TargetDir: targetDir,
		Depth:     opts.GitCloneDepth,
	})

	if err != nil {
		return Result{Error: fmt.Errorf("failed to clone repository: %w", err)}
	}

	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return Result{Error: err}
	}

	return Result{
		FinalPath:  absPath,
		WasCloned:  true,
	}
}

// handleTarballURL downloads and extracts a tarball
func handleTarballURL(opts Options) Result {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "take-*")
	if err != nil {
		return Result{Error: err}
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary file with proper extension
	ext := filepath.Ext(opts.Path)
	if ext == "" {
		ext = ".tar.gz" // default extension
	}
	tmpFile, err := os.CreateTemp(tmpDir, "archive-*"+ext)
	if err != nil {
		return Result{Error: err}
	}
	defer os.Remove(tmpFile.Name())

	// Download file
	resp, err := http.Get(opts.Path)
	if err != nil {
		return Result{Error: ErrDownloadFailed}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{Error: ErrDownloadFailed}
	}

	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return Result{Error: ErrDownloadFailed}
	}
	tmpFile.Close()

	// Extract archive
	cmd := exec.Command("tar", "xf", tmpFile.Name(), "-C", tmpDir)
	if err := cmd.Run(); err != nil {
		return Result{Error: fmt.Errorf("tar extraction failed: %v", err)}
	}

	// Find the extracted directory
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return Result{Error: err}
	}

	var extractedDir string
	for _, entry := range entries {
		if entry.IsDir() {
			extractedDir = entry.Name()
			break
		}
	}

	if extractedDir == "" {
		return Result{Error: fmt.Errorf("no directory found in archive")}
	}

	// Move the extracted directory to the current directory
	finalPath := filepath.Join(".", extractedDir)
	if err := os.Rename(filepath.Join(tmpDir, extractedDir), finalPath); err != nil {
		return Result{Error: err}
	}

	absPath, err := filepath.Abs(finalPath)
	if err != nil {
		return Result{Error: err}
	}

	return Result{
		FinalPath:     absPath,
		WasDownloaded: true,
	}
}

// handleZipURL downloads and extracts a zip file
func handleZipURL(opts Options) Result {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "take-*")
	if err != nil {
		return Result{Error: err}
	}
	defer os.RemoveAll(tmpDir)

	// Create temporary file
	tmpFile, err := os.CreateTemp(tmpDir, "archive-*.zip")
	if err != nil {
		return Result{Error: err}
	}
	defer os.Remove(tmpFile.Name())

	// Download file
	resp, err := http.Get(opts.Path)
	if err != nil {
		return Result{Error: ErrDownloadFailed}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{Error: ErrDownloadFailed}
	}

	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return Result{Error: ErrDownloadFailed}
	}
	tmpFile.Close()

	// Open the zip file for reading
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return Result{Error: fmt.Errorf("failed to open zip: %v", err)}
	}
	defer zipReader.Close()

	// Find the root directory in the zip
	var rootDir string
	for _, file := range zipReader.File {
		name := filepath.ToSlash(file.Name)
		parts := strings.Split(name, "/")

		// Skip files/directories starting with "." or "_"
		if strings.HasPrefix(parts[0], ".") || strings.HasPrefix(parts[0], "_") {
			continue
		}

		// Get the top-level directory
		if rootDir == "" {
			rootDir = parts[0]
		} else if parts[0] != rootDir {
			// If we find a different top-level directory, use the common parent
			if rootDir != filepath.Dir(name) {
				rootDir = ""
				break
			}
		}
	}

	if rootDir == "" {
		// If no root dir found, use the base name of the zip without extension
		rootDir = strings.TrimSuffix(filepath.Base(opts.Path), ".zip")
		// Create the root directory
		if err := os.MkdirAll(filepath.Join(tmpDir, rootDir), 0755); err != nil {
			return Result{Error: fmt.Errorf("failed to create root directory: %v", err)}
		}
	}

	// Extract files
	for _, file := range zipReader.File {
		// Calculate extraction path
		path := filepath.Join(tmpDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return Result{Error: fmt.Errorf("failed to create directory: %v", err)}
			}
			continue
		}

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return Result{Error: fmt.Errorf("failed to create parent directory: %v", err)}
		}

		// Create the file
		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return Result{Error: fmt.Errorf("failed to create file: %v", err)}
		}

		// Open the file in the zip
		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return Result{Error: fmt.Errorf("failed to open file in zip: %v", err)}
		}

		// Copy the contents
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return Result{Error: fmt.Errorf("failed to extract file: %v", err)}
		}
	}

	// Move the extracted directory to the current directory
	finalPath := filepath.Join(".", rootDir)
	// Remove target directory if it exists
	os.RemoveAll(finalPath)
	if err := os.Rename(filepath.Join(tmpDir, rootDir), finalPath); err != nil {
		return Result{Error: fmt.Errorf("failed to move directory: %v", err)}
	}

	absPath, err := filepath.Abs(finalPath)
	if err != nil {
		return Result{Error: err}
	}

	return Result{
		FinalPath:     absPath,
		WasDownloaded: true,
	}
}

// downloadFile downloads a file from a URL to a local file
func downloadFile(url string, file *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrDownloadFailed
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

// expandPath expands the given path, handling home directory (~) expansion
func expandPath(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidPath
	}

	// Expand home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	// Handle relative paths
	if !filepath.IsAbs(path) {
		path = filepath.Clean(path)
	}

	return path, nil
}
