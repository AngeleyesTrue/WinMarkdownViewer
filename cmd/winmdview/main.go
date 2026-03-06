// Package main 은 WinMarkdownViewer의 진입점이다.
// CLI 인자로 전달된 Markdown 파일을 내장 HTTP 서버와 WebView2 윈도우로 실시간 미리보기한다.
// 시스템 트레이 아이콘, 단일 인스턴스 제어, 컨텍스트 메뉴 등록을 지원한다.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/AngeleyesTrue/WinMarkdownViewer/assets"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/app"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/config"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/registry"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/server"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/tray"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/viewer"
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/watcher"
)

// Windows API 상수
const (
	wmSysCommand = 0x0112 // WM_SYSCOMMAND
	scMinimize   = 0xF020 // SC_MINIMIZE
	swHide       = 0      // SW_HIDE
	swShow       = 5      // SW_SHOW
	gwlWndProc   = -4     // GWL_WNDPROC (SetWindowLongPtrW 인덱스)
)

// Windows API 함수
var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procShowWindow        = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procCallWindowProcW   = user32.NewProc("CallWindowProcW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
)

// appFlags 는 CLI 플래그 파싱 결과를 담는 구조체이다.
type appFlags struct {
	register   bool   // --register: 컨텍스트 메뉴 등록
	unregister bool   // --unregister: 컨텍스트 메뉴 해제
	setDefault bool   // --set-default: 기본 프로그램 설정
	filePath   string // 마크다운 파일 경로 (위치 인자)
}

// parseFlags 는 명령줄 인자를 파싱하여 appFlags를 반환한다.
// flag 패키지 대신 수동 파싱하여 위치 인자(파일 경로)와의 충돌을 방지한다.
func parseFlags(args []string) appFlags {
	var f appFlags
	for _, arg := range args {
		switch arg {
		case "--register":
			f.register = true
		case "--unregister":
			f.unregister = true
		case "--set-default":
			f.setDefault = true
		default:
			if !strings.HasPrefix(arg, "--") {
				f.filePath = arg
			}
		}
	}
	return f
}

// handleRegister 는 컨텍스트 메뉴 등록을 수행한다.
func handleRegister() int {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("실행 파일 경로 확인 실패: %v", err))
		return app.ExitError
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("실행 파일 경로 해석 실패: %v", err))
		return app.ExitError
	}

	if err := registry.Register(exePath); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("레지스트리 등록 실패: %v", err))
		return app.ExitError
	}

	fmt.Println("레지스트리 등록이 완료되었습니다. 파일을 열려면 --register 없이 실행하세요.")
	return app.ExitSuccess
}

// handleUnregister 는 컨텍스트 메뉴 등록을 해제한다.
func handleUnregister() int {
	if err := registry.Unregister(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("레지스트리 해제 실패: %v", err))
		return app.ExitError
	}

	fmt.Println("레지스트리 해제가 완료되었습니다.")
	return app.ExitSuccess
}

func main() {
	os.Exit(run())
}

// run 은 애플리케이션 로직을 실행하고 종료 코드를 반환한다.
// @MX:ANCHOR: [AUTO] 애플리케이션의 메인 진입점으로 단일 인스턴스, 트레이, 뷰어를 조율한다
// @MX:REASON: 전체 앱 생명주기를 관리하는 핵심 함수 (fan_in >= 3)
func run() int {
	// 1. CLI 플래그 파싱
	flags := parseFlags(os.Args[1:])

	// 2. 레지스트리 명령 처리 (ACC-014: --register 시 파일 열기 없이 등록만 수행)
	if flags.register {
		return handleRegister()
	}
	if flags.unregister {
		return handleUnregister()
	}

	// 3. 파일 경로가 없으면 사용법 출력
	if flags.filePath == "" {
		fmt.Fprintln(os.Stderr, app.UsageMessage)
		return app.ExitError
	}

	// 4. 파일 검증
	if err := app.ValidateFile(flags.filePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 절대 경로로 변환
	absPath, err := filepath.Abs(flags.filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("경로 변환 오류: %v", err))
		return app.ExitError
	}

	// 5. 단일 인스턴스 제어
	inst := app.NewInstanceLock()
	locked, err := inst.TryLock()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("인스턴스 잠금 오류: %v", err))
		return app.ExitError
	}
	if !locked {
		// 기존 인스턴스에 파일 경로 전달
		if absPath != "" {
			if err := app.SendPath(absPath); err != nil {
				fmt.Fprintln(os.Stderr, fmt.Errorf("파일 경로 전달 실패: %v", err))
				return app.ExitError
			}
		}
		return app.ExitSuccess
	}
	defer inst.Unlock()

	// 6. 설정 로드
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	cfg.LastOpenedFile = absPath

	// 7. 초기 마크다운 렌더링
	rendered, err := app.RenderMarkdown(absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return app.ExitError
	}

	// 8. 내장 HTTP 서버 생성 및 시작
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

	// 9. 파일 감시 시작
	w, err := watcher.NewWatcher(absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("파일 감시 시작 실패: %v", err))
		return app.ExitError
	}

	ctx, cancel := context.WithCancel(context.Background())
	changes := w.Watch(ctx)

	// 파일 변경 이벤트 -> 재렌더링 -> Broadcast 고루틴
	go func() {
		for range changes {
			newHTML, renderErr := app.RenderMarkdown(absPath)
			if renderErr != nil {
				srv.Broadcast(fmt.Sprintf("<p style='color:red;'>렌더링 오류: %v</p>", renderErr))
				continue
			}
			srv.Broadcast(newHTML)
		}
	}()

	// 10. WebView2 윈도우 생성
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

	// 11. 윈도우 설정 및 Navigate
	v.SetTitle(absPath)
	v.Navigate(fmt.Sprintf("http://127.0.0.1:%d", port))

	// 12. Named Pipe 서버 시작 (다른 인스턴스로부터 파일 경로 수신)
	go app.ListenPipe(ctx, func(newPath string) {
		switchFile(newPath, srv, w, v)
	})

	// 13. 시스템 트레이 시작 (별도 고루틴에서 실행)
	hwnd := windows.HWND(uintptr(v.Window()))
	trayInst, trayErr := tray.NewTray(assets.IconData)
	if trayErr == nil {
		go trayInst.Run(
			func() { showWindow(hwnd) },
			func() { v.Terminate() },
		)
	}

	// 14. 윈도우 서브클래싱: 최소화 시 트레이로 숨기기
	if trayErr == nil {
		subclassWindow(hwnd, v)
	}

	// 15. WebView2 이벤트 루프 (윈도우 닫힐 때까지 블로킹)
	v.Run()

	// 16. 종료 처리: 설정 저장 -> 리소스 정리
	cfg.LastOpenedFile = absPath
	_ = config.Save(cfg)

	if trayErr == nil {
		trayInst.Quit()
	}

	cancel()
	w.Close()
	v.Destroy()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	return app.ExitSuccess
}

// switchFile 은 파이프로 수신한 새 파일 경로로 전환한다.
// 마크다운 재렌더링, 서버 업데이트, 감시 대상 변경, 윈도우 타이틀 변경을 수행한다.
func switchFile(newPath string, srv *server.Server, w *watcher.Watcher, v *viewer.Viewer) {
	rendered, err := app.RenderMarkdown(newPath)
	if err != nil {
		return
	}

	filename := filepath.Base(newPath)
	srv.SetTitle(filename + " - WinMarkdownViewer")
	srv.Broadcast(rendered)

	_ = w.SwitchFile(newPath)

	// UI 조작은 메인 스레드에서 실행해야 한다
	v.Dispatch(func() {
		v.SetTitle(newPath)
	})
}

// showWindow 는 숨겨진 윈도우를 표시하고 포그라운드로 가져온다.
func showWindow(hwnd windows.HWND) {
	procShowWindow.Call(uintptr(hwnd), uintptr(swShow))
	procSetForegroundWindow.Call(uintptr(hwnd))
}

// subclassWindow 는 윈도우 프로시저를 서브클래싱하여 최소화 시 트레이로 숨기도록 한다.
// @MX:WARN: [AUTO] Windows API 직접 호출 - unsafe.Pointer 사용
// @MX:REASON: WebView2 윈도우의 WM_SYSCOMMAND/SC_MINIMIZE를 가로채야 한다
func subclassWindow(hwnd windows.HWND, v *viewer.Viewer) {
	// 기존 윈도우 프로시저를 저장할 변수
	var origWndProc uintptr

	newProc := windows.NewCallback(func(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		// WM_SYSCOMMAND + SC_MINIMIZE 인터셉트
		if msg == wmSysCommand && (wParam&0xFFF0) == scMinimize {
			// 최소화 대신 윈도우를 숨긴다
			procShowWindow.Call(hWnd, uintptr(swHide))
			return 0
		}
		// 다른 메시지는 원래 프로시저로 전달한다
		ret, _, _ := procCallWindowProcW.Call(
			origWndProc,
			hWnd, uintptr(msg), wParam, lParam,
		)
		return ret
	})

	// GWL_WNDPROC = -4, 부호 확장하여 uintptr로 변환
	const gwlpWndProc = ^uintptr(3) // 0xFFFF...FFFC (= -4의 uintptr 표현)
	origWndProc, _, _ = procSetWindowLongPtrW.Call(
		uintptr(hwnd),
		gwlpWndProc,
		newProc,
	)
	// origWndProc가 0이면 서브클래싱 실패 (에러를 무시하고 계속 진행)
	_ = origWndProc
	_ = unsafe.Pointer(nil) // unsafe 패키지 사용을 명시적으로 표시
}
