package take

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

// Take creates a directory and/or clones a repository and changes to it
func Take(opts Options) Result {
	if opts.Path == "" {
		return Result{Error: ErrInvalidPath}
	}

	// Handle URLs
	if strings.Contains(opts.Path, "://") || strings.Contains(opts.Path, "@") {
		switch {
		case urlPatterns.git.MatchString(opts.Path):
			return handleGitURL(opts)
		case urlPatterns.tarball.MatchString(opts.Path):
			return handleTarballURL(opts)
		case urlPatterns.zip.MatchString(opts.Path):
			return handleZipURL(opts)
		default:
			return Result{Error: ErrInvalidURL}
		}
	}

	// Handle local directory creation
	return handleDirectory(opts)
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

// handleGitURL clones a git repository
func handleGitURL(opts Options) Result {
	// Extract repository name from URL
	repoName := filepath.Base(opts.Path)
	repoName = strings.TrimSuffix(repoName, ".git")
	targetDir := filepath.Join(".", repoName)

	// Build git clone command
	args := []string{"clone"}
	if opts.GitCloneDepth > 0 {
		args = append(args, "--depth", string(opts.GitCloneDepth))
	}
	args = append(args, opts.Path, targetDir)

	// Execute git clone
	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return Result{Error: ErrGitCloneFailed}
	}

	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return Result{Error: err}
	}

	return Result{
		FinalPath: absPath,
		WasCloned: true,
	}
}

// handleTarballURL downloads and extracts a tarball
func handleTarballURL(opts Options) Result {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "take-*.tar.*")
	if err != nil {
		return Result{Error: err}
	}
	defer os.Remove(tmpFile.Name())

	// Download file
	if err := downloadFile(opts.Path, tmpFile); err != nil {
		return Result{Error: ErrDownloadFailed}
	}

	// Extract archive
	cmd := exec.Command("tar", "xf", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return Result{Error: ErrExtractionFailed}
	}

	// Get the extracted directory name
	cmd = exec.Command("tar", "tf", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return Result{Error: ErrExtractionFailed}
	}

	// Get the root directory name
	firstLine := strings.SplitN(string(output), "\n", 2)[0]
	dirName := strings.Split(firstLine, "/")[0]

	absPath, err := filepath.Abs(dirName)
	if err != nil {
		return Result{Error: err}
	}

	return Result{
		FinalPath:     absPath,
		WasDownloaded: true,
	}
}

// handleZipURL downloads and extracts a ZIP file
func handleZipURL(opts Options) Result {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "take-*.zip")
	if err != nil {
		return Result{Error: err}
	}
	defer os.Remove(tmpFile.Name())

	// Download file
	if err := downloadFile(opts.Path, tmpFile); err != nil {
		return Result{Error: ErrDownloadFailed}
	}

	// Extract archive
	cmd := exec.Command("unzip", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return Result{Error: ErrExtractionFailed}
	}

	// Get the extracted directory name
	cmd = exec.Command("unzip", "-l", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return Result{Error: ErrExtractionFailed}
	}

	// Parse unzip output to get root directory
	lines := strings.Split(string(output), "\n")
	if len(lines) < 4 {
		return Result{Error: ErrExtractionFailed}
	}
	dirName := strings.Split(lines[3], "   ")[len(strings.Split(lines[3], "   "))-1]
	dirName = strings.Split(dirName, "/")[0]

	absPath, err := filepath.Abs(dirName)
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
