# WinMarkdownViewer

Windows용 Markdown 뷰어. Go로 작성되었으며 Microsoft Edge WebView2를 통해 GitHub 스타일로 Markdown 파일을 렌더링합니다.

## 기능

- **GFM 지원**: GitHub Flavored Markdown (테이블, 취소선, 자동 링크, 태스크 리스트)
- **구문 강조**: chroma 기반 코드 블록 구문 강조
- **GitHub 스타일**: 임베딩된 CSS로 네트워크 없이 GitHub과 동일한 스타일 적용
- **실시간 미리보기**: 파일 저장 시 자동으로 변경 내용 반영 (fsnotify + WebSocket)
- **스크롤 위치 유지**: 파일 변경 시 현재 스크롤 위치를 보존하여 끊김 없는 편집 경험 제공
- **사용자 설정**: JSON 기반 설정 시스템 (테마, 폰트 크기, 창 크기/위치 기억)
- **수학 수식**: KaTeX 기반 인라인($...$) 및 블록($$...$$) 수학 수식 렌더링
- **다이어그램**: Mermaid 기반 flowchart, sequence, class, state, gantt, pie 다이어그램 렌더링
- **오류 처리**: 파일 없음, 권한 오류, WebView2 미설치, 빈 파일 안내
- **보안**: CSP 헤더로 XSS 방지, localhost 전용 서버 바인딩
- **순수 Go**: CGO 없이 빌드 가능
- **Windows 컨텍스트 메뉴**: .md 파일 우클릭 시 "마크다운 뷰어로 열기" 메뉴 등록 지원
- **시스템 트레이**: 최소화 시 시스템 트레이로 이동, 트레이에서 복원/종료 지원
- **단일 인스턴스**: 이미 실행 중이면 새 파일을 기존 인스턴스로 전달 (Named Pipe)

## 사전 요구 사항

- **Go 1.26 이상**
- **Microsoft Edge WebView2 Runtime**

WebView2 Runtime이 설치되어 있지 않으면 설치 안내 메시지가 표시됩니다.
다운로드: https://developer.microsoft.com/microsoft-edge/webview2/

## 빌드

```
git clone https://github.com/AngeleyesTrue/WinMarkdownViewer
cd WinMarkdownViewer
go build -o winmdview.exe ./cmd/winmdview
```

콘솔 없이 실행되는 GUI 전용 빌드:

```
go build -ldflags="-H windowsgui" -o winmdview.exe ./cmd/winmdview
```

## 사용법

```
winmdview.exe <파일경로>
```

예시:

```
winmdview.exe README.md
winmdview.exe C:\문서\노트.md
```

인자 없이 실행하면 사용법 안내가 표시됩니다.

파일을 열면 내장 HTTP 서버가 localhost에서 시작되고, 외부 편집기에서 파일을 저장할 때마다 뷰어가 자동으로 업데이트됩니다.

## 사용자 설정

설정 파일 위치: `%APPDATA%\WinMarkdownViewer\config.json`

첫 실행 시 기본값으로 자동 생성됩니다.

| 항목 | 기본값 | 설명 |
|------|--------|------|
| theme | "system" | "light", "dark", "system" 중 선택 |
| fontSize | 16 | 폰트 크기 (14-24) |
| windowWidth | 1024 | 창 너비 |
| windowHeight | 768 | 창 높이 |
| windowX / windowY | -1 | 창 위치 (-1: 시스템 기본) |
| customCSS | "" | 사용자 정의 CSS 파일 경로 |
| lastOpenedFile | "" | 마지막으로 열었던 파일 |

창을 닫을 때 크기와 위치가 자동 저장되며, 다음 실행 시 복원됩니다.

## Windows 통합

### 컨텍스트 메뉴 등록

파일 탐색기에서 .md 파일을 우클릭하면 "마크다운 뷰어로 열기" 메뉴가 표시되도록 등록합니다.

```
winmdview.exe --register    # 컨텍스트 메뉴 등록
winmdview.exe --unregister  # 컨텍스트 메뉴 해제
```

레지스트리는 현재 사용자 범위(HKCU)에서만 수정하므로 관리자 권한이 필요하지 않습니다.

### Open With 프로그램 목록 등록

.md 파일의 "연결 프로그램" 목록에 WinMarkdownViewer를 추가합니다.

```
winmdview.exe --set-default  # Open With 목록에 등록
```

### 시스템 트레이

- 창을 최소화하면 태스크바에서 사라지고 시스템 트레이 아이콘으로 이동합니다.
- 트레이 아이콘 더블클릭: 창 복원
- 트레이 아이콘 우클릭: 컨텍스트 메뉴 (열기 / 종료)

### 단일 인스턴스

두 번째 실행 시 새 창을 열지 않고, Named Pipe를 통해 파일 경로를 기존 실행 중인 인스턴스로 전달합니다. 기존 인스턴스가 해당 파일로 전환하고 전면에 표시됩니다.

## 프로젝트 구조

```
cmd/winmdview/main.go              진입점, CLI 파싱 및 서버/감시 파이프라인
internal/app/app.go                파일 검증 및 렌더링 파이프라인
internal/config/config.go          사용자 설정 구조체 및 기본값
internal/config/loader.go          설정 파일 읽기/쓰기
internal/config/validator.go       설정값 검증 및 보정
internal/markdown/renderer.go      goldmark 렌더링 엔진
internal/server/server.go          내장 HTTP 서버 + WebSocket
internal/viewer/viewer.go          WebView2 윈도우 관리
internal/viewer/errors.go          오류 타입 정의
internal/watcher/watcher.go        fsnotify 기반 파일 변경 감시
internal/registry/registry.go      Windows 레지스트리 컨텍스트 메뉴 관리
internal/app/instance.go           Named Mutex 단일 인스턴스 관리
internal/app/pipe.go               Named Pipe 프로세스 간 통신
internal/tray/tray.go              시스템 트레이 관리
web/templates/viewer.html          HTML 템플릿 (WebSocket 클라이언트 포함)
web/css/github-markdown.css        GitHub Markdown CSS
web/js/render-extensions.js        KaTeX 수식 + Mermaid 다이어그램 렌더링
web/js/katex.min.js                KaTeX 렌더링 엔진 (go:embed)
web/js/mermaid.min.js              Mermaid 렌더링 엔진 (go:embed)
web/css/katex.min.css              KaTeX 수학 스타일시트
web/fonts/                         KaTeX 수학 폰트 (woff2)
web/embed.go                       go:embed 선언
assets/icon.ico                    트레이/컨텍스트 메뉴 아이콘
assets/embed.go                    아이콘 리소스 임베딩
```

## 기술 스택

| 구성 요소 | 라이브러리 |
|-----------|-----------|
| Markdown 파서 | github.com/yuin/goldmark |
| 구문 강조 | github.com/yuin/goldmark-highlighting/v2 |
| WebView2 바인딩 | github.com/jchv/go-webview2 |
| 파일 감시 | github.com/fsnotify/fsnotify |
| WebSocket | github.com/gorilla/websocket |
| 정적 리소스 | Go 표준 go:embed |
| Windows API | golang.org/x/sys/windows |
| 시스템 트레이 | github.com/energye/systray |
| 수학 수식 | KaTeX v0.16.x |
| 다이어그램 | Mermaid v10.x |

## 라이선스

MIT License. 자세한 내용은 LICENSE 파일을 참조하세요.
