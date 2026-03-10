# WinMarkdownViewer - 제품 정의서

## 개요

Windows 환경에서 마크다운(.md) 파일을 빠르고 편리하게 미리보기할 수 있는 경량 뷰어 애플리케이션.
파일 탐색기에서 우클릭 컨텍스트 메뉴 또는 파일 더블클릭으로 즉시 렌더링된 마크다운을 확인할 수 있다.

## 핵심 가치

- **즉시성**: 파일 우클릭 또는 실행 시 0.5초 이내에 렌더링된 결과 표시
- **실시간성**: 외부 편집기에서 파일 수정 시 자동으로 새로고침
- **경량성**: 단일 프로세스, 최소 리소스 사용 (윈도우당 약 50MB)
- **네이티브**: Windows 내장 WebView2 엔진 활용, 추가 브라우저 불필요

## 대상 사용자

- 마크다운 문서를 자주 작성하는 개발자
- README, 기술 문서를 검토하는 팀원
- 마크다운 기반 노트를 사용하는 사용자

## 핵심 기능

### F1: 마크다운 렌더링 (SPEC-UI-001, SPEC-RENDER-001 구현 완료)
- GFM(GitHub Flavored Markdown) 지원
- 코드 블록 구문 강조 (chroma 기반 Syntax Highlighting)
- 테이블, 체크박스, 각주 지원
- 수학 수식 렌더링 (KaTeX, 인라인/블록 지원)
- Mermaid 다이어그램 지원 (flowchart, sequence, class, state, gantt, pie)

### F2: 실시간 미리보기 (SPEC-WATCH-001 구현 완료)
- 파일 시스템 감시 (fsnotify)를 통한 변경 감지
- WebSocket 기반 자동 새로고침
- 스크롤 위치 유지

### F3: Windows 통합 (SPEC-WIN-001 구현 완료)
- 파일 탐색기 우클릭 컨텍스트 메뉴 ("마크다운 뷰어로 열기") - `--register` / `--unregister`
- .md 파일 Open With 프로그램 목록 등록 - `--set-default`
- 시스템 트레이 아이콘 (최소화 시, 더블클릭 복원, 우클릭 메뉴)
- 단일 인스턴스 실행 (Named Mutex + Named Pipe로 기존 인스턴스에 파일 전달)
- HKCU 범위 레지스트리 사용 (관리자 권한 불필요)

### F4: MSI 설치 프로그램 (SPEC-INSTALL-001 구현 완료)
- WiX Toolset v4 기반 MSI 빌드
- 컨텍스트 메뉴 자동 등록/해제
- 파일 연결 자동 설정 (선택적 Feature)
- 프로그램 추가/제거에서 깔끔한 제거
- 시작 메뉴 바로가기 생성

#### 업그레이드 전략 (Major Upgrade 패턴)
- **UpgradeCode (고정 GUID)**: 제품 식별용 고정 GUID, 절대 변경 불가
- **새 버전 설치**: 기존 버전 자동 제거 → 새 버전 설치 (MajorUpgrade)
- **다운그레이드 방지**: WiX MajorUpgrade 설정으로 자동 차단
- **동일 버전 재설치**: WiX 기본 Repair 동작에 의존
- **설정 파일 보존**: %APPDATA%\WinMarkdownViewer\는 MSI 관리 범위 외 → 업그레이드 시 자동 보존
- **컴포넌트 GUID 유지**: 파일별 컴포넌트 GUID 변경 금지 (잔여 파일 방지)

#### 배포 전략
- GitHub Actions에서 `v*` 태그 푸시 시 자동 빌드 + Release 첨부
- 빌드 스크립트: `installer/build-msi.ps1` (Go 빌드 + WiX 컴파일 + MSI 생성)
- 결과물: `WinMarkdownViewer-{version}-x64.msi`

#### 범위 외 (v1.0)
- 자동 업데이트 (WinSparkle 등)
- Chocolatey/Scoop/winget 패키지 매니저 배포
- 포터블(ZIP) 배포판
- ARM64 아키텍처 지원

### F6: 멀티 윈도우 (SPEC-MULTIWIN-001 구현 완료)
- 여러 .md 파일을 각각 독립된 윈도우에서 동시 열기
- 윈도우별 독립 HTTP 서버, 파일 감시자, WebSocket 연결
- 동일 파일 중복 열기 방지 (기존 윈도우 활성화)
- 최대 10개 윈도우 제한 (경고 및 초과 거부)
- 시스템 트레이에 열린 윈도우 동적 목록 표시
- Named Pipe OPEN: 프로토콜로 기존 인스턴스에 새 파일 전달
- PowerShell 빌드 스크립트 (build.ps1): 릴리스/개발/테스트/클린 빌드

### F5: 사용자 설정 (SPEC-CONFIG-001 구현 완료)
- 테마 선택 (라이트/다크/시스템 연동)
- CSS 커스터마이징
- 폰트 크기 조절
- 창 크기/위치 기억

## 비기능 요구사항

| 항목 | 목표 |
|------|------|
| 시작 시간 | 0.5초 이내 |
| 메모리 사용 | 50MB 이하 (일반 문서) |
| 바이너리 크기 | 15MB 이하 |
| 지원 OS | Windows 10 21H2 이상, Windows 11 |
| WebView2 | Edge WebView2 Runtime 필요 (Win11 기본 포함) |

## 제외 범위 (v1.0)

- 마크다운 편집 기능
- 파일 내보내기 (PDF, HTML)
- 플러그인 시스템
- macOS/Linux 지원
