# WinMarkdownViewer Installer for Windows
# Requires PowerShell 5.1 or later
# Supports piped execution: irm https://raw.githubusercontent.com/AngeleyesTrue/WinMarkdownViewer/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "AngeleyesTrue/WinMarkdownViewer"
$BinaryName = "winmdview.exe"

function Print-Info {
    Write-Host "[INFO] " -ForegroundColor Cyan -NoNewline
    Write-Host $args
}

function Print-Success {
    Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $args
}

function Print-Error {
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $args
}

function Print-Warning {
    Write-Host "[WARNING] " -ForegroundColor Yellow -NoNewline
    Write-Host $args
}

# Get latest release version.
#
# Resolved from the /releases/latest redirect rather than the REST API: the API
# allows only 60 unauthenticated requests/hour/IP, which a shared network can
# exhaust; the redirect is not part of the API and is not rate limited. It also
# means "latest stable" specifically, excluding drafts and prereleases.
function Get-LatestVersion {
    $latestUrl = "https://github.com/$Repo/releases/latest"
    $resolved = ""

    try {
        $response = Invoke-WebRequest -Uri $latestUrl -Method Head `
            -MaximumRedirection 10 -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop

        # Windows PowerShell 5.1 returns HttpWebResponse (ResponseUri);
        # PowerShell 6+ returns HttpResponseMessage (RequestMessage.RequestUri).
        $base = $response.BaseResponse
        if ($base.PSObject.Properties['RequestMessage']) {
            $resolved = [string]$base.RequestMessage.RequestUri.AbsoluteUri
        }
        elseif ($base.PSObject.Properties['ResponseUri']) {
            $resolved = [string]$base.ResponseUri.AbsoluteUri
        }
    }
    catch {
        $resolved = ""
    }

    $version = ""
    if ($resolved -match '/tag/(?<tag>[^/?#]+)') {
        $version = $Matches['tag'] -replace '^v', ''
    }

    if ([string]::IsNullOrEmpty($version)) {
        Print-Error "Could not determine the latest version from GitHub"
        Print-Info "GitHub may be unreachable from this network. You can:"
        Write-Host "  1. Install a specific version: .\install.ps1 -version 1.0.5"
        Write-Host "  2. Download the MSI installer instead: https://github.com/$Repo/releases"
        exit 1
    }

    Print-Success "Latest version: $version"
    return $version
}

# Download the portable archive and verify it against the release checksums.
function Download-Binary {
    param([string]$Version)

    $archiveName = "WinMarkdownViewer_${Version}_windows_amd64.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/v$Version/$archiveName"
    $checksumUrl = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"

    $tmpBase = [System.IO.Path]::GetTempPath()
    $tmpName = "winmdview-install-$([System.Guid]::NewGuid().ToString())"
    $tempDir = Join-Path $tmpBase $tmpName
    $archiveFile = Join-Path $tempDir $archiveName
    $checksumFile = Join-Path $tempDir "checksums.txt"

    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    Print-Info "Downloading from: $downloadUrl"

    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

        Invoke-WebRequest -Uri $downloadUrl -OutFile $archiveFile -UseBasicParsing
        Print-Success "Download completed"

        try {
            Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile -UseBasicParsing
            Print-Info "Verifying checksum..."

            $expectedLine = Get-Content $checksumFile | Select-String -Pattern $archiveName | Select-Object -First 1
            if ($expectedLine) {
                $expectedChecksum = ($expectedLine -split '\s+')[0]
                $actualChecksum = (Get-FileHash -Path $archiveFile -Algorithm SHA256).Hash.ToLower()

                if ($expectedChecksum -eq $actualChecksum) {
                    Print-Success "Checksum verified"
                }
                else {
                    Print-Error "Checksum mismatch!"
                    Print-Error "Expected: $expectedChecksum"
                    Print-Error "Actual:   $actualChecksum"
                    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
                    exit 1
                }
            }
        }
        catch {
            Print-Warning "Failed to verify checksum (continuing anyway)"
        }
    }
    catch {
        Print-Error "Download failed: $_"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    Print-Info "Extracting archive..."
    try {
        Expand-Archive -Path $archiveFile -DestinationPath $tempDir -Force
    }
    catch {
        Print-Error "Failed to extract archive: $_"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    $binaryPath = Join-Path $tempDir $BinaryName
    if (-not (Test-Path $binaryPath)) {
        Print-Error "Binary not found in archive"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    return $binaryPath
}

# Install the binary to a per-user, no-admin-required location.
function Install-Binary {
    param([string]$BinaryPath)

    if ($env:WINMDVIEW_INSTALL_DIR) {
        $targetDir = $env:WINMDVIEW_INSTALL_DIR
    }
    else {
        $targetDir = Join-Path $env:LOCALAPPDATA "Programs\WinMarkdownViewer"
    }

    if (-not (Test-Path $targetDir)) {
        Print-Info "Creating directory: $targetDir"
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    $targetPath = Join-Path $targetDir $BinaryName

    Print-Info "Installing to: $targetPath"

    try {
        Copy-Item -Path $BinaryPath -Destination $targetPath -Force
        Print-Success "Installed to: $targetPath"
    }
    catch {
        Print-Error "Failed to install: $_"
        Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue

    return $targetPath
}

# Add the install directory to the user's PATH.
function Add-ToPath {
    param([string]$TargetPath)

    $targetDir = Split-Path $TargetPath -Parent

    $pathEnv = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($pathEnv -like "*$targetDir*") {
        Print-Info "Already in PATH"
        return
    }

    Print-Info "Adding to PATH..."
    try {
        $newPath = "$pathEnv;$targetDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$targetDir"
        Print-Success "Added to PATH"
    }
    catch {
        Print-Warning "Failed to add to PATH automatically"
        Write-Host ""
        Print-Info 'Please add manually by running (PowerShell):'
        Write-Host ""
        Write-Host "    [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$targetDir', 'User')" -ForegroundColor Yellow
        Write-Host ""
    }
}

# Register the .md right-click context menu, matching what the MSI's core
# feature does. Non-interactive by design so piped execution (irm | iex)
# never blocks on stdin; opt out with -skip-register.
function Register-ContextMenu {
    param([string]$TargetPath)

    if ($env:WINMDVIEW_SKIP_REGISTER) {
        Print-Info "Skipped context menu registration. Run '$TargetPath --register' later to enable it."
        return
    }

    Print-Info "Registering .md right-click context menu..."
    try {
        & $TargetPath --register
        Print-Success "Context menu registered"
    }
    catch {
        Print-Warning "Failed to register context menu (continuing anyway): $_"
    }
}

function Verify-Installation {
    param([string]$TargetPath)

    if (-not (Test-Path $TargetPath)) {
        Print-Warning "Installation completed, verify manually"
        return
    }

    $productVersion = (Get-Item $TargetPath).VersionInfo.ProductVersion
    Print-Success "WinMarkdownViewer installed successfully!"
    if ($productVersion) {
        Write-Host "  Version: $productVersion"
    }
    Write-Host ""
    Print-Info "Usage:"
    Write-Host "    winmdview <file.md>          # Open a Markdown file" -ForegroundColor Cyan
    Write-Host "    winmdview --register         # Register the .md context menu" -ForegroundColor Cyan
    Write-Host "    winmdview --unregister       # Remove the context menu" -ForegroundColor Cyan
    Write-Host ""
    Print-Warning "Open a new terminal (or sign out/in) for the updated PATH to take effect."
}

function Main {
    param($Arguments)

    Write-Host ""
    Write-Host "=============================================================="
    Write-Host "               WinMarkdownViewer Installer"
    Write-Host "=============================================================="
    Write-Host ""

    if ($env:OS -ne "Windows_NT" -and -not $IsWindows) {
        Print-Error "WinMarkdownViewer is Windows-only."
        exit 1
    }

    $Version = ""
    $SkipRegister = $false

    for ($i = 0; $i -lt $Arguments.Count; $i++) {
        $arg = $Arguments[$i]
        if ($arg -eq "-version" -or $arg -eq "--version") {
            $Version = $Arguments[$i + 1]
            $i++
        }
        elseif ($arg -eq "-install-dir" -or $arg -eq "--install-dir") {
            $env:WINMDVIEW_INSTALL_DIR = $Arguments[$i + 1]
            $i++
        }
        elseif ($arg -eq "-skip-register" -or $arg -eq "--skip-register") {
            $SkipRegister = $true
        }
        elseif ($arg -eq "-h" -or $arg -eq "--help" -or $arg -eq "-help") {
            Write-Host "Usage: .\install.ps1 [OPTIONS]"
            Write-Host ""
            Write-Host "Options:"
            Write-Host "  -version VERSION       Install a specific version (default: latest)"
            Write-Host "  -install-dir DIR       Install to a custom directory"
            Write-Host "  -skip-register         Do not prompt to register the .md context menu"
            Write-Host "  -h, --help             Show this help message"
            Write-Host ""
            Write-Host "Piped execution:"
            Write-Host "  irm https://raw.githubusercontent.com/$Repo/main/install.ps1 | iex"
            exit 0
        }
        else {
            Print-Error "Unknown option: $arg"
            Write-Host "Use -h for usage information"
            exit 1
        }
    }

    if ($SkipRegister) {
        $env:WINMDVIEW_SKIP_REGISTER = "1"
    }

    if (-not $Version) {
        $Version = Get-LatestVersion
    }
    else {
        Print-Info "Installing version: $Version"
    }

    $binaryPath = Download-Binary -Version $Version
    $targetPath = Install-Binary -BinaryPath $binaryPath

    Add-ToPath -TargetPath $targetPath
    Register-ContextMenu -TargetPath $targetPath
    Verify-Installation -TargetPath $targetPath

    Write-Host ""
    Print-Success "Installation complete!"
    Write-Host ""
    Print-Info "Documentation: https://github.com/$Repo"
}

Main -Arguments $args
