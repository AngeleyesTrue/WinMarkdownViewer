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
