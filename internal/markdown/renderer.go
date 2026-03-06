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
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM, // 테이블, 취소선, 자동링크, 태스크리스트
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // 원본 HTML 허용
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
