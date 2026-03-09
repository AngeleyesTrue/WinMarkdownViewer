# WinMarkdownViewer MSI Build Script
# Builds Go binary and creates MSI installer using WiX Toolset v4

[CmdletBinding()]
param(
    [Parameter()]
    [string]$Version,

    [Parameter()]
    [ValidateSet("Release", "Debug")]
    [string]$Configuration = "Release",

    [Parameter()]
    [string]$OutputDir = "./dist",

    [Parameter()]
    [switch]$SkipGoBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# --- Helper Functions ---

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Yellow
}

function Test-Command {
    param([string]$Command)
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

# --- Version Detection ---

if (-not $Version) {
    Write-Info "No version specified, detecting from git..."
    try {
        $Version = (git describe --tags --always 2>$null)
        if (-not $Version) {
            $Version = "0.0.0-dev"
        }
        # Strip leading 'v' if present
        $Version = $Version.TrimStart('v')
    } catch {
        $Version = "0.0.0-dev"
    }
}

Write-Host "WinMarkdownViewer MSI Build" -ForegroundColor White
Write-Host "  Version:       $Version"
Write-Host "  Configuration: $Configuration"
Write-Host "  Output:        $OutputDir"
Write-Host "  Skip Go Build: $SkipGoBuild"

# --- Step 1: Validate Prerequisites ---

Write-Step "Validating prerequisites"

if (-not $SkipGoBuild) {
    if (-not (Test-Command "go")) {
        Write-Host "[ERROR] Go is not installed or not in PATH" -ForegroundColor Red
        exit 1
    }
    Write-Success "Go found: $(go version)"
} else {
    Write-Info "Skipping Go check (SkipGoBuild specified)"
}

if (-not (Test-Command "dotnet")) {
    Write-Host "[ERROR] .NET SDK is not installed or not in PATH" -ForegroundColor Red
    exit 1
}
Write-Success ".NET SDK found: $(dotnet --version)"

# Check WiX availability (global or local tool)
$wixCommand = $null
try {
    $wixVersion = & wix --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        $wixCommand = "wix"
        Write-Success "WiX Toolset found (global): $wixVersion"
    }
} catch {}

if (-not $wixCommand) {
    try {
        $wixVersion = dotnet tool run wix --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            $wixCommand = "dotnet-tool-run-wix"
            Write-Success "WiX Toolset found (local): $wixVersion"
        }
    } catch {}
}

if (-not $wixCommand) {
    Write-Host "[ERROR] WiX Toolset is not installed. Run: dotnet tool install --global wix" -ForegroundColor Red
    exit 1
}

# --- Step 2: Prepare Output Directory ---

Write-Step "Preparing output directory"

# Resolve OutputDir relative to PowerShell's current directory (not .NET's)
if (-not [System.IO.Path]::IsPathRooted($OutputDir)) {
    $OutputDir = Join-Path $PWD $OutputDir
}
$OutputDir = [System.IO.Path]::GetFullPath($OutputDir)
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    Write-Info "Created output directory: $OutputDir"
} else {
    Write-Info "Output directory exists: $OutputDir"
}

# --- Step 3: Build Go Binary ---

$binaryPath = Join-Path $OutputDir "winmdview.exe"

if ($SkipGoBuild) {
    Write-Step "Skipping Go build (SkipGoBuild specified)"
    if (-not (Test-Path $binaryPath)) {
        Write-Host "[ERROR] Binary not found at $binaryPath. Cannot skip Go build." -ForegroundColor Red
        exit 1
    }
    Write-Info "Using existing binary: $binaryPath"
} else {
    Write-Step "Building Go binary"

    $ldflags = "-s -w -H windowsgui"
    if ($Configuration -eq "Debug") {
        $ldflags = "-H windowsgui"
    }

    Write-Info "Building with ldflags: $ldflags"

    go build -ldflags="$ldflags" -o "$binaryPath" ./cmd/winmdview
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Go build failed" -ForegroundColor Red
        exit 1
    }

    $binarySize = (Get-Item $binaryPath).Length
    Write-Success "Binary built: $binaryPath ($([math]::Round($binarySize / 1MB, 2)) MB)"
}

# --- Step 4: Build MSI with WiX ---

Write-Step "Building MSI installer"

$projectRoot = Split-Path -Parent $PSScriptRoot
$wixDir = Join-Path $PSScriptRoot "wix"
$msiFileName = "WinMarkdownViewer-$Version-x64.msi"
$msiPath = Join-Path $OutputDir $msiFileName

# Collect all .wxs files in the wix directory
$wxsFiles = Get-ChildItem -Path $wixDir -Filter "*.wxs" | ForEach-Object { $_.FullName }

if ($wxsFiles.Count -eq 0) {
    Write-Host "[ERROR] No .wxs files found in $wixDir" -ForegroundColor Red
    exit 1
}

Write-Info "WiX source files:"
foreach ($wxs in $wxsFiles) {
    Write-Info "  - $(Split-Path -Leaf $wxs)"
}

$wixArgs = @()
$wixArgs += "build"
foreach ($wxs in $wxsFiles) {
    $wixArgs += $wxs
}
$wixArgs += "-d"
$wixArgs += "ProductVersion=$Version"
$wixArgs += "-bindpath"
$wixArgs += $OutputDir
$wixArgs += "-o"
$wixArgs += $msiPath

if ($wixCommand -eq "dotnet-tool-run-wix") {
    Write-Info "Running: dotnet tool run wix $($wixArgs -join ' ')"
    & dotnet tool run wix @wixArgs
} else {
    Write-Info "Running: wix $($wixArgs -join ' ')"
    & wix @wixArgs
}
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] WiX build failed" -ForegroundColor Red
    exit 1
}

# --- Step 5: Verify Output ---

Write-Step "Verifying output"

if (-not (Test-Path $msiPath)) {
    Write-Host "[ERROR] MSI file was not created: $msiPath" -ForegroundColor Red
    exit 1
}

$msiSize = (Get-Item $msiPath).Length
if ($msiSize -eq 0) {
    Write-Host "[ERROR] MSI file is empty: $msiPath" -ForegroundColor Red
    exit 1
}

# --- Summary ---

Write-Step "Build Summary"
Write-Success "MSI installer created successfully"
Write-Host ""
Write-Host "  File: $msiFileName"
Write-Host "  Path: $msiPath"
Write-Host "  Size: $([math]::Round($msiSize / 1MB, 2)) MB"
Write-Host ""
