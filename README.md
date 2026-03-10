# WinMarkdownViewer

Windows용 Markdown 뷰어. Go로 작성되었으며 Microsoft Edge WebView2를 통해 GitHub 스타일로 Markdown 파일을 렌더링합니다.

## 기능

- **GFM 지원**: GitHub Flavored Markdown (테이블, 취소선, 자동 링크, 태스크 리스트)
- **구문 강조**: chroma 기반 코드 블록 구문 강조
- **GitHub 스타일**: 임베딩된 CSS로 네트워크 없이 GitHub과 동일한 스타일 적용
- **실시간 미리보기**: 파일 저장 시 자동으로 변경 내용 반영 (fsnotify + WebSocket)
- **스크롤 위치 유지**: 파일 변경 시 현재 스크롤 위치를 보존하여 끊김 없는 편집 경험 제공
- **다크 모드**: 라이트/다크/시스템 세 가지 테마 모드, 토글 버튼(우상단) 및 Ctrl+Shift+D 단축키
- **사용자 설정**: JSON 기반 설정 시스템 (테마, 폰트 크기, 창 크기/위치 기억)
- **수학 수식**: KaTeX 기반 인라인($...$) 및 블록($$...$$) 수학 수식 렌더링
- **다이어그램**: Mermaid 기반 flowchart, sequence, class, state, gantt, pie 다이어그램 렌더링
- **오류 처리**: 파일 없음, 권한 오류, WebView2 미설치, 빈 파일 안내
- **보안**: CSP 헤더로 XSS 방지, localhost 전용 서버 바인딩
- **순수 Go**: CGO 없이 빌드 가능
- **Windows 컨텍스트 메뉴**: .md 파일 우클릭 시 "마크다운 뷰어로 열기" 메뉴 등록 지원
- **시스템 트레이**: 최소화 시 시스템 트레이로 이동, 트레이에서 복원/종료 지원
- **단일 인스턴스**: 이미 실행 중이면 새 파일을 기존 인스턴스의 새 윈도우로 전달 (Named Pipe)
- **멀티 윈도우**: 여러 .md 파일을 각각 별도 윈도우에서 동시에 열기 가능 (최대 10개)

## 설치

### MSI 인스톨러 (권장)

[Releases](https://github.com/AngeleyesTrue/WinMarkdownViewer/releases) 페이지에서 `WinMarkdownViewer-x.x.x-x64.msi` 파일을 다운로드하여 실행합니다.

설치 시 자동으로 등록되는 항목:
- 파일 탐색기 .md 파일 우클릭 컨텍스트 메뉴 ("마크다운 뷰어로 열기")
- 시작 메뉴 바로가기
- 프로그램 추가/제거 등록
- (선택) .md 파일 연결 (Open With 목록)

제거 시 위 항목이 모두 자동으로 정리됩니다. 사용자 설정(`%APPDATA%\WinMarkdownViewer\`)은 보존됩니다.

### MSI 빌드 (개발자)

```powershell
.\installer\build-msi.ps1 -Version "1.0.0"
```

빌드 요구 사항: Go 1.26+, .NET SDK 6+, WiX Toolset v4

## 사전 요구 사항

- **Go 1.26 이상**
- **Microsoft Edge WebView2 Runtime**

WebView2 Runtime이 설치되어 있지 않으면 설치 안내 메시지가 표시됩니다.
다운로드: https://developer.microsoft.com/microsoft-edge/webview2/

## 빌드

```
git clone https://github.com/AngeleyesTrue/WinMarkdownViewer
cd WinMarkdownViewer
```

PowerShell 빌드 스크립트 (권장):

```powershell
.\build.ps1 build     # 릴리스 빌드 (-H windowsgui, 콘솔 숨김)
.\build.ps1 dev       # 개발 빌드 (콘솔 창 표시)
.\build.ps1 test      # 테스트 실행
.\build.ps1 clean     # 빌드 산출물 정리
```

수동 빌드:

```
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

### 테마 설정

WinMarkdownViewer는 세 가지 테마 모드를 지원합니다.

| 모드 | 설명 |
|------|------|
| `system` | OS의 다크/라이트 설정을 자동으로 따름 (기본값) |
| `light` | 라이트 테마 고정 |
| `dark` | 다크 테마 고정 |

**테마 전환 방법**

- **토글 버튼**: 뷰어 우상단의 버튼을 클릭하면 system → light → dark 순으로 순환합니다.
- **키보드 단축키**: `Ctrl+Shift+D`로 테마를 빠르게 전환할 수 있습니다.

**시스템 테마 자동 감지**

`system` 모드에서는 `prefers-color-scheme` 미디어 쿼리를 통해 OS 설정 변경을 실시간으로 감지하여 자동으로 전환됩니다. Windows 설정에서 다크 모드를 켜거나 끄면 뷰어도 즉시 반영됩니다.

**테마 유지**

선택한 테마는 `localStorage` 및 WebSocket을 통해 Go 서버로 전달되어 `config.json`에 저장됩니다. 프로그램을 재시작해도 마지막으로 선택한 테마가 유지됩니다.

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
- 트레이 아이콘 우클릭: 컨텍스트 메뉴 (열린 윈도우 목록 / 종료)

### 단일 인스턴스

두 번째 실행 시 새 창을 열지 않고, Named Pipe를 통해 파일 경로를 기존 실행 중인 인스턴스로 전달합니다. 기존 인스턴스가 해당 파일을 새 윈도우에서 열고 전면에 표시됩니다. 이미 열려 있는 파일이면 해당 윈도우를 활성화합니다.

## 프로젝트 구조

```
build.ps1                              PowerShell 빌드 스크립트 (릴리스/개발/테스트/클린)
cmd/winmdview/main.go              진입점, CLI 파싱 및 서버/감시 파이프라인
cmd/poc/multiwin/main.go           멀티 WebView2 인스턴스 PoC 검증
internal/app/app.go                파일 검증 및 렌더링 파이프라인
internal/config/config.go          사용자 설정 구조체 및 기본값
internal/config/loader.go          설정 파일 읽기/쓰기
internal/config/validator.go       설정값 검증 및 보정
internal/markdown/renderer.go      goldmark 렌더링 엔진
internal/server/server.go          내장 HTTP 서버 + WebSocket
internal/viewer/viewer.go          WebView2 윈도우 관리
internal/viewer/errors.go          오류 타입 정의
internal/watcher/watcher.go        fsnotify 기반 파일 변경 감시
internal/window/manager.go         멀티 윈도우 중앙 관리자 (생성/추적/정리)
internal/window/window.go          개별 윈도우 상태 (서버, 감시자, 뷰어)
internal/window/errors.go          윈도우 관련 오류 타입 정의
internal/registry/registry.go      Windows 레지스트리 컨텍스트 메뉴 관리
internal/app/instance.go           Named Mutex 단일 인스턴스 관리
internal/app/pipe.go               Named Pipe 프로세스 간 통신
internal/tray/tray.go              시스템 트레이 관리
web/templates/viewer.html          HTML 템플릿 (WebSocket 클라이언트 + 테마 토글 버튼)
web/css/github-markdown.css        GitHub Markdown CSS (CSS 변수 기반 테마 대응)
web/css/theme-light.css            라이트 테마 CSS 변수 정의
web/css/theme-dark.css             다크 테마 CSS 변수 정의
web/css/syntax-light.css           코드 구문 강조 라이트 테마
web/css/syntax-dark.css            코드 구문 강조 다크 테마
web/js/theme.js                    테마 전환 로직 (시스템 감지, localStorage, WebSocket 저장)
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
