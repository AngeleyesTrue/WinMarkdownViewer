package markdown

import (
	"strings"
	"testing"
)

// TestRenderBasicMarkdown 기본 마크다운 렌더링을 검증한다.
func TestRenderBasicMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // 결과 HTML에 포함되어야 하는 문자열들
	}{
		{
			name:  "제목 렌더링",
			input: "# Hello World",
			contains: []string{
				"<h1", ">Hello World</h1>",
			},
		},
		{
			name:  "단락 렌더링",
			input: "This is a paragraph.",
			contains: []string{
				"<p>This is a paragraph.</p>",
			},
		},
		{
			name:  "굵은 텍스트",
			input: "**bold text**",
			contains: []string{
				"<strong>bold text</strong>",
			},
		},
		{
			name:  "기울임 텍스트",
			input: "*italic text*",
			contains: []string{
				"<em>italic text</em>",
			},
		},
		{
			name:  "링크 렌더링",
			input: "[Go](https://golang.org)",
			contains: []string{
				`<a href="https://golang.org">Go</a>`,
			},
		},
		{
			name:  "순서 없는 목록",
			input: "- item1\n- item2\n- item3",
			contains: []string{
				"<ul>",
				"<li>item1</li>",
				"<li>item2</li>",
				"<li>item3</li>",
				"</ul>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render([]byte(tt.input))
			if err != nil {
				t.Fatalf("Render() 오류 발생: %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("결과에 %q 가 포함되지 않음.\n결과: %s", want, result)
				}
			}
		})
	}
}

// TestRenderGFMTable GFM 테이블 렌더링을 검증한다.
func TestRenderGFMTable(t *testing.T) {
	input := `| Name | Age |
| --- | --- |
| Alice | 30 |
| Bob | 25 |`

	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	expects := []string{"<table>", "<thead>", "<tbody>", "<th>Name</th>", "<td>Alice</td>"}
	for _, want := range expects {
		if !strings.Contains(result, want) {
			t.Errorf("GFM 테이블 결과에 %q 가 포함되지 않음.\n결과: %s", want, result)
		}
	}
}

// TestRenderGFMStrikethrough GFM 취소선 렌더링을 검증한다.
func TestRenderGFMStrikethrough(t *testing.T) {
	input := "~~deleted text~~"
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	if !strings.Contains(result, "<del>deleted text</del>") {
		t.Errorf("취소선이 렌더링되지 않음.\n결과: %s", result)
	}
}

// TestRenderGFMAutolink GFM 자동 링크를 검증한다.
func TestRenderGFMAutolink(t *testing.T) {
	input := "Visit https://golang.org for more info."
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	if !strings.Contains(result, `<a href="https://golang.org"`) {
		t.Errorf("자동 링크가 렌더링되지 않음.\n결과: %s", result)
	}
}

// TestRenderGFMTaskList GFM 태스크 리스트를 검증한다.
func TestRenderGFMTaskList(t *testing.T) {
	input := "- [x] done\n- [ ] todo"
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	// 태스크 리스트는 checkbox input을 포함해야 한다
	if !strings.Contains(result, `type="checkbox"`) {
		t.Errorf("태스크 리스트 checkbox가 렌더링되지 않음.\n결과: %s", result)
	}
}

// TestRenderCodeBlockWithHighlighting 코드 블록 구문 강조를 검증한다.
func TestRenderCodeBlockWithHighlighting(t *testing.T) {
	input := "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```"
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	// 구문 강조가 적용되면 인라인 스타일 또는 chroma 클래스가 포함된다
	hasHighlighting := strings.Contains(result, "chroma") || strings.Contains(result, "style=\"color:")
	if !hasHighlighting {
		t.Errorf("구문 강조가 적용되지 않음.\n결과: %s", result)
	}
}

// TestRenderCodeBlockWithoutLanguage 언어 미지정 코드 블록을 검증한다.
func TestRenderCodeBlockWithoutLanguage(t *testing.T) {
	input := "```\nplain code\n```"
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	if !strings.Contains(result, "plain code") {
		t.Errorf("코드 블록 내용이 렌더링되지 않음.\n결과: %s", result)
	}
}

// TestRenderEmptyInput 빈 입력 처리를 검증한다.
func TestRenderEmptyInput(t *testing.T) {
	result, err := Render([]byte(""))
	if err != nil {
		t.Fatalf("Render() 빈 입력에서 오류 발생: %v", err)
	}

	// 빈 입력은 빈 문자열 또는 공백만 반환해야 한다
	trimmed := strings.TrimSpace(result)
	if trimmed != "" {
		t.Errorf("빈 입력에 대해 비어있지 않은 결과 반환: %q", result)
	}
}

// TestRenderMermaidCodeBlock mermaid 코드 블록이 language-mermaid 클래스로 렌더링되는지 검증한다.
func TestRenderMermaidCodeBlock(t *testing.T) {
	input := "```mermaid\ngraph TD;\n    A-->B;\n```"
	result, err := Render([]byte(input))
	if err != nil {
		t.Fatalf("Render() 오류 발생: %v", err)
	}

	// goldmark은 코드 블록에 language-mermaid 클래스를 부여해야 한다
	if !strings.Contains(result, "mermaid") {
		t.Errorf("mermaid 코드 블록에 mermaid 관련 클래스/내용이 포함되지 않음.\n결과: %s", result)
	}

	// 내용이 보존되어야 한다
	if !strings.Contains(result, "graph TD") {
		t.Errorf("mermaid 코드 블록 내용이 보존되지 않음.\n결과: %s", result)
	}
}

// TestRenderNilInput nil 입력 처리를 검증한다.
func TestRenderNilInput(t *testing.T) {
	result, err := Render(nil)
	if err != nil {
		t.Fatalf("Render() nil 입력에서 오류 발생: %v", err)
	}

	trimmed := strings.TrimSpace(result)
	if trimmed != "" {
		t.Errorf("nil 입력에 대해 비어있지 않은 결과 반환: %q", result)
	}
}
