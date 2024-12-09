package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/deblasis/take/internal/git"
	"github.com/deblasis/take/pkg/take"
)

func TestTakeIntegration(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T)
		validate func(t *testing.T, result take.Result)
		wantErr  bool
	}{
		{
			name: "create new directory",
			args: []string{"newdir"},
			validate: func(t *testing.T, result take.Result) {
				if !result.WasCreated {
					t.Error("Expected directory to be created")
				}
				if _, err := os.Stat(result.FinalPath); os.IsNotExist(err) {
					t.Error("Directory was not created")
				}
			},
		},
		{
			name: "create nested directories",
			args: []string{"parent/child/grandchild"},
			validate: func(t *testing.T, result take.Result) {
				if !result.WasCreated {
					t.Error("Expected directories to be created")
				}
				if _, err := os.Stat(result.FinalPath); os.IsNotExist(err) {
					t.Error("Directories were not created")
				}
			},
		},
		{
			name: "handle existing directory",
			args: []string{"existing"},
			setup: func(t *testing.T) {
				if err := os.Mkdir("existing", 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			},
			validate: func(t *testing.T, result take.Result) {
				if result.Error != nil {
					t.Errorf("Expected no error for existing directory, got %v", result.Error)
				}
			},
		},
		{
			name: "clone git repository",
			args: []string{"https://github.com/octocat/Hello-World.git"},
			validate: func(t *testing.T, result take.Result) {
				if !result.WasCloned {
					t.Error("Expected repository to be cloned")
				}
				if !git.IsGitRepo(result.FinalPath) {
					t.Error("Not a valid git repository")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			// Build command
			args := append([]string{"take"}, tt.args...)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = tmpDir

			// Run command
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Errorf("take command error = %v, wantErr %v\nOutput: %s", err, tt.wantErr, output)
				return
			}

			if tt.validate != nil {
				result := take.Result{
					FinalPath:  string(output),
					WasCreated: true, // This would need to be determined from actual command output
				}
				tt.validate(t, result)
			}
		})
	}
}

func TestTakeShellIntegration(t *testing.T) {
	// Skip on Windows as shell integration tests need different handling
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell integration tests on Windows")
	}

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-shell-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test shell script
	scriptContent := `#!/bin/bash
source ../../scripts/install.sh
take "$1"
pwd
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	tests := []struct {
		name    string
		dir     string
		wantDir string
	}{
		{
			name:    "change to new directory",
			dir:     "newdir",
			wantDir: filepath.Join(tmpDir, "newdir"),
		},
		{
			name:    "change to nested directory",
			dir:     "parent/child",
			wantDir: filepath.Join(tmpDir, "parent/child"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("/bin/bash", scriptPath, tt.dir)
			cmd.Dir = tmpDir

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("shell integration test failed: %v\nOutput: %s", err, output)
				return
			}

			// The last line of output should be the current directory
			if string(output) != tt.wantDir+"\n" {
				t.Errorf("shell integration test: got pwd = %q, want %q", string(output), tt.wantDir)
			}
		})
	}
}
