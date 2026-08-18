package viewer

import (
	"strings"
	"testing"
)

// mockWebView 는 WebView 인터페이스를 테스트용으로 구현하는 모의 객체이다.
type mockWebView struct {
	title      string
	html       string
	url        string
	width      int
	height     int
	resizable  bool
	runCalled  bool
	destroyed  bool
}

func (m *mockWebView) SetTitle(title string)    { m.title = title }
func (m *mockWebView) SetSize(w, h int, hint Hint) {
	m.width = w
	m.height = h
}
func (m *mockWebView) SetHtml(html string)      { m.html = html }
func (m *mockWebView) Navigate(url string)       { m.url = url }
func (m *mockWebView) Run()                      { m.runCalled = true }
func (m *mockWebView) Destroy()                  { m.destroyed = true }

// TestNewViewerConfig 뷰어 생성 시 기본 설정을 검증한다.
func TestNewViewerConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Width != 1024 {
		t.Errorf("기본 너비: want 1024, got %d", cfg.Width)
	}
	if cfg.Height != 768 {
		t.Errorf("기본 높이: want 768, got %d", cfg.Height)
	}
}

// TestNewWithFactorySuccess 팩토리를 사용한 뷰어 생성을 검증한다.
func TestNewWithFactorySuccess(t *testing.T) {
	mockFactory := func(cfg Config) (WebView, error) {
		return &mockWebView{width: cfg.Width, height: cfg.Height}, nil
	}

	cfg := DefaultConfig()
	v, err := NewWithFactory(cfg, mockFactory)
	if err != nil {
		t.Fatalf("NewWithFactory() 오류: %v", err)
	}
	if v == nil {
		t.Fatal("NewWithFactory()가 nil을 반환함")
	}
}

// TestNewWithFactoryError 팩토리 에러 시 뷰어 생성 실패를 검증한다 (REQ-S-003).
func TestNewWithFactoryError(t *testing.T) {
	mockFactory := func(cfg Config) (WebView, error) {
		return nil, ErrWebView2NotInstalled
	}

	cfg := DefaultConfig()
	v, err := NewWithFactory(cfg, mockFactory)
	if err == nil {
		t.Fatal("WebView2 미설치 시 에러가 반환되어야 한다")
	}
	if v != nil {
		t.Fatal("에러 발생 시 뷰어가 nil이어야 한다")
	}
	if !strings.Contains(err.Error(), "WMV-E001") {
		t.Errorf("에러 메시지에 에러 코드(WMV-E001)가 포함되어야 함: %v", err)
	}
	if !strings.Contains(err.Error(), "WebView2") {
		t.Errorf("에러 메시지에 WebView2 설치 안내가 포함되어야 함: %v", err)
	}
}

// TestViewerSetTitle 윈도우 타이틀 설정을 검증한다.
func TestViewerSetTitle(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	v.SetTitle("test.md")

	expected := "test.md - WinMarkdownViewer"
	if mock.title != expected {
		t.Errorf("타이틀: want %q, got %q", expected, mock.title)
	}
}

// TestViewerSetTitleWithPath 경로가 포함된 파일명에서 파일명만 추출하여 타이틀을 설정하는지 검증한다.
func TestViewerSetTitleWithPath(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	v.SetTitle(`C:\Users\test\docs\README.md`)

	expected := "README.md - WinMarkdownViewer"
	if mock.title != expected {
		t.Errorf("타이틀: want %q, got %q", expected, mock.title)
	}
}

// TestViewerLoadHTML HTML 콘텐츠 로딩을 검증한다.
func TestViewerLoadHTML(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	htmlContent := "<html><body><h1>Hello</h1></body></html>"
	v.LoadHTML(htmlContent)

	if mock.html != htmlContent {
		t.Errorf("HTML 콘텐츠가 올바르게 로드되지 않음.\nwant: %s\ngot: %s", htmlContent, mock.html)
	}
}

// TestViewerDestroy 뷰어 정리를 검증한다.
func TestViewerDestroy(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	v.Destroy()

	if !mock.destroyed {
		t.Error("Destroy()가 호출되지 않음")
	}
}

// TestBuildFullHTML 전체 HTML 문서 생성을 검증한다.
func TestBuildFullHTML(t *testing.T) {
	result, err := BuildFullHTML("Test Title", "<p>Hello World</p>")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}

	checks := []string{
		"Test Title",
		"<p>Hello World</p>",
		"markdown-body",
		"<!DOCTYPE html>",
	}

	for _, want := range checks {
		if !strings.Contains(result, want) {
			t.Errorf("BuildFullHTML 결과에 %q 가 포함되지 않음", want)
		}
	}
}

// TestViewerRun 이벤트 루프 시작을 검증한다.
func TestViewerRun(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	v.Run()

	if !mock.runCalled {
		t.Error("Run()이 호출되지 않음")
	}
}

// TestBuildFullHTMLEmptyContent 빈 콘텐츠로 HTML 생성을 검증한다.
func TestBuildFullHTMLEmptyContent(t *testing.T) {
	result, err := BuildFullHTML("Empty", "")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("빈 콘텐츠에서도 HTML 문서 구조가 생성되어야 한다")
	}
}

// TestBuildFullHTMLSpecialCharactersInTitle 타이틀에 특수문자가 있을 때 이스케이프를 검증한다.
func TestBuildFullHTMLSpecialCharactersInTitle(t *testing.T) {
	result, err := BuildFullHTML("test<script>alert(1)</script>.md", "<p>safe</p>")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}
	// 타이틀의 HTML 특수문자가 이스케이프되어야 한다
	if strings.Contains(result, "<script>alert(1)</script>") {
		t.Error("타이틀의 스크립트 태그가 이스케이프되지 않음")
	}
	// 이스케이프된 형태가 존재하는지 검증
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Error("타이틀의 스크립트 태그가 HTML 엔티티로 이스케이프되어야 한다")
	}
}

// TestViewerNavigate Navigate 메서드가 WebView에 URL을 전달하는지 검증한다.
func TestViewerNavigate(t *testing.T) {
	mock := &mockWebView{}
	v := &Viewer{webview: mock}

	v.Navigate("http://127.0.0.1:8080")

	if mock.url != "http://127.0.0.1:8080" {
		t.Errorf("Navigate URL: want %q, got %q", "http://127.0.0.1:8080", mock.url)
	}
}

// TestBuildFullHTMLIncludesCSS 생성된 HTML에 CSS가 포함되는지 검증한다.
func TestBuildFullHTMLIncludesCSS(t *testing.T) {
	result, err := BuildFullHTML("Title", "<p>content</p>")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}

	// github-markdown.css 내용이 인라인으로 포함되어야 한다
	if !strings.Contains(result, ".markdown-body") {
		t.Error("BuildFullHTML 결과에 GitHub 마크다운 CSS가 포함되지 않음")
	}
}

// TestBuildFullHTMLWithTheme 테마 파라미터를 지정한 HTML 생성을 검증한다.
func TestBuildFullHTMLWithTheme(t *testing.T) {
	tests := []struct {
		name          string
		theme         string
		expectedTheme string
	}{
		{
			name:          "라이트 테마 지정",
			theme:         "light",
			expectedTheme: "light",
		},
		{
			name:          "다크 테마 지정",
			theme:         "dark",
			expectedTheme: "dark",
		},
		{
			name:          "시스템 테마 지정",
			theme:         "system",
			expectedTheme: "system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildFullHTML("Title", "<p>content</p>", tt.theme)
			if err != nil {
				t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
			}
			if !strings.Contains(result, "<!DOCTYPE html>") {
				t.Error("HTML 문서 구조가 생성되어야 한다")
			}
		})
	}
}

// TestBuildFullHTMLDefaultTheme 테마 미지정 시 기본값(system)이 사용되는지 검증한다.
func TestBuildFullHTMLDefaultTheme(t *testing.T) {
	// 테마를 지정하지 않고 호출
	result, err := BuildFullHTML("Title", "<p>content</p>")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("HTML 문서 구조가 생성되어야 한다")
	}
}

// TestBuildFullHTMLEmptyThemeFallback 빈 테마 문자열 지정 시 기본값이 사용되는지 검증한다.
func TestBuildFullHTMLEmptyThemeFallback(t *testing.T) {
	result, err := BuildFullHTML("Title", "<p>content</p>", "")
	if err != nil {
		t.Fatalf("BuildFullHTML() 오류 발생: %v", err)
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("빈 테마에서도 HTML 문서 구조가 생성되어야 한다")
	}
}
