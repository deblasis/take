package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidURL(t *testing.T) {
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
			name: "valid URL with target dir",
			opts: CloneOptions{
				URL:       "https://github.com/octocat/Hello-World.git",
				TargetDir: filepath.Join(tmpDir, "hello-world"),
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
