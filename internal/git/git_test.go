package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsValidURL(t *testing.T) {
	// Create a temporary git repo for testing
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid HTTPS URL",
			url:  "https://github.com/user/repo.git",
			want: true,
		},
		{
			name: "valid SSH URL",
			url:  "git@github.com:user/repo.git",
			want: true,
		},
		{
			name: "valid local repo",
			url:  tmpDir,
			want: true,
		},
		{
			name: "invalid URL - no .git suffix",
			url:  "https://github.com/user/repo",
			want: false,
		},
		{
			name: "invalid URL - not git URL",
			url:  "https://example.com/file.txt",
			want: false,
		},
		{
			name: "invalid URL - empty",
			url:  "",
			want: false,
		},
		{
			name: "invalid URL - malformed SSH",
			url:  "git@github.com/user/repo.git",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidURL(tt.url); got != tt.want {
				t.Errorf("IsValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "HTTPS URL",
			url:  "https://github.com/user/repo.git",
			want: "repo",
		},
		{
			name: "SSH URL",
			url:  "git@github.com:user/repo.git",
			want: "repo",
		},
		{
			name: "URL without .git",
			url:  "https://github.com/user/repo",
			want: "repo",
		},
		{
			name: "URL with dashes",
			url:  "git@github.com:user/my-awesome-repo.git",
			want: "my-awesome-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRepoName(tt.url); got != tt.want {
				t.Errorf("GetRepoName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "valid git repo",
			dir:  tmpDir,
			want: true,
		},
		{
			name: "not a git repo",
			dir:  os.TempDir(),
			want: false,
		},
		{
			name: "non-existent directory",
			dir:  "/path/that/does/not/exist",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGitRepo(tt.dir); got != tt.want {
				t.Errorf("IsGitRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClone(t *testing.T) {
	if !IsGitInstalled() {
		t.Skip("Git is not installed, skipping clone tests")
	}

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-clone-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test repo
	testRepoDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}
	defer os.RemoveAll(testRepoDir)

	// Initialize test repo
	if err := exec.Command("git", "init", testRepoDir).Run(); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	// Create a test file and commit it
	testFile := filepath.Join(testRepoDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git username: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testRepoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	tests := []struct {
		name    string
		opts    CloneOptions
		wantErr bool
	}{
		{
			name: "invalid URL",
			opts: CloneOptions{
				URL:       "not-a-git-url",
				TargetDir: filepath.Join(tmpDir, "invalid"),
			},
			wantErr: true,
		},
		{
			name: "valid local repo",
			opts: CloneOptions{
				URL:       testRepoDir,
				TargetDir: filepath.Join(tmpDir, "valid"),
				Depth:     1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Clone(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !IsGitRepo(tt.opts.TargetDir) {
					t.Error("Clone() did not create a valid git repository")
				}
			}
		})
	}
}
