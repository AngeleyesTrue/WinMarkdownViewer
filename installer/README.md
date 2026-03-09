# WinMarkdownViewer MSI 인스톨러 빌드

## 사전 요구 사항

- [Go](https://go.dev/dl/) (go.mod에 명시된 버전)
- [.NET SDK 8.0+](https://dotnet.microsoft.com/download/dotnet/8.0)
- [WiX Toolset v4](https://wixtoolset.org/)

## WiX 설치

```powershell
dotnet tool install --global wix
dotnet tool run wix extension add WixToolset.UI.wixext
```

## MSI 빌드

```powershell
# 기본 빌드 (버전은 git tag에서 자동 감지)
.\installer\build-msi.ps1

# 버전 지정 빌드
.\installer\build-msi.ps1 -Version "1.0.0"

# Debug 빌드 (디버그 심볼 포함)
.\installer\build-msi.ps1 -Version "1.0.0" -Configuration Debug

# Go 바이너리가 이미 있는 경우 (Go 빌드 생략)
.\installer\build-msi.ps1 -Version "1.0.0" -SkipGoBuild
```

## 출력 위치

빌드 결과물은 `dist/` 디렉토리에 생성됩니다:

```
dist/
  winmdview.exe                          # Go 바이너리
  WinMarkdownViewer-1.0.0-x64.msi       # MSI 인스톨러
```

## 테스트 실행

```powershell
# Pester 5.x 필요
Install-Module -Name Pester -Force -SkipPublisherCheck
Invoke-Pester ./installer/tests/
```

## CI/CD

GitHub에 `v*` 형식의 태그를 푸시하면 자동으로 MSI가 빌드되고 GitHub Release가 생성됩니다.

```bash
git tag v1.0.0
git push origin v1.0.0
```
