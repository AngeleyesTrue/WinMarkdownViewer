# WinMarkdownViewer - 프로젝트 구조

## 디렉토리 구조

```
WinMarkdownViewer/
├── build.ps1                              # PowerShell 빌드 스크립트 (릴리스/개발/테스트/클린)
├── cmd/
│   ├── winmdview/
│   │   └── main.go              # 애플리케이션 진입점, 서버/감시 파이프라인
│   └── poc/
│       └── multiwin/
│           └── main.go          # 멀티 WebView2 인스턴스 PoC 검증
├── internal/
│   ├── app/
│   │   ├── app.go               # 파일 검증 및 렌더링 파이프라인
│   │   ├── app_test.go
│   │   ├── constants.go         # 공통 상수 (Mutex 이름, Pipe 경로 등)
│   │   ├── instance.go          # Named Mutex 단일 인스턴스 관리
│   │   ├── instance_test.go
│   │   ├── pipe.go              # Named Pipe 프로세스 간 통신
│   │   └── pipe_test.go
│   ├── config/
│   │   ├── config.go            # 사용자 설정 구조체 및 기본값
│   │   ├── config_test.go
│   │   ├── loader.go            # 설정 파일 읽기/쓰기 (JSON)
│   │   ├── loader_test.go
│   │   ├── validator.go         # 설정값 검증 및 보정
│   │   └── validator_test.go
│   ├── markdown/
│   │   ├── renderer.go          # goldmark 기반 마크다운 렌더링 (GFM + 구문 강조)
│   │   └── renderer_test.go
│   ├── server/
│   │   ├── server.go            # 내장 HTTP 서버 + WebSocket 핸들러
│   │   └── server_test.go
│   ├── watcher/
│   │   ├── watcher.go           # 파일 변경 감시 (fsnotify + debounce)
│   │   └── watcher_test.go
│   ├── viewer/
│   │   ├── viewer.go            # WebView2 창 관리 (트레이 최소화/복원 지원)
│   │   ├── viewer_test.go
│   │   ├── viewer_ext_test.go   # WebView2 확장 테스트 (BuildFullHTML 등)
│   │   └── errors.go            # 오류 타입 정의
│   ├── registry/
│   │   ├── registry.go          # Windows 레지스트리 컨텍스트 메뉴 관리
│   │   └── registry_test.go
│   ├── tray/
│   │   ├── tray.go              # 시스템 트레이 관리
│   │   └── tray_test.go
│   └── window/
│       ├── manager.go           # 멀티 윈도우 중앙 관리자 (생성/추적/정리)
│       ├── manager_test.go
│       ├── window.go            # 개별 윈도우 상태 (서버, 감시자, 뷰어)
│       ├── window_test.go
│       └── errors.go            # 윈도우 관련 오류 타입 정의
├── assets/
│   ├── icon.ico                 # 트레이/컨텍스트 메뉴 아이콘 (16x16, 32x32)
│   └── embed.go                 # 아이콘 리소스 임베딩
├── web/
│   ├── templates/
│   │   └── viewer.html          # 마크다운 뷰어 HTML 템플릿 (WebSocket 클라이언트 포함)
│   ├── css/
│   │   ├── github-markdown.css  # GitHub 스타일 마크다운 CSS
│   │   └── katex.min.css        # KaTeX 수학 수식 스타일시트
│   ├── js/
│   │   ├── katex.min.js         # KaTeX 렌더링 엔진 (go:embed)
│   │   ├── mermaid.min.js       # Mermaid 다이어그램 렌더링 엔진 (go:embed)
│   │   └── render-extensions.js # KaTeX 수식 + Mermaid 다이어그램 렌더링 로직
│   ├── fonts/                   # KaTeX 수학 폰트 (woff2, 20개)
│   ├── embed.go                 # go:embed 디렉티브
│   └── embed_test.go            # 임베딩 검증 테스트
├── installer/
│   ├── wix/
│   │   ├── Package.wxs              # WiX v4 메인 패키지 정의 (UpgradeCode, MajorUpgrade)
│   │   ├── Directories.wxs          # 디렉토리 구조 정의
│   │   ├── Components.wxs           # 컴포넌트 및 파일 정의 (winmdview.exe)
│   │   ├── Registry.wxs             # 레지스트리 항목 정의 (컨텍스트 메뉴, 파일 연결)
│   │   ├── Shortcuts.wxs            # 시작 메뉴 바로가기 정의
│   │   └── Variables.wxi            # 공통 변수 (버전, 제품명 등)
│   ├── tests/
│   │   └── build-msi.Tests.ps1      # Pester 테스트
│   ├── build-msi.ps1                # MSI 빌드 스크립트
│   └── README.md                    # 빌드 방법 안내
├── .github/
│   └── workflows/
│       └── release.yml              # 릴리스 CI/CD 워크플로우 (v* 태그 → MSI 빌드 → Release)
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## 모듈 의존성

```
cmd/winmdview/main.go
  ├── internal/app          # 앱 초기화, 렌더링 파이프라인, 단일 인스턴스, Pipe
  ├── internal/window       # 멀티 윈도우 관리 (WindowManager)
  │   ├── internal/viewer   # WebView2 창
  │   ├── internal/server   # HTTP + WebSocket 서버
  │   └── internal/watcher  # 파일 감시
  ├── internal/config       # 설정 관리
  ├── internal/registry     # Windows 레지스트리 컨텍스트 메뉴
  ├── internal/tray         # 시스템 트레이 (동적 윈도우 목록)
  └── internal/markdown     # 마크다운 렌더링
```

## 데이터 흐름

```
1. 사용자가 .md 파일을 인자로 실행
1-1. Named Mutex로 기존 인스턴스 확인 → 이미 실행 중이면 Named Pipe로 OPEN:{path} 전송 후 종료
2. config.Load()로 사용자 설정 로드 (%APPDATA%\WinMarkdownViewer\config.json)
3. WindowManager가 윈도우별 독립 HTTP 서버 시작 (localhost:랜덤포트)
4. goldmark으로 .md -> HTML 변환 (GFM + chroma 구문 강조)
5. WebView2 창에서 http://localhost:{port} Navigate
6. fsnotify로 파일 감시 시작
7. 파일 변경 시 -> 재렌더링 -> WebSocket으로 HTML 전송 -> DOM 업데이트 (스크롤 위치 유지)
8. 클라이언트에서 KaTeX 수식 + Mermaid 다이어그램 후처리 렌더링
9. 창 닫기 시 -> 설정 자동 저장 -> graceful shutdown (서버/감시/WebSocket 정리)
```

## 빌드 명령어

```powershell
# PowerShell 빌드 스크립트 (권장)
.\build.ps1 build     # 릴리스 빌드 (-H windowsgui, 콘솔 숨김)
.\build.ps1 dev       # 개발 빌드 (콘솔 창 표시)
.\build.ps1 test      # 테스트 실행
.\build.ps1 clean     # 빌드 산출물 정리

# 수동 빌드
go build -o winmdview.exe ./cmd/winmdview
go build -ldflags="-s -w -H windowsgui" -o winmdview.exe ./cmd/winmdview
go test ./...
```
