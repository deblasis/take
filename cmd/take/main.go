package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deblasis/take/internal/git"
	"github.com/deblasis/take/pkg/take"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse flags
	depth := flag.Int("depth", 0, "Git clone depth (0 for full clone)")
	force := flag.Bool("force", false, "Force operation even if directory exists")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("take version %s (%s) built on %s\n", version, commit, date)
		os.Exit(0)
	}

	// Get target path from arguments
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: take [-depth N] [-force] <directory|git-url>")
		os.Exit(1)
	}

	target := flag.Arg(0)

	// Create options
	opts := take.Options{
		Path:          target,
		GitCloneDepth: *depth,
		Force:         *force,
	}

	// Execute take command
	result := take.Take(opts)
	if result.Error != nil {
		fmt.Fprintln(os.Stderr, result.Error)
		os.Exit(1)
	}

	// Print the final path to stdout
	// This will be used by the shell function to cd into the directory
	fmt.Println(result.FinalPath)
}

// handleGitURL handles git repository cloning
func handleGitURL(url string, depth int) (string, error) {
	if !git.IsGitInstalled() {
		return "", fmt.Errorf("git is not installed")
	}

	// Get repository name for the target directory
	repoName := git.GetRepoName(url)
	targetDir := filepath.Join(".", repoName)

	// Clone the repository
	err := git.Clone(git.CloneOptions{
		URL:       url,
		TargetDir: targetDir,
		Depth:     depth,
	})
	if err != nil {
		return "", err
	}

	return targetDir, nil
}
