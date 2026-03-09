# WinMarkdownViewer - 프로젝트 구조

## 디렉토리 구조

```
WinMarkdownViewer/
├── cmd/
│   └── winmdview/
│       └── main.go              # 애플리케이션 진입점, 서버/감시 파이프라인
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
│   └── tray/
│       ├── tray.go              # 시스템 트레이 관리
│       └── tray_test.go
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
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## 모듈 의존성

```
cmd/winmdview/main.go
  ├── internal/app          # 앱 초기화, 렌더링 파이프라인, 단일 인스턴스, Pipe
  │   ├── internal/viewer   # WebView2 창 (트레이 최소화/복원 지원)
  │   ├── internal/server   # HTTP + WebSocket 서버 (정적 파일 서빙 포함)
  │   ├── internal/watcher  # 파일 감시
  │   └── internal/config   # 설정 관리
  ├── internal/registry     # Windows 레지스트리 컨텍스트 메뉴
  ├── internal/tray         # 시스템 트레이
  └── internal/markdown     # 마크다운 렌더링 (독립, goldmark + chroma)
```

## 데이터 흐름

```
1. 사용자가 .md 파일을 인자로 실행
2. config.Load()로 사용자 설정 로드 (%APPDATA%\WinMarkdownViewer\config.json)
3. 내장 HTTP 서버 시작 (localhost:랜덤포트)
4. goldmark으로 .md -> HTML 변환 (GFM + chroma 구문 강조)
5. WebView2 창에서 http://localhost:{port} Navigate
6. fsnotify로 파일 감시 시작
7. 파일 변경 시 -> 재렌더링 -> WebSocket으로 HTML 전송 -> DOM 업데이트 (스크롤 위치 유지)
8. 클라이언트에서 KaTeX 수식 + Mermaid 다이어그램 후처리 렌더링
9. 창 닫기 시 -> 설정 자동 저장 -> graceful shutdown (서버/감시/WebSocket 정리)
```

## 빌드 명령어

```bash
# 개발 빌드
go build -o winmdview.exe ./cmd/winmdview

# 릴리스 빌드 (콘솔 숨김)
go build -ldflags="-s -w -H windowsgui" -o winmdview.exe ./cmd/winmdview

# 테스트
go test ./...

# 테스트 (커버리지)
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
```
