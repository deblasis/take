function take {
    param(
        [Parameter(Mandatory=$true, Position=0)]
        [string]$Path
    )

    # Handle git URLs
    if ($Path -match '^(https://|git@)') {
        $repoName = $Path -replace '.*[:/]([^/]+)/([^/]+)(\.git)?$', '$2'
        git clone $Path $repoName
        if ($?) {
            Set-Location $repoName
        }
        return
    }

    # Expand ~ to user home directory
    if ($Path.StartsWith("~")) {
        $Path = $Path.Replace("~", $HOME)
    }

    # Create directory if it doesn't exist
    if (!(Test-Path $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }

    # Change to the directory
    Set-Location $Path
}

# Add tab completion for existing directories
Register-ArgumentCompleter -CommandName take -ParameterName Path -ScriptBlock {
    param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)

    Get-ChildItem -Directory -Path "$wordToComplete*" |
        ForEach-Object {
            [System.Management.Automation.CompletionResult]::new(
                $_.FullName,
                $_.Name,
                'ParameterValue',
                $_.Name
            )
        }
}