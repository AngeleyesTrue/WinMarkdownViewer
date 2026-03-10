# WinMarkdownViewer 빌드 스크립트
# 릴리스/개발 빌드, 테스트, 정리 기능을 제공한다.

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "dev", "test", "clean", "help")]
    [string]$Target = "help"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# --- 헬퍼 함수 ---

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

function Test-GoCommand {
    $null = Get-Command "go" -ErrorAction SilentlyContinue
    return $?
}

# --- 타겟 함수 ---

function Invoke-GenerateVersionInfo {
    $versionJson = "cmd/winmdview/versioninfo.json"
    $sysoFile = "cmd/winmdview/resource.syso"

    if (-not (Test-Path $versionJson)) {
        Write-Info "versioninfo.json 이 없습니다. 리소스 임베딩을 건너뜁니다."
        return
    }

    $null = Get-Command "goversioninfo" -ErrorAction SilentlyContinue
    if (-not $?) {
        Write-Info "goversioninfo 가 설치되지 않았습니다. 리소스 임베딩을 건너뜁니다."
        Write-Info "설치: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest"
        return
    }

    Write-Info "goversioninfo 로 리소스 생성 중..."
    Push-Location "cmd/winmdview"
    try {
        goversioninfo -64 -o resource.syso versioninfo.json
    } finally {
        Pop-Location
    }
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] goversioninfo 실패" -ForegroundColor Red
        exit 1
    }
    Write-Success "리소스 파일 생성: $sysoFile"
}

function Invoke-Build {
    Write-Step "릴리스 빌드"

    if (-not (Test-GoCommand)) {
        Write-Host "[ERROR] Go 가 설치되지 않았거나 PATH 에 없습니다." -ForegroundColor Red
        exit 1
    }

    if (-not (Test-Path "dist")) {
        New-Item -ItemType Directory -Path "dist" -Force | Out-Null
    }

    Invoke-GenerateVersionInfo

    $ldflags = "-s -w -H windowsgui"
    Write-Info "ldflags: $ldflags"

    go build -ldflags="$ldflags" -o "dist/winmdview.exe" ./cmd/winmdview
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] 빌드 실패" -ForegroundColor Red
        exit 1
    }

    $size = (Get-Item "dist/winmdview.exe").Length
    Write-Success "dist/winmdview.exe 빌드 완료 ($([math]::Round($size / 1MB, 2)) MB)"
}

function Invoke-Dev {
    Write-Step "개발 빌드"

    if (-not (Test-GoCommand)) {
        Write-Host "[ERROR] Go 가 설치되지 않았거나 PATH 에 없습니다." -ForegroundColor Red
        exit 1
    }

    if (-not (Test-Path "dist")) {
        New-Item -ItemType Directory -Path "dist" -Force | Out-Null
    }

    Invoke-GenerateVersionInfo

    Write-Info "콘솔 창 표시 모드 (디버그용)"

    go build -o "dist/winmdview-dev.exe" ./cmd/winmdview
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] 개발 빌드 실패" -ForegroundColor Red
        exit 1
    }

    $size = (Get-Item "dist/winmdview-dev.exe").Length
    Write-Success "dist/winmdview-dev.exe 빌드 완료 ($([math]::Round($size / 1MB, 2)) MB)"
}

function Invoke-Test {
    Write-Step "테스트 실행"

    if (-not (Test-GoCommand)) {
        Write-Host "[ERROR] Go 가 설치되지 않았거나 PATH 에 없습니다." -ForegroundColor Red
        exit 1
    }

    go test -coverprofile=coverage.out ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] 테스트 실패" -ForegroundColor Red
        exit 1
    }

    Write-Success "테스트 통과"
}

function Invoke-Clean {
    Write-Step "정리"

    $targets = @("dist/winmdview.exe", "dist/winmdview-dev.exe", "winmdview.exe", "winmdview-dev.exe", "coverage.out", "cmd/winmdview/resource.syso")

    # *.out 파일도 정리 대상에 추가
    $outFiles = Get-ChildItem -Path "." -Filter "*.out" -File -ErrorAction SilentlyContinue
    foreach ($f in $outFiles) {
        if ($targets -notcontains $f.Name) {
            $targets += $f.Name
        }
    }

    foreach ($target in $targets) {
        if (Test-Path $target) {
            Remove-Item $target -Force
            Write-Info "삭제: $target"
        }
    }

    Write-Success "정리 완료"
}

function Invoke-Help {
    Write-Host ""
    Write-Host "WinMarkdownViewer 빌드 스크립트" -ForegroundColor White
    Write-Host ""
    Write-Host "사용법: .\build.ps1 <타겟>" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "타겟:"
    Write-Host "  build   릴리스 빌드 (-H windowsgui, 콘솔 창 숨김)"
    Write-Host "  dev     개발 빌드 (콘솔 창 표시, 디버그용)"
    Write-Host "  test    테스트 실행 (커버리지 포함)"
    Write-Host "  clean   빌드 산출물 정리"
    Write-Host "  help    이 도움말 표시"
    Write-Host ""
}

# --- 메인 ---

switch ($Target) {
    "build" { Invoke-Build }
    "dev"   { Invoke-Dev }
    "test"  { Invoke-Test }
    "clean" { Invoke-Clean }
    "help"  { Invoke-Help }
}
