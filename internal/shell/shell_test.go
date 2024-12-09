package shell

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestGetCurrentShell(t *testing.T) {
	// Save original environment
	origShell := os.Getenv("SHELL")
	origPSModulePath := os.Getenv("PSModulePath")
	defer func() {
		os.Setenv("SHELL", origShell)
		os.Setenv("PSModulePath", origPSModulePath)
	}()

	tests := []struct {
		name     string
		setup    func()
		want     string
		platform string
	}{
		{
			name: "detect zsh",
			setup: func() {
				os.Setenv("SHELL", "/bin/zsh")
				os.Unsetenv("PSModulePath")
			},
			want:     "zsh",
			platform: "linux",
		},
		{
			name: "detect bash",
			setup: func() {
				os.Setenv("SHELL", "/bin/bash")
				os.Unsetenv("PSModulePath")
			},
			want:     "bash",
			platform: "linux",
		},
		{
			name: "detect powershell",
			setup: func() {
				os.Setenv("PSModulePath", "some/path")
			},
			want:     "powershell",
			platform: "windows",
		},
		{
			name: "default to cmd on windows",
			setup: func() {
				os.Unsetenv("PSModulePath")
			},
			want:     "cmd",
			platform: "windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.platform != runtime.GOOS {
				t.Skipf("Skipping test on %s (requires %s)", runtime.GOOS, tt.platform)
			}

			tt.setup()
			got := GetCurrentShell()
			if got.Name() != tt.want {
				t.Errorf("GetCurrentShell() = %v, want %v", got.Name(), tt.want)
			}
		})
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		platform string
	}{
		{
			name:     "simple path unix",
			input:    "/path/to/dir",
			want:     "'/path/to/dir'",
			platform: "linux",
		},
		{
			name:     "path with spaces unix",
			input:    "/path/to/my dir",
			want:     "'/path/to/my dir'",
			platform: "linux",
		},
		{
			name:     "path with single quotes unix",
			input:    "/path/to/O'Neil",
			want:     "'/path/to/O'\\''Neil'",
			platform: "linux",
		},
		{
			name:     "simple path windows",
			input:    `C:\path\to\dir`,
			want:     `"C:\path\to\dir"`,
			platform: "windows",
		},
		{
			name:     "path with spaces windows",
			input:    `C:\path\to\my dir`,
			want:     `"C:\path\to\my dir"`,
			platform: "windows",
		},
		{
			name:     "path with quotes windows",
			input:    `C:\path\to\"dir"`,
			want:     `"C:\path\to\""dir"""`,
			platform: "windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.platform != runtime.GOOS {
				t.Skipf("Skipping test on %s (requires %s)", runtime.GOOS, tt.platform)
			}

			got := shellQuote(tt.input)
			if got != tt.want {
				t.Errorf("shellQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellImplementations(t *testing.T) {
	shells := []Shell{
		&Zsh{},
		&Bash{},
		&PowerShell{},
		&CMD{},
	}

	for _, sh := range shells {
		t.Run(sh.Name(), func(t *testing.T) {
			// Test Name()
			if sh.Name() == "" {
				t.Error("Name() returned empty string")
			}

			// Test ChangeDir()
			cd := sh.ChangeDir("/test/path")
			if cd == "" {
				t.Error("ChangeDir() returned empty string")
			}

			// Test SetupScript()
			script := sh.SetupScript()
			if script == "" {
				t.Error("SetupScript() returned empty string")
			}

			// Verify script contains essential elements
			if !strings.Contains(script, "take") {
				t.Error("SetupScript() missing 'take' function/alias")
			}
		})
	}
}
