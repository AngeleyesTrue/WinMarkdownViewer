// Package web 은 HTML 템플릿과 CSS 등 정적 리소스를 go:embed로 임베딩한다.
package web

import "embed"

// TemplateFS 는 HTML 템플릿 파일을 포함하는 임베디드 파일시스템이다.
//
//go:embed templates/viewer.html
var TemplateFS embed.FS

// CSSFS 는 CSS 파일을 포함하는 임베디드 파일시스템이다.
//
//go:embed css/github-markdown.css
var CSSFS embed.FS

// ViewerHTML 은 뷰어 HTML 템플릿의 원본 바이트이다.
//
//go:embed templates/viewer.html
var ViewerHTML []byte

// GitHubMarkdownCSS 는 GitHub 마크다운 스타일 CSS의 원본 바이트이다.
//
//go:embed css/github-markdown.css
var GitHubMarkdownCSS []byte

// ExtensionAssets 는 확장 렌더링에 필요한 JS, CSS, 폰트 파일을 포함하는 임베디드 파일시스템이다.
// KaTeX 수학 렌더링과 Mermaid 다이어그램에 사용되는 정적 리소스를 제공한다.
//
//go:embed js css fonts
var ExtensionAssets embed.FS
