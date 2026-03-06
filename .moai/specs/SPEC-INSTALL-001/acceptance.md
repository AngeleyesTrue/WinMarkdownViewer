---
spec-id: SPEC-INSTALL-001
title: "MSI Installer - Acceptance Criteria"
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-INSTALL-001: MSI 설치 프로그램 - 수용 기준

## 1. 설치 시나리오

### AC-INST-001: MSI 설치로 실행 파일 배치

```gherkin
Given Windows 10/11 x64 시스템에 WinMarkdownViewer가 설치되어 있지 않을 때
When 사용자가 WinMarkdownViewer-{version}-x64.msi를 더블클릭하고 설치를 완료하면
Then winmdview.exe가 "%ProgramFiles%\WinMarkdownViewer\" 디렉토리에 존재해야 한다
And winmdview.exe가 실행 가능한 상태여야 한다
```

### AC-INST-002: 컨텍스트 메뉴 및 시작 메뉴 자동 등록

```gherkin
Given MSI 설치가 완료된 상태일 때
When 파일 탐색기에서 .md 파일을 우클릭하면
Then "마크다운 뷰어로 열기" 메뉴 항목이 표시되어야 한다

Given MSI 설치가 완료된 상태일 때
When 시작 메뉴에서 "WinMarkdownViewer"를 검색하면
Then WinMarkdownViewer 바로가기가 검색 결과에 나타나야 한다
```

### AC-INST-003: 선택적 파일 연결 등록

```gherkin
Given MSI 설치 중 ".md 파일 연결" 옵션이 선택된 상태일 때
When 설치가 완료되면
Then .md 파일 더블클릭 시 WinMarkdownViewer가 실행되어야 한다
And HKCU\Software\Classes\.md\OpenWithProgids에 WinMarkdownViewer.md 항목이 존재해야 한다

Given MSI 설치 중 ".md 파일 연결" 옵션이 선택되지 않은 상태일 때(기본값)
When 설치가 완료되면
Then .md 파일의 기존 파일 연결이 변경되지 않아야 한다
```

### AC-INST-004: 프로그램 추가/제거 표시

```gherkin
Given MSI 설치가 완료된 상태일 때
When 설정 > 앱 > 앱 및 기능에서 "WinMarkdownViewer"를 검색하면
Then "WinMarkdownViewer" 항목이 표시되어야 한다
And 버전 정보가 올바르게 표시되어야 한다
And "제거" 버튼이 활성화되어야 한다
```

---

## 2. 제거 시나리오

### AC-UNINST-001: 깔끔한 제거

```gherkin
Given WinMarkdownViewer가 MSI를 통해 설치된 상태일 때
When 사용자가 프로그램 추가/제거에서 "WinMarkdownViewer"를 제거하면
Then "%ProgramFiles%\WinMarkdownViewer\" 디렉토리가 삭제되어야 한다
And .md 파일 우클릭 컨텍스트 메뉴에서 "마크다운 뷰어로 열기" 항목이 사라져야 한다
And 시작 메뉴의 WinMarkdownViewer 바로가기가 삭제되어야 한다
And HKCU\Software\Classes\.md\shell\WinMarkdownViewer 레지스트리 키가 삭제되어야 한다
```

### AC-UNINST-002: 사용자 설정 보존

```gherkin
Given WinMarkdownViewer가 설치되어 있고 "%APPDATA%\WinMarkdownViewer\config.json"이 존재할 때
When 사용자가 WinMarkdownViewer를 제거하면
Then "%APPDATA%\WinMarkdownViewer\" 디렉토리와 그 내용이 보존되어야 한다
```

---

## 3. 업그레이드 시나리오

### AC-UPGRADE-001: Major Upgrade

```gherkin
Given WinMarkdownViewer v1.0.0이 설치된 상태일 때
When 사용자가 WinMarkdownViewer v1.1.0 MSI를 실행하면
Then 기존 v1.0.0이 자동으로 제거되어야 한다
And v1.1.0이 설치되어야 한다
And 레지스트리 항목이 유지되어야 한다
And 사용자 설정 파일(%APPDATA%)이 보존되어야 한다
```

### AC-UPGRADE-002: 동일 버전 재설치 (수리)

```gherkin
Given WinMarkdownViewer v1.0.0이 설치된 상태일 때
When 사용자가 동일 버전(v1.0.0) MSI를 실행하면
Then WiX 기본 Repair 동작이 제공되어야 한다 (커스텀 수리 UI는 구현하지 않음)
And 수리 실행 시 손상된 파일이나 레지스트리가 복원되어야 한다
```

---

## 4. 빌드 시나리오

### AC-BUILD-001: 로컬 MSI 빌드

```gherkin
Given Go 1.26.0과 .NET SDK가 설치된 개발 환경에서
When `.\installer\build-msi.ps1 -Version "1.0.0"` 명령을 실행하면
Then "dist/WinMarkdownViewer-1.0.0-x64.msi" 파일이 생성되어야 한다
And MSI 파일 크기가 0보다 커야 한다
And 빌드 과정에서 오류가 발생하지 않아야 한다
```

### AC-BUILD-002: GitHub Actions 릴리스

```gherkin
Given GitHub 리포지토리에 release.yml 워크플로우가 설정된 상태에서
When "v1.0.0" 태그가 푸시되면
Then GitHub Actions 워크플로우가 자동으로 실행되어야 한다
And Go 바이너리 빌드가 성공해야 한다
And MSI 패키지 생성이 성공해야 한다
And "v1.0.0" GitHub Release가 생성되어야 한다
And MSI 파일이 Release 에셋으로 첨부되어야 한다
```

---

## 5. 레지스트리 시나리오

### AC-REG-001: 필수 레지스트리 항목 등록

```gherkin
Given MSI 설치가 완료된 상태일 때
When 레지스트리 편집기에서 다음 키를 확인하면
Then HKCU\Software\Classes\.md\shell\WinMarkdownViewer 키가 존재해야 한다
And HKCU\Software\Classes\.md\shell\WinMarkdownViewer\command 키의 기본값이 winmdview.exe 경로를 포함해야 한다
```

### AC-REG-002: 선택적 파일 연결 레지스트리

```gherkin
Given MSI 설치 중 ".md 파일 연결" 옵션을 선택했을 때
When 레지스트리 편집기에서 확인하면
Then HKCU\Software\Classes\.md\OpenWithProgids 키에 "WinMarkdownViewer.md" 값이 존재해야 한다
And HKCU\Software\Classes\WinMarkdownViewer.md 키가 존재해야 한다
And 해당 키의 shell\open\command 값이 winmdview.exe 경로를 포함해야 한다
```

---

## 6. Edge Case 시나리오

### EC-001: 레지스트리 등록 시 관리자 권한 불필요

```gherkin
Given 비관리자 계정으로 로그인한 상태에서
When WinMarkdownViewer MSI를 더블클릭하면
Then 레지스트리 등록은 HKCU 범위에서 수행되므로 관리자 권한이 불필요하다
  And %ProgramFiles% 설치를 위해 UAC 프롬프트가 필요할 수 있다
  And 사용자 디렉토리에 설치하는 경우 관리자 권한 없이 설치가 가능해야 한다
```

### EC-002: 설치 경로에 공백/한글 포함

```gherkin
Given 사용자가 설치 경로를 "C:\Program Files\마크다운 뷰어\"로 변경했을 때
When 설치를 완료하면
Then winmdview.exe가 해당 경로에 정상 설치되어야 한다
And 레지스트리의 command 값에 경로가 따옴표로 감싸져야 한다
```

### EC-003: WebView2 Runtime 미설치 상태

```gherkin
Given WebView2 Runtime이 설치되지 않은 시스템에서
When MSI 설치를 시도하면
Then MSI 설치 시 WebView2 Runtime 존재 여부를 Launch Condition으로 확인해야 한다
  And 미설치 시 설치 안내 메시지를 표시하고 Evergreen Bootstrapper 다운로드 URL(https://go.microsoft.com/fwlink/p/?LinkId=2124703)을 제공해야 한다
```

---

## 7. Quality Gates

### 7.1 빌드 품질
- [ ] `build-msi.ps1`이 오류 없이 MSI를 생성한다
- [ ] 생성된 MSI가 Windows Installer 표준을 준수한다
- [ ] GitHub Actions 워크플로우가 `v*` 태그에서 정상 실행된다

### 7.2 설치 품질
- [ ] 설치 후 winmdview.exe가 실행 가능하다
- [ ] 컨텍스트 메뉴가 정상 등록된다
- [ ] 시작 메뉴 바로가기가 동작한다
- [ ] 프로그램 추가/제거에 표시된다

### 7.3 제거 품질
- [ ] 모든 설치 파일이 삭제된다
- [ ] 모든 레지스트리 항목이 제거된다
- [ ] 사용자 설정(%APPDATA%)이 보존된다

### 7.4 Definition of Done
- [ ] 모든 AC(수용 기준) 시나리오 통과
- [ ] WiX v4 빌드가 로컬 및 CI 환경에서 성공
- [ ] MSI 설치/제거/업그레이드 수동 검증 완료
- [ ] Edge case 시나리오 검증 완료
- [ ] 코드 리뷰 완료
