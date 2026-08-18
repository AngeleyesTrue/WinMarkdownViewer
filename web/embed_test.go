package web

import (
	"strings"
	"testing"
)

// TestViewerHTMLEmbed 뷰어 HTML 템플릿이 정상적으로 임베딩되었는지 검증한다.
func TestViewerHTMLEmbed(t *testing.T) {
	if len(ViewerHTML) == 0 {
		t.Fatal("ViewerHTML이 비어있음: 템플릿이 임베딩되지 않았다")
	}

	html := string(ViewerHTML)

	// HTML5 문서 구조 확인
	requiredStrings := []string{
		"<!DOCTYPE html>",
		"<meta charset=\"UTF-8\">",
		"{{.Title}}",
		"{{.Content}}",
		"{{.CSS}}",
		"markdown-body",
	}

	for _, want := range requiredStrings {
		if !strings.Contains(html, want) {
			t.Errorf("ViewerHTML에 %q 가 포함되지 않음", want)
		}
	}
}

// TestGitHubMarkdownCSSEmbed GitHub 마크다운 CSS가 정상적으로 임베딩되었는지 검증한다.
func TestGitHubMarkdownCSSEmbed(t *testing.T) {
	if len(GitHubMarkdownCSS) == 0 {
		t.Fatal("GitHubMarkdownCSS가 비어있음: CSS가 임베딩되지 않았다")
	}

	css := string(GitHubMarkdownCSS)

	// 핵심 CSS 클래스 존재 확인
	requiredClasses := []string{
		".markdown-body",
		".markdown-body h1",
		".markdown-body table",
		".markdown-body code",
		".markdown-body pre",
	}

	for _, want := range requiredClasses {
		if !strings.Contains(css, want) {
			t.Errorf("GitHubMarkdownCSS에 %q 클래스가 포함되지 않음", want)
		}
	}
}

// TestTemplateFSReadable TemplateFS에서 파일을 읽을 수 있는지 검증한다.
func TestTemplateFSReadable(t *testing.T) {
	data, err := TemplateFS.ReadFile("templates/viewer.html")
	if err != nil {
		t.Fatalf("TemplateFS에서 viewer.html 읽기 실패: %v", err)
	}
	if len(data) == 0 {
		t.Error("TemplateFS의 viewer.html이 비어있음")
	}
}

// TestViewerHTMLContainsExtensionTags viewer.html에 확장 렌더링 관련 태그가 포함되어 있는지 검증한다.
func TestViewerHTMLContainsExtensionTags(t *testing.T) {
	html := string(ViewerHTML)

	// KaTeX CSS 링크 확인
	if !strings.Contains(html, "/static/css/katex.min.css") {
		t.Error("ViewerHTML에 KaTeX CSS 링크가 포함되지 않음")
	}

	// KaTeX JS 스크립트 확인
	if !strings.Contains(html, "/static/js/katex.min.js") {
		t.Error("ViewerHTML에 KaTeX JS 스크립트가 포함되지 않음")
	}

	// Mermaid JS 스크립트 확인
	if !strings.Contains(html, "/static/js/mermaid.min.js") {
		t.Error("ViewerHTML에 Mermaid JS 스크립트가 포함되지 않음")
	}

	// render-extensions.js 스크립트 확인
	if !strings.Contains(html, "/static/js/render-extensions.js") {
		t.Error("ViewerHTML에 render-extensions.js 스크립트가 포함되지 않음")
	}

	// renderExtensions 호출 확인 (WebSocket 업데이트 시)
	if !strings.Contains(html, "renderExtensions") {
		t.Error("ViewerHTML에 renderExtensions 호출이 포함되지 않음")
	}

	// CSP 정책에 script-src 로컬 서버 허용 확인
	if !strings.Contains(html, "script-src 'unsafe-inline' http://127.0.0.1:*") {
		t.Error("ViewerHTML CSP에 로컬 서버 스크립트 허용이 포함되지 않음")
	}

	// CSP 정책에 font-src 확인
	if !strings.Contains(html, "font-src http://127.0.0.1:*") {
		t.Error("ViewerHTML CSP에 font-src 로컬 서버 허용이 포함되지 않음")
	}
}

// TestMermaidJSEmbed Mermaid JS 파일이 정상적으로 임베딩되었는지 검증한다.
func TestMermaidJSEmbed(t *testing.T) {
	data, err := ExtensionAssets.ReadFile("js/mermaid.min.js")
	if err != nil {
		t.Fatalf("Mermaid JS 파일 읽기 실패: %v", err)
	}
	if len(data) == 0 {
		t.Error("mermaid.min.js가 비어있음")
	}
}

// TestRenderExtensionsJSEmbed render-extensions.js가 정상적으로 임베딩되었는지 검증한다.
func TestRenderExtensionsJSEmbed(t *testing.T) {
	data, err := ExtensionAssets.ReadFile("js/render-extensions.js")
	if err != nil {
		t.Fatalf("render-extensions.js 파일 읽기 실패: %v", err)
	}
	jsStr := string(data)

	// 핵심 함수 존재 확인
	requiredFunctions := []string{
		"renderMath",
		"renderMermaid",
		"renderExtensions",
		"renderKaTeX",
	}
	for _, fn := range requiredFunctions {
		if !strings.Contains(jsStr, fn) {
			t.Errorf("render-extensions.js에 %q 함수가 포함되지 않음", fn)
		}
	}
}

// TestCSSFSReadable CSSFS에서 파일을 읽을 수 있는지 검증한다.
func TestCSSFSReadable(t *testing.T) {
	data, err := CSSFS.ReadFile("css/github-markdown.css")
	if err != nil {
		t.Fatalf("CSSFS에서 github-markdown.css 읽기 실패: %v", err)
	}
	if len(data) == 0 {
		t.Error("CSSFS의 github-markdown.css가 비어있음")
	}
}

// TestKaTeXFilesEmbed KaTeX JS, CSS, 폰트 파일이 ExtensionAssets에 정상적으로 임베딩되었는지 검증한다.
func TestKaTeXFilesEmbed(t *testing.T) {
	// KaTeX JS 파일 확인
	jsData, err := ExtensionAssets.ReadFile("js/katex.min.js")
	if err != nil {
		t.Fatalf("KaTeX JS 파일 읽기 실패: %v", err)
	}
	if len(jsData) == 0 {
		t.Error("katex.min.js가 비어있음")
	}

	// KaTeX CSS 파일 확인
	cssData, err := ExtensionAssets.ReadFile("css/katex.min.css")
	if err != nil {
		t.Fatalf("KaTeX CSS 파일 읽기 실패: %v", err)
	}
	if len(cssData) == 0 {
		t.Error("katex.min.css가 비어있음")
	}

	// CSS에 수정된 폰트 경로 확인 (../fonts/ 형태여야 한다)
	cssStr := string(cssData)
	if !strings.Contains(cssStr, "url(../fonts/") {
		t.Error("KaTeX CSS에 수정된 폰트 경로(../fonts/)가 포함되지 않음")
	}

	// woff2 폰트 파일 최소 1개 확인
	entries, err := ExtensionAssets.ReadDir("fonts")
	if err != nil {
		t.Fatalf("fonts 디렉토리 읽기 실패: %v", err)
	}

	woff2Count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".woff2") {
			woff2Count++
		}
	}
	if woff2Count == 0 {
		t.Error("KaTeX woff2 폰트 파일이 하나도 없음")
	}
}

// TestExtensionAssetsEmbed ExtensionAssets 임베디드 파일시스템이 js, css, fonts 디렉토리를 포함하는지 검증한다.
func TestExtensionAssetsEmbed(t *testing.T) {
	// js 디렉토리 접근 가능 확인
	entries, err := ExtensionAssets.ReadDir("js")
	if err != nil {
		t.Fatalf("ExtensionAssets에서 js 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Error("ExtensionAssets의 js 디렉토리가 비어있음")
	}

	// css 디렉토리 접근 가능 확인
	entries, err = ExtensionAssets.ReadDir("css")
	if err != nil {
		t.Fatalf("ExtensionAssets에서 css 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Error("ExtensionAssets의 css 디렉토리가 비어있음")
	}

	// fonts 디렉토리 접근 가능 확인
	entries, err = ExtensionAssets.ReadDir("fonts")
	if err != nil {
		t.Fatalf("ExtensionAssets에서 fonts 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Error("ExtensionAssets의 fonts 디렉토리가 비어있음")
	}
}
