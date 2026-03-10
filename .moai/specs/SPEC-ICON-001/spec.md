---
id: SPEC-ICON-001
title: Application Icon & File Association Icon
version: 1.0.0
status: Planned
created: 2026-03-10
updated: 2026-03-10
author: Claud Archive
priority: High
related_specs:
  - SPEC-WIN-001
  - SPEC-INSTALL-001
lifecycle: spec-first
---

# SPEC-ICON-001: 애플리케이션 아이콘 및 파일 연결 아이콘 개선

## 1. 환경 (Environment)

### 1.1 현재 상태

- `assets/icon.ico`: 16x16 단일 크기, 1,118 바이트 (약 1.1KB)
- `go:embed`를 통해 `assets.IconData`로 런타임 바이트 배열 제공
- `systray.SetIcon(t.iconData)`로 시스템 트레이 아이콘에만 사용
- `.syso` 리소스 파일 없음 - exe에 Windows 리소스 아이콘 미포함
- 탐색기에서 exe 파일이 기본 아이콘으로 표시됨
- WiX `Package.wxs`에서 `<Icon Id="AppIcon.exe" SourceFile="$(var.ExeFileName)" />`으로 exe를 아이콘 소스로 참조하지만, exe에 리소스가 없어 정상 동작하지 않음
- `Registry.wxs`에서 `WinMarkdownViewer.md` ProgID에 `DefaultIcon` 항목 없음 - .md 파일이 앱 아이콘으로 표시되지 않음

### 1.2 기술 스택

- Go 1.26+
- WiX Toolset v4 (NuGet)
- 빌드 스크립트: PowerShell (`build.ps1`, `installer/build-msi.ps1`)
- 시스템 트레이: `github.com/energye/systray` (Pure Go, CGO 불필요)

### 1.3 관련 파일

| 파일 | 역할 |
|------|------|
| `assets/icon.ico` | 아이콘 원본 파일 |
| `assets/embed.go` | go:embed로 IconData 제공 |
| `assets/embed_test.go` | ICO 파일 유효성 테스트 |
| `internal/tray/tray.go` | 시스템 트레이 아이콘 설정 |
| `build.ps1` | Go 빌드 스크립트 |
| `installer/wix/Package.wxs` | MSI 패키지 메인 (Icon 참조) |
| `installer/wix/Registry.wxs` | 레지스트리 (컨텍스트 메뉴, 파일 연결) |
| `installer/wix/Shortcuts.wxs` | 시작 메뉴 바로가기 (Icon 참조) |

## 2. 가정 (Assumptions)

- A1: goversioninfo 도구를 사용하여 `.syso` 파일을 생성할 수 있다 (CGO 불필요)
- A2: 멀티 사이즈 ICO 파일(16~256px)은 수동 또는 외부 도구로 제작한다
- A3: `.syso` 파일은 빌드 시 Go 컴파일러가 자동으로 링크한다 (`cmd/winmdview/` 디렉터리에 배치)
- A4: WiX는 exe에 임베딩된 리소스 아이콘을 정상적으로 추출할 수 있다
- A5: DefaultIcon 레지스트리 키에 exe 경로와 아이콘 인덱스를 지정하면 탐색기가 .md 파일에 해당 아이콘을 표시한다
- A6: 기존 `assets.IconData` (go:embed) 기반 시스템 트레이 로직은 그대로 유지한다

## 3. 요구사항 (Requirements)

### 3.1 멀티 사이즈 아이콘 (R1)

- **[UBIQUITOUS]** 시스템은 **항상** 다음 크기의 아이콘을 포함하는 ICO 파일을 사용해야 한다: 16x16, 32x32, 48x48, 64x64, 128x128, 256x256
- **[UBIQUITOUS]** 아이콘 디자인은 마크다운 뷰어의 정체성을 반영해야 한다 (마크다운 관련 시각 요소 포함)

### 3.2 EXE 리소스 임베딩 (R2)

- **[UBIQUITOUS]** 시스템은 **항상** Windows 리소스로 아이콘이 임베딩된 exe를 생성해야 한다
- **[EVENT]** WHEN `go build`가 실행될 때 THEN `.syso` 리소스 파일이 자동으로 exe에 링크되어야 한다
- **[EVENT]** WHEN 사용자가 파일 탐색기에서 exe를 볼 때 THEN 애플리케이션 아이콘이 표시되어야 한다

### 3.3 인스톨러 아이콘 통합 (R3)

- **[EVENT]** WHEN MSI 인스톨러가 설치를 완료하면 THEN 프로그램 추가/제거 목록에서 앱 아이콘이 표시되어야 한다
- **[EVENT]** WHEN 시작 메뉴 바로가기가 생성되면 THEN 앱 아이콘이 바로가기에 표시되어야 한다
- **[EVENT]** WHEN 컨텍스트 메뉴에 "마크다운 뷰어로 열기"가 표시될 때 THEN 해당 항목 옆에 앱 아이콘이 표시되어야 한다
  - 참고: 현재 `Registry.wxs`에 이미 `<RegistryValue Name="Icon" Value="[INSTALLFOLDER]$(var.ExeFileName)" />`이 존재하므로, exe에 리소스 아이콘이 임베딩되면 자동으로 충족됨

### 3.4 파일 연결 아이콘 (R4)

- **[EVENT]** WHEN 사용자가 파일 연결(FileAssociation) 기능을 활성화한 상태에서 .md 파일을 탐색기에서 볼 때 THEN .md 파일이 WinMarkdownViewer 아이콘으로 표시되어야 한다
- **[STATE]** IF 파일 연결이 활성화된 상태 AND DefaultIcon 레지스트리가 등록된 상태이면 THEN 탐색기가 .md 파일에 앱 아이콘을 표시해야 한다

### 3.5 시스템 트레이 호환성 (R5)

- **[UBIQUITOUS]** 시스템은 **항상** 개선된 멀티 사이즈 ICO 파일을 go:embed로 임베딩하여 시스템 트레이에 사용해야 한다
- **[UNWANTED]** 시스템은 시스템 트레이 아이콘 로직을 **변경하지 않아야 한다** (기존 `assets.IconData` → `systray.SetIcon()` 흐름 유지)

### 3.6 빌드 파이프라인 통합 (R6)

- **[EVENT]** WHEN `build.ps1 build` (릴리스 빌드)가 실행될 때 THEN goversioninfo가 `.syso` 파일을 생성한 후 go build가 실행되어야 한다
- **[EVENT]** WHEN `build.ps1 dev` (개발 빌드)가 실행될 때 THEN 동일하게 `.syso` 파일이 포함되어야 한다
- **[UNWANTED]** 빌드 프로세스에서 CGO 의존성이 **추가되지 않아야 한다**

## 4. 명세 (Specifications)

### 4.1 아이콘 파일 구조

```
assets/
  icon.ico          # 멀티 사이즈 ICO (16, 32, 48, 64, 128, 256)
  embed.go          # go:embed (기존 유지)

cmd/winmdview/
  versioninfo.json  # goversioninfo 매니페스트 (아이콘 + 버전 정보)
  resource.syso     # 빌드 산출물 (git에서 제외, 빌드 시 자동 생성)
```

### 4.2 goversioninfo 매니페스트 (`versioninfo.json`)

goversioninfo가 `versioninfo.json`을 읽어 `resource.syso`를 생성한다.

주요 설정:
- `IconPath`: `../../assets/icon.ico` (상대 경로)
- `FileDescription`: "WinMarkdownViewer"
- `ProductName`: "WinMarkdownViewer"
- `FileVersion` / `ProductVersion`: 빌드 시 지정

최소 샘플:
```json
{
  "FixedFileInfo": {
    "FileVersion": { "Major": 1, "Minor": 0, "Patch": 0, "Build": 0 },
    "ProductVersion": { "Major": 1, "Minor": 0, "Patch": 0, "Build": 0 },
    "FileType": "VFT_APP"
  },
  "StringFileInfo": {
    "FileDescription": "WinMarkdownViewer",
    "ProductName": "WinMarkdownViewer",
    "CompanyName": "WinMarkdownViewer",
    "LegalCopyright": "",
    "InternalName": "winmdview",
    "OriginalFilename": "winmdview.exe"
  },
  "IconPath": "../../assets/icon.ico"
}
```

### 4.3 빌드 파이프라인 변경

**현재:**
```
go build -ldflags="-s -w -H windowsgui" -o dist/winmdview.exe ./cmd/winmdview
```

**변경 후:**
```
# 1. goversioninfo로 .syso 생성
goversioninfo -o cmd/winmdview/resource.syso cmd/winmdview/versioninfo.json

# 2. go build (Go 컴파일러가 .syso 자동 링크)
go build -ldflags="-s -w -H windowsgui" -o dist/winmdview.exe ./cmd/winmdview
```

`build.ps1`의 `Invoke-Build` 및 `Invoke-Dev` 함수에 goversioninfo 실행 단계를 추가한다.

### 4.4 WiX 레지스트리 변경 (`Registry.wxs`)

**DefaultIcon 추가:**
```xml
<!-- ProgID에 DefaultIcon 추가 -->
<Component Id="ProgIdDefaultIconReg" Guid="*">
  <RegistryKey Root="HKCU"
               Key="Software\Classes\WinMarkdownViewer.md\DefaultIcon">
    <RegistryValue Type="string" Value="[INSTALLFOLDER]winmdview.exe,0" />
  </RegistryKey>
</Component>
```

해당 컴포넌트를 `FileAssociationComponents` 그룹에 추가한다.

### 4.5 WiX Package.wxs 변경 없음

exe에 아이콘 리소스가 임베딩되면 기존 `<Icon Id="AppIcon.exe" SourceFile="$(var.ExeFileName)" />`이 정상 동작한다. 변경 불필요. 단, `$(var.ExeFileName)`이 WiX 빌드 시 `-bindpath`를 통해 올바르게 resolve되는지 빌드 검증이 필요하다.

### 4.6 아이콘 파이프라인 (최종)

```
아이콘 디자인 (PNG/SVG 원본)
  --> 멀티 사이즈 ICO 변환 (16, 32, 48, 64, 128, 256)
  --> assets/icon.ico (업데이트)
  --> go:embed --> assets.IconData (런타임 트레이 아이콘)
  --> goversioninfo --> resource.syso (Windows 리소스)
  --> go build (syso 자동 링크) --> exe에 아이콘 임베딩
  --> WiX가 exe에서 아이콘 추출 --> MSI 인스톨러 아이콘
  --> Registry DefaultIcon --> .md 파일 연결 아이콘
```

## 5. 제약 조건 (Constraints)

- C1: CGO 사용 금지 - 순수 Go 빌드 체인 유지
- C2: 기존 `assets.IconData` 인터페이스 변경 금지 - 하위 호환성 유지
- C3: 기존 WiX 컴포넌트 GUID 변경 금지 - MSI 업그레이드 안정성 유지
- C4: `resource.syso`는 git 추적 대상에서 제외 (`.gitignore` 추가)
- C5: goversioninfo는 `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest`로 관리 (Go 1.24+ tool 디렉티브는 향후 마이그레이션 고려)

## 6. 추적성 (Traceability)

| TAG | 설명 | 관련 파일 |
|-----|------|-----------|
| SPEC-ICON-001-R1 | 멀티 사이즈 아이콘 | `assets/icon.ico` |
| SPEC-ICON-001-R2 | EXE 리소스 임베딩 | `cmd/winmdview/versioninfo.json`, `resource.syso` |
| SPEC-ICON-001-R3 | 인스톨러 아이콘 | `installer/wix/Package.wxs` |
| SPEC-ICON-001-R4 | 파일 연결 아이콘 | `installer/wix/Registry.wxs` |
| SPEC-ICON-001-R5 | 시스템 트레이 호환 | `assets/embed.go`, `internal/tray/tray.go` |
| SPEC-ICON-001-R6 | 빌드 파이프라인 | `build.ps1` |
