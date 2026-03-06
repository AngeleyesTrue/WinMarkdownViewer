// Package viewer 는 WebView2 윈도우를 통해 렌더링된 HTML을 표시하는 기능을 제공한다.
package viewer

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/AngeleyesTrue/WinMarkdownViewer/web"
	"github.com/jchv/go-webview2"
)

// Hint 는 윈도우 크기 조정 힌트 타입이다. webview2.Hint와 동일하다.
type Hint = webview2.Hint

// WebView 는 WebView2 윈도우의 핵심 동작을 추상화하는 인터페이스이다.
// 테스트에서 모의 객체로 대체할 수 있다.
type WebView interface {
	SetTitle(title string)
	SetSize(w, h int, hint Hint)
	SetHtml(html string)
	Navigate(url string)
	Run()
	Destroy()
}

// Config 는 뷰어 윈도우의 설정을 정의한다.
type Config struct {
	Width  int // 윈도우 기본 너비
	Height int // 윈도우 기본 높이
}

// DefaultConfig 는 기본 뷰어 설정을 반환한다.
func DefaultConfig() Config {
	return Config{
		Width:  1024,
		Height: 768,
	}
}

// WebViewFactory 는 WebView 인스턴스를 생성하는 팩토리 함수 타입이다.
// 테스트에서 모의 객체를 주입할 수 있다.
type WebViewFactory func(cfg Config) (WebView, error)

// Viewer 는 WebView2 윈도우를 관리하는 구조체이다.
type Viewer struct {
	webview WebView
	config  Config
}

// defaultWebViewFactory 는 실제 WebView2 인스턴스를 생성하는 기본 팩토리이다.
func defaultWebViewFactory(cfg Config) (WebView, error) {
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "WinMarkdownViewer",
			Width:  uint(cfg.Width),
			Height: uint(cfg.Height),
			Center: true,
		},
	})
	if w == nil {
		return nil, ErrWebView2NotInstalled
	}
	return w, nil
}

// New 는 실제 WebView2 인스턴스를 사용하여 새 뷰어를 생성한다.
// WebView2 Runtime이 설치되지 않은 경우 에러를 반환한다.
func New(cfg Config) (*Viewer, error) {
	return NewWithFactory(cfg, defaultWebViewFactory)
}

// NewWithFactory 는 주어진 팩토리를 사용하여 새 뷰어를 생성한다.
// 테스트에서 모의 WebView를 주입할 수 있다.
func NewWithFactory(cfg Config, factory WebViewFactory) (*Viewer, error) {
	w, err := factory(cfg)
	if err != nil {
		return nil, err
	}
	return &Viewer{
		webview: w,
		config:  cfg,
	}, nil
}

// SetTitle 은 파일 경로에서 파일명을 추출하여 윈도우 타이틀을 설정한다.
// 형식: "filename.md - WinMarkdownViewer"
func (v *Viewer) SetTitle(filePath string) {
	filename := filepath.Base(filePath)
	v.webview.SetTitle(filename + " - WinMarkdownViewer")
}

// LoadHTML 은 렌더링된 HTML 콘텐츠를 WebView2에 로드한다.
func (v *Viewer) LoadHTML(html string) {
	v.webview.SetHtml(html)
}

// Run 은 WebView2 이벤트 루프를 시작한다.
// 윈도우가 닫힐 때까지 블로킹된다.
func (v *Viewer) Run() {
	v.webview.Run()
}

// Destroy 는 WebView2 리소스를 정리한다.
func (v *Viewer) Destroy() {
	v.webview.Destroy()
}

// templateData 는 HTML 템플릿에 전달할 데이터이다.
type templateData struct {
	Title   string
	CSS     template.CSS
	Content template.HTML
}

// viewerTmpl 은 모듈 초기화 시 파싱되는 캐싱된 HTML 템플릿이다.
// 매 호출마다 재파싱하지 않아 성능이 향상된다.
var viewerTmpl = template.Must(template.New("viewer").Parse(string(web.ViewerHTML)))

// BuildFullHTML 은 제목과 마크다운 렌더링 결과를 조합하여 완전한 HTML 문서를 생성한다.
// 임베디드 HTML 템플릿과 CSS를 사용한다.
//
// 보안 모델:
// - Title 필드는 html/template에 의해 자동 이스케이프된다.
// - Content 필드는 template.HTML 타입으로 goldmark이 생성한 안전한 HTML을 전달한다.
//   goldmark은 기본적으로 위험한 HTML을 필터링하며, WithUnsafe()는 원본 마크다운 내
//   HTML 블록(예: <details>, <summary>)을 허용하기 위함이다.
// - CSP 헤더가 viewer.html에 설정되어 스크립트 실행을 차단한다.
func BuildFullHTML(title string, renderedContent string) (string, error) {
	data := templateData{
		Title:   title,
		CSS:     template.CSS(web.GitHubMarkdownCSS),
		Content: template.HTML(renderedContent),
	}

	var buf bytes.Buffer
	if err := viewerTmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
