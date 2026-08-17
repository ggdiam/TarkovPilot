$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$previousGoos = [Environment]::GetEnvironmentVariable("GOOS", "Process")

Push-Location -LiteralPath $PSScriptRoot
try {
    $wails = Get-Command wails -ErrorAction SilentlyContinue
    if ($null -eq $wails) {
        throw "Wails CLI was not found. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    }

    [Environment]::SetEnvironmentVariable("GOOS", "windows", "Process")

    # app version — version.go is the source of truth
    # read via ReadAllText: Get-Content in PS 5.1 treats BOM-less UTF-8 as ANSI and mangles non-ASCII
    $versionGoPath = Join-Path $PSScriptRoot "version.go"
    $versionGo = [IO.File]::ReadAllText($versionGoPath)
    if ($versionGo -notmatch 'const Version = "([^"]+)"') {
        throw "Could not parse Version from version.go"
    }
    $version = $Matches[1]

    # if the version wasn't bumped since the last commit — auto-bump the patch.
    # git via cmd: in PS 5.1 redirected stderr of native commands turns into errors
    $headGo = cmd /c "git -C `"$PSScriptRoot`" show HEAD:app/version.go 2>nul"
    if ($LASTEXITCODE -eq 0 -and "$headGo" -match 'const Version = "([^"]+)"') {
        $headVersion = $Matches[1]
        if ([version]$version -le [version]$headVersion) {
            $v = [version]$headVersion
            $version = "{0}.{1}.{2}" -f $v.Major, $v.Minor, ($v.Build + 1)
            $versionGo = $versionGo -replace 'const Version = "[^"]+"', ('const Version = "' + $version + '"')
            [IO.File]::WriteAllText($versionGoPath, $versionGo, (New-Object System.Text.UTF8Encoding $false))
            Write-Host "version.go: $headVersion -> $version (auto-bump)" -ForegroundColor Yellow
        }
    }
    $wailsJsonPath = Join-Path $PSScriptRoot "wails.json"
    $wailsJson = [IO.File]::ReadAllText($wailsJsonPath)
    $updated = $wailsJson -replace '"productVersion":\s*"[^"]*"', ('"productVersion": "' + $version + '"')
    if ($updated -ne $wailsJson) {
        [IO.File]::WriteAllText($wailsJsonPath, $updated, (New-Object System.Text.UTF8Encoding $false))
        Write-Host "wails.json: productVersion -> $version"
    }

    & $wails.Source build
    if ($LASTEXITCODE -ne 0) {
        throw "Wails build failed with exit code $LASTEXITCODE"
    }

    $outputPath = Join-Path $PSScriptRoot "build\bin\TarkovPilot.exe"
    if (-not (Test-Path -LiteralPath $outputPath)) {
        throw "Build completed, but the output file was not found: $outputPath"
    }

    Write-Host "Done: $outputPath" -ForegroundColor Green
}
finally {
    [Environment]::SetEnvironmentVariable("GOOS", $previousGoos, "Process")
    Pop-Location
}
