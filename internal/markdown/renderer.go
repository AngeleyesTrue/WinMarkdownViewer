// Package markdown 는 goldmark을 사용하여 Markdown을 HTML로 변환하는 렌더링 엔진을 제공한다.
package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// md 는 GFM 확장과 구문 강조가 설정된 goldmark 인스턴스이다.
// goldmark은 Markdown을 안전한 HTML로 변환하며, XSS에 취약한 스크립트 태그를 생성하지 않는다.
// WithUnsafe() 옵션은 원본 마크다운 내 HTML 블록(예: <details>, <summary>)을 허용하기 위해 사용되며,
// WebView2의 CSP 정책(script-src 'none')이 추가 방어 계층으로 작동한다.
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM, // 테이블, 취소선, 자동링크, 태스크리스트
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			// WithFormatOptions()는 인라인 스타일 기반 구문 강조를 사용하여
			// 외부 CSS 파일 의존 없이 안전하게 코드 하이라이팅을 적용한다.
			highlighting.WithFormatOptions(),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // 원본 HTML 블록 허용 (CSP로 스크립트 실행 차단)
	),
)

// Render 는 Markdown 바이트 슬라이스를 HTML 문자열로 변환한다.
// 빈 입력이나 nil 입력의 경우 빈 문자열을 반환한다.
func Render(content []byte) (string, error) {
	if len(content) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
