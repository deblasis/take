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

# take - Create a new directory and change to it
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
'@

Add-Content -Path $profilePath -Value $takeFunction

Write-Host "Installed Take function to $profilePath"
Write-Host "Please restart PowerShell or run: . $profilePath"