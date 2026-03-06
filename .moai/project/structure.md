# WinMarkdownViewer - 프로젝트 구조

## 디렉토리 구조

```
WinMarkdownViewer/
├── cmd/
│   └── winmdview/
│       └── main.go              # 애플리케이션 진입점
├── internal/
│   ├── app/
│   │   ├── app.go               # 애플리케이션 초기화 및 생명주기
│   │   └── instance.go          # 단일 인스턴스 관리 (Named Mutex/Pipe)
│   ├── markdown/
│   │   ├── renderer.go          # goldmark 기반 마크다운 렌더링
│   │   └── renderer_test.go
│   ├── server/
│   │   ├── server.go            # 내장 HTTP 서버
│   │   ├── websocket.go         # WebSocket 핸들러 (실시간 새로고침)
│   │   └── server_test.go
│   ├── watcher/
│   │   ├── watcher.go           # 파일 변경 감시 (fsnotify)
│   │   └── watcher_test.go
│   ├── viewer/
│   │   ├── viewer.go            # WebView2 창 관리
│   │   └── viewer_test.go
│   ├── registry/
│   │   ├── contextmenu.go       # 우클릭 컨텍스트 메뉴 등록
│   │   ├── fileassoc.go         # 파일 연결 등록
│   │   └── registry_test.go
│   ├── tray/
│   │   ├── tray.go              # 시스템 트레이 아이콘
│   │   └── tray_test.go
│   └── config/
│       ├── config.go            # 사용자 설정 관리
│       └── config_test.go
├── web/
│   ├── templates/
│   │   └── viewer.html          # 마크다운 뷰어 HTML 템플릿
│   ├── css/
│   │   ├── github-markdown.css  # GitHub 스타일 마크다운 CSS
│   │   ├── dark.css             # 다크 테마
│   │   └── light.css            # 라이트 테마
│   ├── js/
│   │   ├── websocket.js         # WebSocket 클라이언트 (실시간 새로고침)
│   │   └── scroll.js            # 스크롤 위치 유지
│   └── embed.go                 # go:embed 디렉티브
├── assets/
│   ├── icon.ico                 # 애플리케이션 아이콘
│   ├── icon.png                 # 시스템 트레이 아이콘
│   └── app.manifest             # Windows 매니페스트
├── installer/
│   ├── wix/
│   │   ├── Product.wxs          # WiX MSI 정의
│   │   └── Registry.wxs         # 레지스트리 설정
│   └── build-msi.ps1            # MSI 빌드 스크립트
├── docs/
│   ├── product.md               # 제품 정의서
│   ├── tech.md                  # 기술 스택
│   └── structure.md             # 프로젝트 구조 (이 파일)
├── go.mod
├── go.sum
├── Makefile                     # 빌드 명령어
├── README.md
└── .gitignore
```

## 모듈 의존성

```
cmd/winmdview/main.go
  ├── internal/app          # 앱 초기화
  │   ├── internal/viewer   # WebView2 창
  │   ├── internal/server   # HTTP + WebSocket 서버
  │   ├── internal/watcher  # 파일 감시
  │   ├── internal/tray     # 시스템 트레이
  │   └── internal/config   # 설정 관리
  ├── internal/markdown     # 마크다운 렌더링 (독립)
  └── internal/registry     # Windows 레지스트리 (독립)
```

## 데이터 흐름

```
1. 사용자가 .md 파일 우클릭 → "마크다운 뷰어로 열기"
2. winmdview.exe가 파일 경로를 인자로 실행
3. 단일 인스턴스 체크 (instance.go)
   ├── 새 인스턴스: 앱 초기화
   └── 기존 인스턴스: Named Pipe로 파일 경로 전달 → 기존 창에서 열기
4. 내장 HTTP 서버 시작 (localhost:랜덤포트)
5. goldmark으로 .md → HTML 변환
6. WebView2 창에서 localhost URL 로드
7. fsnotify로 파일 감시 시작
8. 파일 변경 시 → 재렌더링 → WebSocket으로 HTML 전송 → DOM 업데이트
```

## 빌드 명령어

```bash
# 개발 빌드
go build -o winmdview.exe ./cmd/winmdview

# 릴리스 빌드 (버전 정보 임베딩)
go build -ldflags "-s -w -H windowsgui -X main.version=1.0.0" -o winmdview.exe ./cmd/winmdview

# 테스트
go test -race ./...

# MSI 빌드
powershell -File installer/build-msi.ps1
```
