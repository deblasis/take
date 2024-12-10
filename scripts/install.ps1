# PowerShell installation script for take command

# Get PowerShell profile path
$profilePath = $PROFILE.CurrentUserAllHosts
$profileDir = Split-Path -Parent $profilePath

# Create profile directory if it doesn't exist
if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
}

# Create profile file if it doesn't exist
if (-not (Test-Path $profilePath)) {
    New-Item -ItemType File -Path $profilePath -Force | Out-Null
}

# Check if take function already exists
$profileContent = Get-Content $profilePath -ErrorAction SilentlyContinue
if ($profileContent -match "function Take\s*{") {
    Write-Host "Take function already exists in $profilePath"
    exit 0
}

# Add take function to profile
$takeFunction = @'

# take - Create a new directory and change to it, or download and extract archives
function Take {
    param([string]$Path)

    if (-not $Path) {
        Write-Error "Usage: Take <directory|git-url|archive-url>"
        return
    }

    # Handle URLs with query parameters
    $target = $Path
    if ($target -like "*?*") {
        $target = $target.Split("?")[0]
    }

    # Check if it's a URL
    if ($target -match "^(https?|ftp|git|ssh|rsync).*$" -or $target -match "^[A-Za-z0-9]+@.*$") {
        # Create a temporary directory for downloads
        $tmpdir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_.FullName }
        try {
            # Handle different URL types
            if ($target -match "\.(tar\.gz|tgz|tar\.bz2|tar\.xz)$") {
                # Download and extract tarball
                $archive = Join-Path $tmpdir.FullName "archive.tar"
                Invoke-WebRequest -Uri $target -OutFile $archive
                if (-not $?) {
                    Write-Error "Failed to download archive"
                    return
                }
                $dir = tar -tf $archive | Select-Object -First 1 | ForEach-Object { $_.Split("/")[0] }
                tar xf $archive
                if (-not $?) {
                    Write-Error "Failed to extract archive"
                    return
                }
                Set-Location $dir
            }
            elseif ($target -match "\.zip$") {
                # Download and extract ZIP
                $archive = Join-Path $tmpdir.FullName "archive.zip"
                Invoke-WebRequest -Uri $target -OutFile $archive
                if (-not $?) {
                    Write-Error "Failed to download archive"
                    return
                }
                $dir = (Get-Content $archive -Raw | Select-String -Pattern "^.*?/").Matches[0].Value.TrimEnd("/")
                Expand-Archive -Path $archive -DestinationPath .
                if (-not $?) {
                    Write-Error "Failed to extract archive"
                    return
                }
                Set-Location $dir
            }
            elseif ($target -match "\.git$" -or $target -match "^git@") {
                # Clone git repository
                $repo = Split-Path -Leaf $target
                $repo = $repo -replace "\.git$", ""
                git clone $target $repo
                if (-not $?) {
                    Write-Error "Failed to clone repository"
                    return
                }
                Set-Location $repo
            }
            else {
                Write-Error "Unsupported URL format"
                return
            }
        }
        finally {
            Remove-Item -Recurse -Force $tmpdir -ErrorAction SilentlyContinue
        }
    }
    else {
        # Create and change to directory
        New-Item -ItemType Directory -Path $target -Force | Out-Null
        Set-Location $target
    }
}
'@

Add-Content -Path $profilePath -Value $takeFunction

Write-Host "Installed Take function to $profilePath"
Write-Host "Please restart PowerShell or run: . $profilePath"