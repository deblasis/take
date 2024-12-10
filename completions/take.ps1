function take {
    param([string]$Path)

    if ($Path -match '^https?://') {
        $repoName = [System.IO.Path]::GetFileNameWithoutExtension($Path)
        if (git clone $Path $repoName) {
            Set-Location $repoName
        } else {
            return
        }
    } else {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
        Set-Location $Path
    }
    Get-Location | Select-Object -ExpandProperty Path
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