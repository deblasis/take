# take

A cross-platform implementation of the ZSH `take` command. Creates a new directory and changes to it in one command. Also supports cloning git repositories.

## Features

- Create and change to a new directory in one command
- Create parent directories automatically
- Clone git repositories (HTTPS and SSH)
- Cross-platform support (Linux, macOS, Windows)
- Shell integration (Bash, Zsh, PowerShell, CMD)

## Installation

### Prerequisites

- Go 1.16 or later
- Git (optional, for repository cloning)

### From Source

```bash
go install github.com/deblasis/take/cmd/take@latest
```

### Shell Integration

#### Unix (Bash/Zsh)

```bash
curl -o- https://raw.githubusercontent.com/deblasis/take/main/scripts/install.sh | bash
```

Or manually add to your `.bashrc`/`.zshrc`:

```bash
take() {
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
}
```

#### PowerShell

```powershell
iwr https://raw.githubusercontent.com/deblasis/take/main/scripts/install.ps1 -useb | iex
```

Or manually add to your PowerShell profile:

```powershell
function Take {
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
}
```

## Usage

### Create and Change to Directory

```bash
# Create single directory
take mynewdir

# Create nested directories
take path/to/new/dir

# Use absolute paths
take /absolute/path/to/dir

# Use home directory
take ~/new/dir
```

### Clone Git Repository

```bash
# Clone via HTTPS
take https://github.com/user/repo.git

# Clone via SSH
take git@github.com:user/repo.git

# Shallow clone
take -depth 1 https://github.com/user/repo.git
```

### Options

```
-depth N    Git clone depth (0 for full clone)
-force      Force operation even if directory exists
-version    Show version information
```

## Development

### Building

```bash
go build ./cmd/take
```

### Testing

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.