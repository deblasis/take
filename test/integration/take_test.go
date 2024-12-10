package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/deblasis/take/internal/git"
	"github.com/deblasis/take/pkg/take"
)

func init() {
	// Build the binary before running tests
	cmd := exec.Command("../../scripts/build.sh")
	if err := cmd.Run(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}

	// Add the binary directory to PATH
	binDir, _ := filepath.Abs("../../bin")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

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

	// Create shell script with take function
	scriptContent := `#!/bin/bash
take() {
    if [[ "$1" =~ ^(https://|git@) ]]; then
        git clone "$1" && cd "$(basename "$1" .git)"
    else
        mkdir -p "$1" && cd "$1"
    fi
    pwd
}
take "$1"
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T)
		validate func(t *testing.T, dir string)
		wantErr  bool
	}{
		{
			name: "create new directory",
			args: []string{"newdir"},
			validate: func(t *testing.T, dir string) {
				if _, err := os.Stat(filepath.Join(tmpDir, "newdir")); os.IsNotExist(err) {
					t.Error("Directory was not created")
				}
			},
		},
		{
			name: "create nested directories",
			args: []string{"parent/child/grandchild"},
			validate: func(t *testing.T, dir string) {
				if _, err := os.Stat(filepath.Join(tmpDir, "parent/child/grandchild")); os.IsNotExist(err) {
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
			validate: func(t *testing.T, dir string) {
				if dir != filepath.Join(tmpDir, "existing") {
					t.Errorf("Expected to be in existing directory, got %s", dir)
				}
			},
		},
		{
			name: "clone git repository",
			args: []string{"https://github.com/deblasis/take.git"},
			validate: func(t *testing.T, dir string) {
				repoDir := filepath.Join(tmpDir, "take")
				if !git.IsGitRepo(repoDir) {
					t.Error("Not a valid git repository")
				}
				if dir != repoDir {
					t.Errorf("Expected to be in take directory, got %s", dir)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			// Clean up any existing directories
			if len(tt.args) > 0 {
				os.RemoveAll(filepath.Join(tmpDir, filepath.Base(tt.args[0])))
			}

			// Run command
			cmd := exec.Command("/bin/bash", scriptPath, tt.args[0])
			cmd.Dir = tmpDir
			output, err := cmd.Output()
			if (err != nil) != tt.wantErr {
				t.Errorf("take command error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, strings.TrimSpace(string(output)))
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

	// Create a simple test shell script
	scriptContent := `#!/bin/bash
take() {
    mkdir -p "$1" && cd "$1"
}
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
			// Clean up any existing directories
			os.RemoveAll(filepath.Join(tmpDir, tt.dir))

			cmd := exec.Command("/bin/bash", scriptPath, tt.dir)
			cmd.Dir = tmpDir

			output, err := cmd.Output()
			if err != nil {
				t.Errorf("shell integration test failed: %v\nOutput: %s", err, output)
				return
			}

			got := strings.TrimSpace(string(output))
			if got != tt.wantDir {
				t.Errorf("shell integration test: got pwd = %q, want %q", got, tt.wantDir)
			}
		})
	}
}

func TestTakeWindowsIntegration(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific tests on non-Windows platform")
	}

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-windows-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T)
		validate func(t *testing.T, result take.Result)
		wantErr  bool
	}{
		{
			name: "windows backslash path",
			args: []string{"test\\windows\\path"},
			validate: func(t *testing.T, result take.Result) {
				expected := filepath.Join(tmpDir, "test", "windows", "path")
				if result.FinalPath != expected {
					t.Errorf("Expected path %s, got %s", expected, result.FinalPath)
				}
			},
		},
		{
			name: "windows drive letter path",
			args: []string{"C:\\test\\path"},
			validate: func(t *testing.T, result take.Result) {
				if !filepath.IsAbs(result.FinalPath) {
					t.Error("Expected absolute path with drive letter")
				}
			},
		},
		{
			name: "windows UNC path",
			args: []string{"\\\\server\\share\\path"},
			validate: func(t *testing.T, result take.Result) {
				if !strings.HasPrefix(result.FinalPath, "\\\\") {
					t.Error("Expected UNC path")
				}
			},
		},
		{
			name: "windows home directory",
			args: []string{"~\\Documents\\test"},
			validate: func(t *testing.T, result take.Result) {
				home, _ := os.UserHomeDir()
				expected := filepath.Join(home, "Documents", "test")
				if result.FinalPath != expected {
					t.Errorf("Expected path %s, got %s", expected, result.FinalPath)
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
					WasCreated: true,
				}
				tt.validate(t, result)
			}
		})
	}
}

func TestTakeWindowsShellIntegration(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows shell integration tests on non-Windows platform")
	}

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-windows-shell-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test PowerShell script
	scriptContent := `
. ..\..\completions\take.ps1
take $args[0]
Get-Location | Select-Object -ExpandProperty Path
`
	scriptPath := filepath.Join(tmpDir, "test.ps1")
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
			dir:     "parent\\child",
			wantDir: filepath.Join(tmpDir, "parent", "child"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("powershell", "-File", scriptPath, tt.dir)
			cmd.Dir = tmpDir

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("PowerShell integration test failed: %v\nOutput: %s", err, output)
				return
			}

			// The last line of output should be the current directory
			if string(output) != tt.wantDir+"\r\n" {
				t.Errorf("PowerShell integration test: got pwd = %q, want %q", string(output), tt.wantDir)
			}
		})
	}
}
