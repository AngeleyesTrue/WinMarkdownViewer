---
id: SPEC-INSTALL-001
title: "MSI Installer - WiX Toolset v4 Based Package"
version: 1.0.0
status: completed
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P2
lifecycle: spec-first
tags: [wix, msi, installer, registry, context-menu]
---

# SPEC-INSTALL-001: MSI 설치 프로그램 - WiX Toolset v4 기반 패키지

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- x64 아키텍처 전용
- .NET Framework 불필요 (WiX v4는 .NET 6+ SDK 기반 빌드)

### 1.2 기술 스택
- WiX Toolset v4 (NuGet 패키지: `wix`)
- PowerShell 5.1+ (빌드 스크립트)
- GitHub Actions (CI/CD 워크플로우)
- Go 1.26.0 (winmdview.exe 바이너리 빌드)

### 1.3 디렉토리 구조
```
WinMarkdownViewer/
  installer/
    wix/
      Package.wxs              # WiX v4 메인 패키지 정의
      Directories.wxs          # 디렉토리 구조 정의
      Components.wxs           # 컴포넌트 및 파일 정의
      Registry.wxs             # 레지스트리 항목 정의
      Shortcuts.wxs            # 시작 메뉴 바로가기 정의
      Variables.wxi            # 공통 변수 (버전, 제품명 등)
    tests/
      build-msi.Tests.ps1      # Pester 테스트
    build-msi.ps1              # MSI 빌드 스크립트
    README.md                  # 빌드 방법 안내
  .github/
    workflows/
      release.yml              # 릴리스 CI/CD 워크플로우
```

### 1.4 참고 사항
- **docs/structure.md 업데이트 필요**: 프로젝트 문서에서 `Product.wxs`로 기재된 파일명을 WiX v4 표준인 `Package.wxs`로 수정해야 한다 (이 SPEC 범위 외, 별도 작업 필요)

### 1.5 선행 조건
- SPEC-WIN-001 완료 (컨텍스트 메뉴 레지스트리 로직이 구현되어 있어야 함)
- winmdview.exe 바이너리가 정상 빌드 가능한 상태

---

## 2. Assumptions (가정)

### 2.1 기술적 가정
- A1: WiX Toolset v4가 GitHub Actions의 windows-latest 러너에서 dotnet tool로 설치 가능하다
- A2: winmdview.exe는 단일 바이너리로 외부 DLL 의존성이 없다 (CGO 미사용)
- A3: WebView2 Runtime은 Windows 10/11에 기본 설치되어 있거나 사용자가 별도 설치한다
- A4: MSI 패키지의 대상 설치 경로는 `%ProgramFiles%\WinMarkdownViewer\`이다

### 2.2 비즈니스 가정
- A5: 레지스트리 등록은 HKCU 범위에서 수행하므로 관리자 권한이 불필요하다. 단, MSI 설치 자체는 %ProgramFiles% 쓰기를 위해 관리자 권한이 필요할 수 있다
- A6: 한 시스템에 하나의 WinMarkdownViewer 인스턴스만 설치된다
- A7: 업그레이드 시 기존 설정 파일(%APPDATA%)은 보존된다

### 2.3 범위 외 (Out of Scope)
- 자동 업데이트 기능 (WinSparkle 등)
- Chocolatey/Scoop/winget 패키지 매니저 배포
- 포터블(ZIP) 배포판
- ARM64 아키텍처 지원

---

## 3. Requirements (요구사항)

### 3.1 설치 요구사항

**REQ-INST-001** (Ubiquitous)
시스템은 **항상** MSI 패키지를 통해 winmdview.exe를 `%ProgramFiles%\WinMarkdownViewer\` 디렉토리에 설치해야 한다.

**REQ-INST-002** (Event-Driven)
**WHEN** 사용자가 MSI 파일을 더블클릭하여 설치를 완료하면 **THEN** 시스템은 다음 항목을 자동으로 등록해야 한다:
- .md 파일 우클릭 컨텍스트 메뉴 ("마크다운 뷰어로 열기")
- 시작 메뉴 바로가기 (WinMarkdownViewer 폴더 하위)

**REQ-INST-003** (Optional)
**가능하면** 설치 과정에서 .md 파일 연결(기본 프로그램) 등록 옵션을 제공한다.
- 설치 UI에서 체크박스로 선택 가능 (기본값: 선택 안 됨)
- 선택 시 .md 파일 더블클릭 = WinMarkdownViewer로 열기

**REQ-INST-004** (Ubiquitous)
시스템은 **항상** 프로그램 추가/제거(Apps & Features) 목록에 "WinMarkdownViewer"를 표시해야 한다.
- 제품명, 버전, 게시자, 설치 크기 포함

### 3.2 제거 요구사항

**REQ-UNINST-001** (Event-Driven)
**WHEN** 사용자가 프로그램 추가/제거에서 WinMarkdownViewer를 제거하면 **THEN** 시스템은 다음을 수행해야 한다:
- 컨텍스트 메뉴 레지스트리 항목 삭제
- .md 파일 연결 해제 (등록된 경우)
- 시작 메뉴 바로가기 삭제
- 설치 디렉토리의 모든 파일 삭제

**REQ-UNINST-002** (Unwanted)
시스템은 제거 시 사용자 설정 파일(`%APPDATA%\WinMarkdownViewer\`)을 삭제**하지 않아야 한다**.

### 3.3 업그레이드 요구사항

**REQ-UPGRADE-001** (Event-Driven)
**WHEN** 새 버전의 MSI를 설치하면 **THEN** 시스템은 기존 버전을 자동으로 제거하고 새 버전을 설치해야 한다 (Major Upgrade 패턴).

**REQ-UPGRADE-002** (State-Driven)
**IF** 기존 버전이 설치되어 있는 상태에서 동일 버전의 MSI를 실행하면 **THEN** 시스템은 WiX 기본 Repair 동작에 의존한다. 커스텀 수리 UI는 구현하지 않는다.

### 3.4 빌드 요구사항

**REQ-BUILD-001** (Ubiquitous)
시스템은 **항상** `installer/build-msi.ps1` 스크립트 하나로 MSI 패키지를 빌드할 수 있어야 한다.
- Go 바이너리 빌드 + WiX 컴파일 + MSI 생성을 순차적으로 수행
- 빌드 결과물: `WinMarkdownViewer-{version}-x64.msi`

**REQ-BUILD-002** (Event-Driven)
**WHEN** GitHub에 `v*` 태그가 푸시되면 **THEN** GitHub Actions 워크플로우가 자동으로:
1. Go 바이너리를 빌드한다
2. MSI 패키지를 생성한다
3. GitHub Release에 MSI 파일을 첨부한다

### 3.5 레지스트리 요구사항

**REQ-REG-001** (Ubiquitous)
시스템은 **항상** 다음 레지스트리 키를 등록해야 한다:
- `HKCU\Software\Classes\.md\shell\WinMarkdownViewer` (컨텍스트 메뉴)
- `HKCU\Software\Classes\.md\shell\WinMarkdownViewer\command` (실행 명령)
- `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\{ProductCode}` (프로그램 추가/제거)

**REQ-REG-002** (State-Driven)
**IF** 사용자가 .md 파일 연결 옵션을 선택한 상태에서 설치하면 **THEN** 시스템은 추가로 다음 레지스트리 키를 등록해야 한다:
- `HKCU\Software\Classes\.md\OpenWithProgids` (파일 연결)
- `HKCU\Software\Classes\WinMarkdownViewer.md` (ProgId 정의)

---

## 4. Specifications (명세)

### 4.1 WiX v4 패키지 구조

#### Package.wxs
- `<Package>` 요소: 제품명, 버전, 제조업체, UpgradeCode(고정 GUID)
- `<MajorUpgrade>` 요소: 이전 버전 자동 제거, 다운그레이드 방지
- `<Feature>` 정의: 필수(Core) + 선택(FileAssociation)

#### Components.wxs
- winmdview.exe 파일 컴포넌트 (KeyPath)
- 각 컴포넌트는 고유 GUID 보유

#### Registry.wxs
- 컨텍스트 메뉴 레지스트리 컴포넌트
- 파일 연결 레지스트리 컴포넌트 (선택적 Feature에 포함)
- 설치/제거 시 자동 등록/해제

#### Shortcuts.wxs
- 시작 메뉴 폴더 생성
- 시작 메뉴 바로가기 (winmdview.exe 대상)
- 바로가기 아이콘 설정

### 4.2 빌드 스크립트 (build-msi.ps1)

```
[매개변수]
- -Version: 빌드 버전 (기본값: git describe)
- -Configuration: Release|Debug (기본값: Release)
- -OutputDir: 출력 디렉토리 (기본값: ./dist)

[빌드 단계]
1. Go 바이너리 빌드: go build -ldflags "-s -w" -o dist/winmdview.exe
2. WiX 컴파일: dotnet tool run wix build -o dist/WinMarkdownViewer-{version}-x64.msi
3. 빌드 결과 검증: MSI 파일 존재 여부 및 크기 확인
```

### 4.3 GitHub Actions 워크플로우 (release.yml)

```
[트리거]
- push.tags: v*

[단계]
1. Go 환경 설정 (Go 1.26.0)
2. .NET SDK 설정 (WiX v4 빌드용)
3. WiX Toolset 설치 (dotnet tool install)
4. build-msi.ps1 실행
5. GitHub Release 생성 + MSI 첨부
```

### 4.4 설치 흐름

```
MSI 더블클릭
  -> Windows Installer 실행
  -> 라이선스 동의 화면
  -> 설치 경로 선택 (기본: %ProgramFiles%\WinMarkdownViewer)
  -> 옵션: .md 파일 연결 체크박스
  -> 설치 진행
    -> winmdview.exe 복사
    -> 레지스트리 항목 등록
    -> 시작 메뉴 바로가기 생성
    -> (선택) 파일 연결 등록
  -> 설치 완료
```

---

## 5. Traceability (추적성)

| 요구사항 ID | plan.md 참조 | acceptance.md 참조 | 관련 파일 |
|-------------|-------------|-------------------|-----------|
| REQ-INST-001 | Task 1.1 | AC-INST-001 | installer/wix/Components.wxs |
| REQ-INST-002 | Task 1.2, 1.3 | AC-INST-002 | installer/wix/Registry.wxs, Shortcuts.wxs |
| REQ-INST-003 | Task 1.4 | AC-INST-003 | installer/wix/Registry.wxs, Package.wxs |
| REQ-INST-004 | Task 1.1 | AC-INST-004 | installer/wix/Package.wxs |
| REQ-UNINST-001 | Task 2.1 | AC-UNINST-001 | installer/wix/Registry.wxs |
| REQ-UNINST-002 | Task 2.2 | AC-UNINST-002 | installer/wix/Package.wxs |
| REQ-UPGRADE-001 | Task 3.1 | AC-UPGRADE-001 | installer/wix/Package.wxs |
| REQ-UPGRADE-002 | Task 3.2 | AC-UPGRADE-002 | installer/wix/Package.wxs |
| REQ-BUILD-001 | Task 4.1 | AC-BUILD-001 | installer/build-msi.ps1 |
| REQ-BUILD-002 | Task 4.2 | AC-BUILD-002 | .github/workflows/release.yml |
| REQ-REG-001 | Task 1.2 | AC-REG-001 | installer/wix/Registry.wxs |
| REQ-REG-002 | Task 1.4 | AC-REG-002 | installer/wix/Registry.wxs |
