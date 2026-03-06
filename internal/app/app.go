// Package app 은 CLI 인자 파싱과 렌더링 파이프라인을 조율하는 애플리케이션 로직을 제공한다.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/markdown"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/viewer"
)

const (
	// ExitSuccess 정상 종료 코드
	ExitSuccess = 0
	// ExitError 오류 종료 코드
	ExitError = 1
)

// UsageMessage 는 프로그램 사용법 안내 메시지이다.
const UsageMessage = "사용법: winmdview <파일경로.md>\n\n" +
	"Markdown 파일을 WebView2 윈도우에서 렌더링하여 표시합니다."

// ParseArgs 는 명령줄 인자를 파싱하여 Markdown 파일 경로를 반환한다.
// 인자가 없으면 에러를 반환한다.
func ParseArgs(args []string) (string, error) {
	// args[0]은 프로그램 경로이므로 args[1]부터 확인
	if len(args) < 2 {
		return "", fmt.Errorf("%s", UsageMessage)
	}
	return args[1], nil
}

// ValidateFile 은 파일의 존재 여부와 읽기 권한을 검증한다.
func ValidateFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("파일을 찾을 수 없습니다: %s", filePath)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("파일 읽기 권한이 없습니다: %s", filePath)
		}
		return fmt.Errorf("파일 접근 오류: %v", err)
	}
	if info.IsDir() {
		return fmt.Errorf("디렉토리는 열 수 없습니다: %s", filePath)
	}
	return nil
}

// ReadFile 은 파일 내용을 바이트 슬라이스로 읽는다.
// ValidateFile()로 사전 검증된 파일 경로를 전제로 하므로,
// 권한 오류는 ValidateFile과 ReadFile 사이에 권한이 변경된 경쟁 조건 대비용이다.
func ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("파일 읽기 권한이 없습니다: %s", filePath)
		}
		return nil, fmt.Errorf("파일 읽기 오류: %v", err)
	}
	return data, nil
}

// RenderPipeline 은 Markdown 파일을 읽어 완전한 HTML 문서를 생성하는 파이프라인이다.
// 빈 파일의 경우 "내용이 없습니다" 메시지를 포함한 HTML을 반환한다.
func RenderPipeline(filePath string) (string, error) {
	content, err := ReadFile(filePath)
	if err != nil {
		return "", err
	}

	filename := filepath.Base(filePath)

	// 빈 파일 처리 (REQ-O-002)
	// 공백만 있는 파일도 빈 파일로 취급하여 사용자에게 안내 메시지를 표시한다.
	if len(strings.TrimSpace(string(content))) == 0 {
		return viewer.BuildFullHTML(filename, "<p><em>내용이 없습니다.</em></p>")
	}

	rendered, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("마크다운 렌더링 오류: %v", err)
	}

	return viewer.BuildFullHTML(filename, rendered)
}
