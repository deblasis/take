package shell

import (
	"os"
	"runtime"
	"strings"
)

// Shell represents a shell environment
type Shell interface {
	// Name returns the shell name
	Name() string
	// ChangeDir generates the command to change directory
	ChangeDir(path string) string
	// SetupScript returns the shell-specific setup script
	SetupScript() string
}

// GetCurrentShell detects and returns the current shell
func GetCurrentShell() Shell {
	// Check if we're running in PowerShell
	if _, ok := os.LookupEnv("PSModulePath"); ok {
		return &PowerShell{}
	}

	// On Windows, default to CMD if not PowerShell
	if runtime.GOOS == "windows" {
		return &CMD{}
	}

	// Check $SHELL environment variable
	shell := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(shell, "zsh"):
		return &Zsh{}
	case strings.HasSuffix(shell, "bash"):
		return &Bash{}
	default:
		// Default to Bash-compatible shell
		return &Bash{}
	}
}

// Zsh implementation
type Zsh struct{}

func (z *Zsh) Name() string { return "zsh" }
func (z *Zsh) ChangeDir(path string) string {
	return "cd " + shellQuote(path)
}
func (z *Zsh) SetupScript() string {
	return `take() {
	if [ -z "$1" ]; then
		echo "Usage: take <directory or git-url>" >&2
		return 1
	fi
	take_result=$(take-cli "$1")
	if [ $? -eq 0 ]; then
		cd "$take_result"
	else
		echo "$take_result" >&2
		return 1
	fi
}`
}

// Bash implementation
type Bash struct{}

func (b *Bash) Name() string { return "bash" }
func (b *Bash) ChangeDir(path string) string {
	return "cd " + shellQuote(path)
}
func (b *Bash) SetupScript() string {
	return `take() {
	if [ -z "$1" ]; then
		echo "Usage: take <directory or git-url>" >&2
		return 1
	fi
	take_result=$(take-cli "$1")
	if [ $? -eq 0 ]; then
		cd "$take_result"
	else
		echo "$take_result" >&2
		return 1
	fi
}`
}

// PowerShell implementation
type PowerShell struct{}

func (p *PowerShell) Name() string { return "powershell" }
func (p *PowerShell) ChangeDir(path string) string {
	return "Set-Location " + shellQuote(path)
}
func (p *PowerShell) SetupScript() string {
	return `function Take {
	param([string]$Path)
	if (-not $Path) {
		Write-Error "Usage: Take <directory or git-url>"
		return
	}
	$result = take-cli $Path
	if ($LASTEXITCODE -eq 0) {
		Set-Location $result
	} else {
		Write-Error $result
	}
}`
}

// CMD implementation
type CMD struct{}

func (c *CMD) Name() string { return "cmd" }
func (c *CMD) ChangeDir(path string) string {
	return "cd /d " + shellQuote(path)
}
func (c *CMD) SetupScript() string {
	return `@echo off
doskey take=for /f "tokens=*" %%i in ('take-cli $*') do cd /d %%i`
}

// shellQuote quotes a string for shell usage
func shellQuote(s string) string {
	if runtime.GOOS == "windows" {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return `'` + strings.ReplaceAll(s, `'`, `'\''`) + `'`
}
