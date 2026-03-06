// Package main 은 WinMarkdownViewer의 진입점이다.
// CLI 인자로 전달된 Markdown 파일을 WebView2 윈도우에서 렌더링하여 표시한다.
package main

import (
	"fmt"
	"os"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/app"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/viewer"
)

func main() {
	os.Exit(run())
}

// run 은 애플리케이션 로직을 실행하고 종료 코드를 반환한다.
func run() int {
	// 1. CLI 인자 파싱
	filePath, err := app.ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 2. 파일 검증
	if err := app.ValidateFile(filePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 3. 렌더링 파이프라인 실행
	html, err := app.RenderPipeline(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 4. WebView2 윈도우 생성
	cfg := viewer.DefaultConfig()
	v, err := viewer.New(cfg)
	if err != nil {
		// REQ-S-003: WebView2 Runtime 미설치 안내
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}
	defer v.Destroy()

	// 5. 윈도우 설정 및 표시
	v.SetTitle(filePath)
	v.LoadHTML(html)
	v.Run()

	return app.ExitSuccess
}
