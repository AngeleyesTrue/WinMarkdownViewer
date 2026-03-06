# WinMarkdownViewer

Windows용 Markdown 뷰어. Go로 작성되었으며 Microsoft Edge WebView2를 통해 GitHub 스타일로 Markdown 파일을 렌더링합니다.

## 기능

- **GFM 지원**: GitHub Flavored Markdown (테이블, 취소선, 자동 링크, 태스크 리스트)
- **구문 강조**: chroma 기반 코드 블록 구문 강조
- **GitHub 스타일**: 임베딩된 CSS로 네트워크 없이 GitHub과 동일한 스타일 적용
- **오류 처리**: 파일 없음, 권한 오류, WebView2 미설치, 빈 파일 안내
- **보안**: CSP 헤더로 XSS 방지, 외부 네트워크 요청 없음
- **순수 Go**: CGO 없이 빌드 가능

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

## 프로젝트 구조

```
cmd/winmdview/main.go              진입점, CLI 파싱 및 파이프라인 실행
internal/app/app.go                파일 검증 및 렌더링 파이프라인
internal/markdown/renderer.go      goldmark 렌더링 엔진
internal/viewer/viewer.go          WebView2 윈도우 관리
internal/viewer/errors.go          오류 타입 정의
web/templates/viewer.html          HTML 템플릿 (CSP 포함)
web/css/github-markdown.css        GitHub Markdown CSS
web/embed.go                       go:embed 선언
```

## 기술 스택

| 구성 요소 | 라이브러리 |
|-----------|-----------|
| Markdown 파서 | github.com/yuin/goldmark |
| 구문 강조 | github.com/yuin/goldmark-highlighting/v2 |
| WebView2 바인딩 | github.com/jchv/go-webview2 |
| 정적 리소스 | Go 표준 go:embed |

## 라이선스

MIT License. 자세한 내용은 LICENSE 파일을 참조하세요.
