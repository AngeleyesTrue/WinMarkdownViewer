---
id: SPEC-ICON-001
title: Application Icon & File Association Icon - Acceptance Criteria
version: 1.0.0
status: Planned
created: 2026-03-10
updated: 2026-03-10
author: Claud Archive
priority: High
---

# SPEC-ICON-001: 인수 기준

## 1. 멀티 사이즈 아이콘 [SPEC-ICON-001-R1]

### Scenario 1.1: ICO 파일이 6개 크기를 포함한다

```gherkin
Given assets/icon.ico 파일이 존재할 때
When ICO 파일의 ICONDIR 헤더를 파싱하면
Then 이미지 개수(Count)가 6이어야 한다
And 16x16, 32x32, 48x48, 64x64, 128x128, 256x256 크기가 모두 포함되어야 한다
```

### Scenario 1.2: ICO 파일이 유효한 형식이다

```gherkin
Given assets/icon.ico 파일이 존재할 때
When ICO 파일 헤더를 검증하면
Then Reserved 필드가 0이어야 한다
And Type 필드가 1 (icon)이어야 한다
And 각 이미지 엔트리의 오프셋과 크기가 유효해야 한다
```

### Scenario 1.3: go:embed 호환성

```gherkin
Given assets/icon.ico가 멀티 사이즈 ICO로 교체되었을 때
When go test ./assets/... 를 실행하면
Then TestIconDataNotEmpty가 통과해야 한다
And TestIconDataIsValidICO가 통과해야 한다
And TestIconDataMinimumSize가 통과해야 한다
```

## 2. EXE 리소스 임베딩 [SPEC-ICON-001-R2]

### Scenario 2.1: 빌드된 exe에 아이콘 리소스가 포함된다

```gherkin
Given goversioninfo가 설치되어 있고
And cmd/winmdview/versioninfo.json이 올바른 IconPath를 포함할 때
When build.ps1 build 를 실행하면
Then dist/winmdview.exe가 생성되어야 한다
And exe 파일에 RT_ICON 리소스가 포함되어야 한다
And exe 파일에 RT_GROUP_ICON 리소스가 포함되어야 한다
```

### Scenario 2.2: 탐색기에서 exe 아이콘이 표시된다

```gherkin
Given dist/winmdview.exe가 빌드되었을 때
When 파일 탐색기에서 해당 exe를 확인하면
Then 기본 Windows 아이콘이 아닌 WinMarkdownViewer 전용 아이콘이 표시되어야 한다
```

### Scenario 2.3: exe 파일 속성에 버전 정보가 표시된다

```gherkin
Given versioninfo.json에 버전 메타데이터가 설정되었을 때
When dist/winmdview.exe 파일의 속성 > 자세히 탭을 확인하면
Then 파일 설명에 "WinMarkdownViewer"가 표시되어야 한다
And 제품 이름에 "WinMarkdownViewer"가 표시되어야 한다
```

자동 검증 스크립트:
```powershell
$info = [System.Diagnostics.FileVersionInfo]::GetVersionInfo("dist/winmdview.exe")
if ($info.FileDescription -ne "WinMarkdownViewer") { throw "FileDescription mismatch" }
if ($info.ProductName -ne "WinMarkdownViewer") { throw "ProductName mismatch" }
```

## 3. 인스톨러 아이콘 [SPEC-ICON-001-R3]

### Scenario 3.1: 프로그램 추가/제거에서 아이콘이 표시된다

```gherkin
Given MSI 인스톨러로 WinMarkdownViewer를 설치했을 때
When Windows 설정 > 앱 > 설치된 앱 목록을 확인하면
Then WinMarkdownViewer 항목에 앱 아이콘이 표시되어야 한다
```

### Scenario 3.2: 시작 메뉴 바로가기에 아이콘이 표시된다

```gherkin
Given MSI 인스톨러로 WinMarkdownViewer를 설치했을 때
When 시작 메뉴에서 WinMarkdownViewer를 검색하면
Then 바로가기에 앱 아이콘이 표시되어야 한다
```

### Scenario 3.3: 컨텍스트 메뉴에 아이콘이 표시된다

```gherkin
Given MSI 인스톨러로 WinMarkdownViewer를 설치했을 때
When .md 파일을 우클릭하면
Then "마크다운 뷰어로 열기" 메뉴 항목 옆에 앱 아이콘이 표시되어야 한다
```

## 4. 파일 연결 아이콘 [SPEC-ICON-001-R4]

### Scenario 4.1: .md 파일에 앱 아이콘이 표시된다

```gherkin
Given MSI 설치 시 파일 연결(FileAssociation) 기능을 활성화했을 때
And WinMarkdownViewer.md ProgID가 기본 연결로 설정되었을 때
When 탐색기에서 .md 파일을 확인하면
Then .md 파일이 WinMarkdownViewer 아이콘으로 표시되어야 한다
```

### Scenario 4.2: DefaultIcon 레지스트리가 올바르게 설정된다

```gherkin
Given MSI 설치 시 파일 연결 기능이 활성화되었을 때
When HKCU\Software\Classes\WinMarkdownViewer.md\DefaultIcon 키를 확인하면
Then 기본값이 "[설치경로]\winmdview.exe,0" 형태이어야 한다
```

### Scenario 4.3: 파일 연결 비활성화 시 DefaultIcon이 없다

```gherkin
Given MSI 설치 시 파일 연결 기능을 비활성화했을 때
When HKCU\Software\Classes\WinMarkdownViewer.md\DefaultIcon 키를 확인하면
Then 해당 키가 존재하지 않아야 한다
```

## 5. 시스템 트레이 호환성 [SPEC-ICON-001-R5]

### Scenario 5.1: 시스템 트레이에 개선된 아이콘이 표시된다

```gherkin
Given 멀티 사이즈 ICO 파일이 go:embed로 임베딩되었을 때
When WinMarkdownViewer를 실행하면
Then 시스템 트레이에 앱 아이콘이 표시되어야 한다
And 아이콘이 흐릿하거나 깨지지 않아야 한다
```

### Scenario 5.2: 기존 트레이 API 호환성 유지

```gherkin
Given assets.IconData가 멀티 사이즈 ICO 데이터를 포함할 때
When internal/tray/tray.go의 NewTray(assets.IconData)가 호출되면
Then 에러 없이 트레이 아이콘이 설정되어야 한다
And systray.SetIcon()이 정상 동작해야 한다
```

## 6. 빌드 파이프라인 [SPEC-ICON-001-R6]

### Scenario 6.1: 릴리스 빌드에서 .syso가 자동 생성된다

```gherkin
Given goversioninfo가 시스템에 설치되어 있을 때
When build.ps1 build 를 실행하면
Then cmd/winmdview/resource.syso가 생성되어야 한다
And dist/winmdview.exe가 리소스 아이콘을 포함해야 한다
```

### Scenario 6.2: 개발 빌드에서도 .syso가 생성된다

```gherkin
Given goversioninfo가 시스템에 설치되어 있을 때
When build.ps1 dev 를 실행하면
Then cmd/winmdview/resource.syso가 생성되어야 한다
And dist/winmdview-dev.exe가 리소스 아이콘을 포함해야 한다
```

### Scenario 6.3: clean이 .syso를 삭제한다

```gherkin
Given cmd/winmdview/resource.syso가 존재할 때
When build.ps1 clean 을 실행하면
Then cmd/winmdview/resource.syso가 삭제되어야 한다
```

### Scenario 6.4: .syso가 git에서 추적되지 않는다

```gherkin
Given .gitignore에 *.syso가 포함되었을 때
When cmd/winmdview/resource.syso를 생성한 후
And git status를 실행하면
Then resource.syso가 untracked 파일 목록에 나타나지 않아야 한다
```

### Scenario 6.5: CGO 없이 빌드된다

```gherkin
Given CGO_ENABLED=0 환경에서
When build.ps1 build 를 실행하면
Then 빌드가 성공해야 한다
And exe에 리소스 아이콘이 포함되어야 한다
```

## 7. 기존 테스트 통과 [SPEC-ICON-001-R1, R5]

### Scenario 7.1: 전체 테스트 스위트 통과

```gherkin
Given assets/icon.ico가 멀티 사이즈 ICO로 교체되었을 때
When go test ./... 를 실행하면
Then 모든 기존 테스트가 통과해야 한다
And 새로 추가된 ICO 크기 검증 테스트도 통과해야 한다
```

## 8. 품질 게이트 (Quality Gate)

### Definition of Done

- [ ] `assets/icon.ico`가 6개 크기(16, 32, 48, 64, 128, 256)를 포함하는 멀티 사이즈 ICO 파일이다
- [ ] `cmd/winmdview/versioninfo.json`이 생성되어 올바른 아이콘 경로와 버전 정보를 포함한다
- [ ] `build.ps1`에 goversioninfo 실행 단계가 추가되었다
- [ ] `dist/winmdview.exe`가 Windows 리소스 아이콘을 포함한다
- [ ] `installer/wix/Registry.wxs`에 DefaultIcon 레지스트리 항목이 추가되었다
- [ ] `.gitignore`에 `*.syso`가 추가되었다
- [ ] `go test ./...`가 모든 테스트를 통과한다
- [ ] 탐색기에서 exe, .md 파일, 바로가기, 컨텍스트 메뉴에 아이콘이 표시된다
- [ ] 시스템 트레이 아이콘이 정상 동작한다
- [ ] CGO 의존성이 추가되지 않았다

### 검증 방법

| 항목 | 검증 방법 |
|------|-----------|
| ICO 파일 유효성 | `go test ./assets/...` (자동) |
| exe 리소스 존재 | ResourceHacker 또는 sigcheck로 확인 (수동) |
| exe 버전 정보 | `[System.Diagnostics.FileVersionInfo]::GetVersionInfo()` (자동) |
| 탐색기 아이콘 | 육안 확인 (수동) |
| .md 파일 아이콘 | MSI 설치 후 육안 확인 (수동) |
| 트레이 아이콘 | 앱 실행 후 육안 확인 (수동) |
| 빌드 성공 | `build.ps1 build` 실행 (자동) |
| 전체 테스트 | `go test ./...` (자동) |
