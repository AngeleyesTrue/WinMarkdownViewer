// Package main 은 WinMarkdownViewer의 진입점이다.
// CLI 인자로 전달된 Markdown 파일을 내장 HTTP 서버와 WebView2 윈도우로 실시간 미리보기한다.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/app"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/config"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/server"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/viewer"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/watcher"
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

	// 절대 경로로 변환
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("경로 변환 오류: %v", err))
		return app.ExitError
	}

	// 3. 설정 로드
	cfg, err := config.Load()
	if err != nil {
		// 설정 로드 실패 시 기본값으로 계속 진행
		cfg = config.Default()
	}

	// 마지막 열린 파일 경로 저장
	cfg.LastOpenedFile = absPath

	// 4. 초기 마크다운 렌더링
	rendered, err := app.RenderMarkdown(absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 5. 내장 HTTP 서버 생성 및 시작
	srv, err := server.NewServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("서버 생성 실패: %v", err))
		return app.ExitError
	}

	filename := filepath.Base(absPath)
	srv.SetTitle(filename + " - WinMarkdownViewer")
	srv.SetFontSize(cfg.FontSize)
	srv.SetContent(rendered)

	port, err := srv.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 6. 파일 감시 시작
	w, err := watcher.NewWatcher(absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("파일 감시 시작 실패: %v", err))
		return app.ExitError
	}

	ctx, cancel := context.WithCancel(context.Background())
	changes := w.Watch(ctx)

	// 파일 변경 이벤트 → 재렌더링 → Broadcast 고루틴
	go func() {
		for range changes {
			newHTML, err := app.RenderMarkdown(absPath)
			if err != nil {
				// 렌더링 실패 시 에러 메시지를 브로드캐스트
				srv.Broadcast(fmt.Sprintf("<p style='color:red;'>렌더링 오류: %v</p>", err))
				continue
			}
			srv.Broadcast(newHTML)
		}
	}()

	// 7. WebView2 윈도우 생성
	viewerCfg := viewer.Config{
		Width:  cfg.WindowWidth,
		Height: cfg.WindowHeight,
	}
	v, err := viewer.New(viewerCfg)
	if err != nil {
		cancel()
		w.Close()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 8. 윈도우 설정 및 Navigate
	v.SetTitle(absPath)
	v.Navigate(fmt.Sprintf("http://127.0.0.1:%d", port))

	// 9. WebView2 이벤트 루프 (윈도우 닫힐 때까지 블로킹)
	v.Run()

	// 10. 종료 처리: 설정 저장 → 리소스 정리
	cfg.LastOpenedFile = absPath
	_ = config.Save(cfg)

	cancel()
	w.Close()
	v.Destroy()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	return app.ExitSuccess
}
