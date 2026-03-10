// Package main 은 WinMarkdownViewer의 진입점이다.
// CLI 인자로 전달된 Markdown 파일을 내장 HTTP 서버와 WebView2 윈도우로 실시간 미리보기한다.
// 시스템 트레이 아이콘, 단일 인스턴스 제어, 컨텍스트 메뉴 등록을 지원한다.
// WindowManager를 통해 다중 윈도우를 관리하며, 각 윈도우는 별도 고루틴에서 실행된다.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/window"
)

// Windows API 상수
const (
	wmSysCommand = 0x0112 // WM_SYSCOMMAND
	scMinimize   = 0xF020 // SC_MINIMIZE
	swHide       = 0      // SW_HIDE
	swShow       = 5      // SW_SHOW
	gwlWndProc   = -4     // GWL_WNDPROC (SetWindowLongPtrW 인덱스)
)

// Windows 메시지 펌프 상수
const (
	pmRemove = 0x0001 // PM_REMOVE
)

// Windows API 함수
var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procCallWindowProcW     = user32.NewProc("CallWindowProcW")
	procSetWindowLongPtrW   = user32.NewProc("SetWindowLongPtrW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procPeekMessageW        = user32.NewProc("PeekMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
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

// pipeResult 는 handlePipeMessage의 처리 결과를 나타내는 상수이다.
type pipeResult int

const (
	// pipeResultNewWindow 새 윈도우가 생성되었다.
	pipeResultNewWindow pipeResult = iota
	// pipeResultActivated 기존 윈도우가 활성화되었다.
	pipeResultActivated
	// pipeResultError 에러가 발생했다.
	pipeResultError
)

// windowOpener 는 WindowManager.OpenFile 동작을 추상화하는 인터페이스이다.
// 테스트에서 모의 객체로 대체할 수 있다.
type windowOpener interface {
	OpenFile(path string) (int, error)
}

// handlePipeMessage 는 파이프로 수신한 파일 경로를 WindowManager를 통해 처리한다.
// 새 파일이면 윈도우를 생성하고, 이미 열린 파일이면 activateFn 콜백으로 기존 윈도우를 활성화한다.
// 에러 발생 시 pipeResultError를 반환한다 (MessageBox 표시는 호출자가 처리).
func handlePipeMessage(opener windowOpener, filePath string, activateFn func(windowID int)) pipeResult {
	_, err := opener.OpenFile(filePath)
	if err == nil {
		return pipeResultNewWindow
	}

	// 이미 열린 파일이면 기존 윈도우를 활성화한다
	var alreadyOpen *window.FileAlreadyOpenError
	if errors.As(err, &alreadyOpen) {
		activateFn(alreadyOpen.WindowID)
		return pipeResultActivated
	}

	// 기타 에러 (ErrMaxWindowsReached, ErrFileNotFound 등)
	log.Printf("파이프 메시지 처리 실패: %v", err)
	return pipeResultError
}

// serverAdapter 는 server.Server를 window.ServerHandle 인터페이스에 맞추는 어댑터이다.
// Broadcast 메서드를 통해 감시자가 재렌더링된 HTML을 서버로 전달할 수 있다.
type serverAdapter struct {
	srv  *server.Server
	port int
}

func (a *serverAdapter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return a.srv.Shutdown(ctx)
}

func (a *serverAdapter) Port() int {
	return a.port
}

// Broadcast 는 모든 WebSocket 클라이언트에 HTML을 전송한다.
func (a *serverAdapter) Broadcast(html string) {
	a.srv.Broadcast(html)
}

// watcherAdapter 는 watcher.Watcher를 window.Closeable 인터페이스에 맞추는 어댑터이다.
// 파일 변경 이벤트를 수신하여 서버로 브로드캐스트하는 고루틴도 관리한다.
type watcherAdapter struct {
	w      *watcher.Watcher
	cancel context.CancelFunc
}

func (a *watcherAdapter) Close() error {
	a.cancel()
	return a.w.Close()
}

// windowEntry 는 활성 윈도우의 viewer 핸들과 HWND를 추적하는 구조체이다.
type windowEntry struct {
	viewer *viewer.Viewer
	hwnd   windows.HWND
	active bool // 윈도우가 아직 Run() 중인지 여부
}

// windowTracker 는 활성 윈도우 ID -> viewer/HWND 매핑을 관리한다.
// @MX:NOTE: [AUTO] 다중 윈도우 고루틴에서 안전하게 접근하기 위해 sync.Mutex로 보호한다
type windowTracker struct {
	mu      sync.Mutex
	entries map[int]*windowEntry
	wg      sync.WaitGroup
	// windowCount 는 열린 윈도우 수를 추적한다.
	// 모든 윈도우가 닫히면 allClosed 채널을 닫는다.
	windowCount int
	allClosed   chan struct{}
}

// newWindowTracker 는 새로운 windowTracker를 생성한다.
func newWindowTracker() *windowTracker {
	return &windowTracker{
		entries:   make(map[int]*windowEntry),
		allClosed: make(chan struct{}),
	}
}

// add 는 윈도우 항목을 추가한다.
func (t *windowTracker) add(windowID int, v *viewer.Viewer, hwnd windows.HWND) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[windowID] = &windowEntry{viewer: v, hwnd: hwnd, active: true}
	t.windowCount++
	t.wg.Add(1)
}

// remove 는 윈도우를 비활성으로 표시하고 카운트를 감소시킨다.
// 엔트리는 삭제하지 않는다 (Destroy는 앱 종료 시 일괄 처리).
// @MX:NOTE: [AUTO] v.Destroy()를 개별 호출하면 go-webview2의 공유 Environment가
// 해제되어 다른 윈도우가 응답 불가 상태가 된다. 앱 종료 시 destroyAll()에서 일괄 처리한다.
func (t *windowTracker) remove(windowID int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, ok := t.entries[windowID]; ok {
		e.active = false
	}
	t.windowCount--
	t.wg.Done()
	if t.windowCount == 0 {
		select {
		case <-t.allClosed:
			// 이미 닫힘
		default:
			close(t.allClosed)
		}
	}
}

// get 은 윈도우 항목을 조회한다.
func (t *windowTracker) get(windowID int) (*windowEntry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[windowID]
	return e, ok
}

// terminateAll 은 모든 윈도우의 이벤트 루프를 종료시킨다.
func (t *windowTracker) terminateAll() {
	t.mu.Lock()
	entries := make([]*windowEntry, 0, len(t.entries))
	for _, e := range t.entries {
		entries = append(entries, e)
	}
	t.mu.Unlock()

	for _, e := range entries {
		e.viewer.Terminate()
	}
}

// wait 는 모든 윈도우 고루틴이 종료될 때까지 대기한다.
func (t *windowTracker) wait() {
	t.wg.Wait()
}

// destroyAll 은 모든 viewer를 일괄 파괴한다.
// 앱 종료 시 호출하여 WebView2 리소스를 정리한다.
// @MX:NOTE: [AUTO] 개별 윈도우 닫기 시 Destroy()를 호출하면 go-webview2의 공유
// ICoreWebView2Environment가 해제되어 남은 윈도우가 응답 불가 상태가 된다.
// 따라서 모든 윈도우 고루틴이 종료된 후 일괄 파괴한다.
func (t *windowTracker) destroyAll() {
	t.mu.Lock()
	entries := make([]*windowEntry, 0, len(t.entries))
	for _, e := range t.entries {
		entries = append(entries, e)
	}
	t.entries = make(map[int]*windowEntry)
	t.mu.Unlock()

	for _, e := range entries {
		e.viewer.Destroy()
	}
}

// viewerFactory 는 viewer.New를 추상화하는 팩토리 함수 타입이다.
// 테스트에서 모의 객체로 대체할 수 있다.
type viewerFactory func(cfg viewer.Config) (*viewer.Viewer, error)

// openWindowParams 는 openWindow 함수에 전달할 매개변수를 담는 구조체이다.
type openWindowParams struct {
	wm            *window.WindowManager
	tracker       *windowTracker
	cfg           *config.Config
	filePath      string
	newViewerFn   viewerFactory
	shutdownCh    <-chan struct{} // 앱 종료 시그널 (OS 스레드 유지용)
}

// openWindow 는 WindowManager를 통해 서버/감시자를 생성하고,
// 별도 고루틴에서 WebView2 윈도우를 실행한다.
// @MX:ANCHOR: [AUTO] 다중 윈도우 생성의 핵심 함수. 각 윈도우를 독립 고루틴에서 실행한다
// @MX:REASON: 고루틴별 runtime.LockOSThread + viewer.Run 블로킹 패턴 (fan_in >= 3)
func openWindow(ctx context.Context, params openWindowParams) (int, error) {
	// 1. WindowManager로 서버/감시자 생성
	windowID, err := params.wm.OpenFile(params.filePath)
	if err != nil {
		return 0, err
	}

	// 2. 윈도우 정보 조회 (포트 번호)
	winInfo, ok := params.wm.GetWindow(windowID)
	if !ok {
		return 0, fmt.Errorf("윈도우 정보 조회 실패: windowID=%d", windowID)
	}

	// 3. 별도 고루틴에서 WebView2 윈도우 실행
	// @MX:WARN: [AUTO] runtime.LockOSThread 필수 - WebView2는 COM 기반이므로 스레드 고정 필요
	// @MX:REASON: go-webview2는 OS 스레드에 고정되어야 정상 동작한다 (PoC 검증 완료)
	errCh := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// viewer 생성
		viewerCfg := viewer.Config{
			Width:  params.cfg.WindowWidth,
			Height: params.cfg.WindowHeight,
		}
		v, vErr := params.newViewerFn(viewerCfg)
		if vErr != nil {
			errCh <- vErr
			return
		}
		errCh <- nil

		// 윈도우 설정
		v.SetTitle(params.filePath)
		v.Navigate(fmt.Sprintf("http://127.0.0.1:%d", winInfo.Port))

		// HWND 획득 및 tracker 등록
		hwnd := windows.HWND(uintptr(v.Window()))
		params.tracker.add(windowID, v, hwnd)

		// WebView2 이벤트 루프 (윈도우 닫힐 때까지 블로킹)
		v.Run()

		// 윈도우 닫힘 -> 서버/감시자 정리
		params.tracker.remove(windowID)
		_ = params.wm.CloseWindow(windowID)

		// @MX:WARN: [AUTO] 윈도우 닫힌 후에도 메시지 펌프를 유지해야 한다.
		// @MX:REASON: WebView2의 COM 객체가 이 스레드의 Apartment에 바인딩되어 있어,
		// 메시지 펌프가 멈추면 다른 WebView2 인스턴스의 크로스-스레드 COM 콜백이
		// 블로킹되어 응답 불가 상태가 된다. 앱 종료까지 메시지를 펌핑한다.
		if params.shutdownCh != nil {
			pumpMessagesUntilShutdown(params.shutdownCh)
		}

		// 앱 종료 시 viewer 정리
		v.Destroy()
	}()

	// viewer 생성 결과 대기
	if err := <-errCh; err != nil {
		// viewer 생성 실패 시 서버/감시자도 정리
		_ = params.wm.CloseWindow(windowID)
		return 0, fmt.Errorf("WebView2 윈도우 생성 실패: %w", err)
	}

	return windowID, nil
}

// pumpMessagesUntilShutdown 은 현재 OS 스레드에서 Windows 메시지를 계속 펌핑한다.
// shutdownCh 가 닫힐 때까지 메시지 루프를 유지하여 COM STA 크로스-아파트먼트 호출을
// 처리할 수 있도록 한다.
// @MX:WARN: [AUTO] v.Run() 반환 후 반드시 호출해야 한다 - 메시지 펌프가 멈추면
// WebView2 런타임의 크로스-스레드 COM 콜백이 블로킹되어 다른 윈도우가 응답 불가 상태가 된다
// @MX:REASON: WebView2의 ICoreWebView2Environment는 생성 스레드의 COM Apartment에 바인딩된다
func pumpMessagesUntilShutdown(shutdownCh <-chan struct{}) {
	// 48바이트 = sizeof(MSG) on AMD64 Windows
	var msg [48]byte
	for {
		select {
		case <-shutdownCh:
			return
		default:
		}

		ret, _, _ := procPeekMessageW.Call(
			uintptr(unsafe.Pointer(&msg[0])),
			0, 0, 0,
			uintptr(pmRemove),
		)
		if ret != 0 {
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg[0])))
		} else {
			// 메시지가 없으면 짧게 대기하여 CPU 사용률을 줄인다
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func main() {
	os.Exit(run())
}

// run 은 애플리케이션 로직을 실행하고 종료 코드를 반환한다.
// @MX:ANCHOR: [AUTO] 애플리케이션의 메인 진입점으로 단일 인스턴스, 트레이, 다중 윈도우를 조율한다
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

	// 7. WindowManager 생성 (서버/감시자 팩토리 주입)
	// @MX:NOTE: [AUTO] ServerFactory와 WatcherFactory는 각 윈도우별 독립 서버/감시자를 생성한다
	// lastCreatedServer 는 ServerFactory에서 생성한 서버를 WatcherFactory에 전달하는 클로저 변수이다.
	// WindowManager.OpenFile()이 ServerFactory를 먼저 호출하고 WatcherFactory를 호출하므로 안전하다.
	var lastCreatedServer *serverAdapter
	wm := window.NewWindowManager(
		window.WithServerFactory(func(filePath string) (window.ServerHandle, error) {
			adapter, err := createServerForFile(filePath, cfg)
			if err != nil {
				return nil, err
			}
			lastCreatedServer = adapter
			return adapter, nil
		}),
		window.WithWatcherFactory(func(filePath string) (window.Closeable, error) {
			return createWatcherForFile(filePath, lastCreatedServer)
		}),
	)

	// 8. 윈도우 tracker 생성
	tracker := newWindowTracker()

	// 9. 종료 시그널 채널
	// threadShutdownCh 는 모든 WebView2 고루틴의 OS 스레드를 앱 종료까지 유지하는 채널이다.
	// 윈도우가 닫혀도 고루틴을 종료하지 않고 이 채널이 닫힐 때까지 대기한다.
	// 이유: LockOSThread 고루틴이 종료되면 OS 스레드가 파괴되고,
	// WebView2의 공유 COM Apartment가 무효화되어 다른 윈도우가 응답 불가 상태가 된다.
	threadShutdownCh := make(chan struct{})
	quitCh := make(chan struct{}, 1)
	triggerQuit := func() {
		select {
		case quitCh <- struct{}{}:
		default:
		}
	}

	// 10. 시스템 트레이 생성 및 시작
	trayInst, trayErr := tray.NewTray(assets.IconData)
	if trayErr == nil {
		// 트레이에 윈도우 목록 제공자와 활성화 콜백 설정
		trayInst.SetWindowList(
			func() []tray.WindowListItem {
				wins := wm.GetWindows()
				items := make([]tray.WindowListItem, len(wins))
				for i, w := range wins {
					items[i] = tray.WindowListItem{ID: w.ID, Title: w.Title}
				}
				return items
			},
			func(windowID int) {
				entry, ok := tracker.get(windowID)
				if ok {
					showWindow(entry.hwnd)
				}
			},
		)

		// WindowManager 콜백으로 트레이 메뉴 갱신
		wm.OnWindowOpened(func(_ window.WindowInfo) {
			trayInst.RefreshMenu()
		})
		wm.OnWindowClosed(func(_ window.WindowInfo) {
			trayInst.RefreshMenu()
		})

		// 트레이를 별도 고루틴에서 실행
		go trayInst.Run(
			nil, // onReady: 추가 초기화 불필요
			func() {
				// 트레이 종료 시 모든 윈도우를 닫고 앱을 종료한다
				tracker.terminateAll()
				triggerQuit()
			},
		)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 11. 첫 번째 윈도우 열기
	_, err = openWindow(ctx, openWindowParams{
		wm:          wm,
		tracker:     tracker,
		cfg:         cfg,
		filePath:    absPath,
		newViewerFn: viewer.New,
		shutdownCh:  threadShutdownCh,
	})
	if err != nil {
		if trayErr == nil {
			trayInst.Quit()
		}
		wm.Shutdown()
		fmt.Fprintln(os.Stderr, fmt.Errorf("첫 번째 윈도우 열기 실패: %v", err))
		return app.ExitError
	}

	// 12. Named Pipe 서버 시작 (다른 인스턴스로부터 파일 경로 수신)
	// @MX:NOTE: [AUTO] 파이프 핸들러는 openWindow를 통해 새 윈도우를 생성하거나 기존 윈도우를 활성화한다
	go func() {
		_ = app.ListenPipe(ctx, func(newPath string) {
			result := handlePipeMessage(wm, newPath, func(windowID int) {
				// 이미 열린 파일이면 기존 윈도우를 활성화한다
				entry, ok := tracker.get(windowID)
				if ok {
					showWindow(entry.hwnd)
				}
			})
			switch result {
			case pipeResultNewWindow:
				// wm.OpenFile은 handlePipeMessage에서 이미 호출됨 (서버/감시자 생성 완료)
				// 해당 윈도우에 대한 WebView2 윈도우를 열어야 한다
				// 하지만 handlePipeMessage는 windowID를 반환하지 않으므로
				// FindByPath로 다시 조회한다
				winInfo, found := wm.FindByPath(newPath)
				if found {
					go func() {
						runtime.LockOSThread()
						defer runtime.UnlockOSThread()

						viewerCfg := viewer.Config{
							Width:  cfg.WindowWidth,
							Height: cfg.WindowHeight,
						}
						v, vErr := viewer.New(viewerCfg)
						if vErr != nil {
							log.Printf("파이프: WebView2 윈도우 생성 실패: %v", vErr)
							_ = wm.CloseWindow(winInfo.ID)
							return
						}

						v.SetTitle(newPath)
						v.Navigate(fmt.Sprintf("http://127.0.0.1:%d", winInfo.Port))

						hwnd := windows.HWND(uintptr(v.Window()))
						tracker.add(winInfo.ID, v, hwnd)

						v.Run()

						// 서버/감시자 정리
						tracker.remove(winInfo.ID)
						_ = wm.CloseWindow(winInfo.ID)

						// 메시지 펌프 유지 (COM Apartment 크로스-스레드 콜백 처리)
						pumpMessagesUntilShutdown(threadShutdownCh)

						// 앱 종료 시 viewer 정리
						v.Destroy()
					}()
				}
			case pipeResultError:
				log.Printf("파이프 파일 열기 실패: %s", newPath)
			}
		})
	}()

	// 13. 종료 대기: 모든 윈도우가 닫히거나 트레이에서 종료를 선택할 때까지 대기
	select {
	case <-tracker.allClosed:
		// 모든 윈도우가 닫힘 -> 앱 종료
	case <-quitCh:
		// 트레이에서 종료 선택 -> 모든 윈도우 종료 대기
		tracker.wait()
	}

	// 14. 종료 처리: 설정 저장 -> 리소스 정리
	cfg.LastOpenedFile = absPath
	_ = config.Save(cfg)

	if trayErr == nil {
		trayInst.Quit()
	}

	cancel()

	// WindowManager가 모든 윈도우의 서버/감시자를 정리한다
	wm.Shutdown()

	// OS 스레드 유지 채널을 닫아 고루틴들이 v.Destroy() 후 종료되도록 한다
	close(threadShutdownCh)

	// 고루틴들이 v.Destroy()를 완료할 시간을 준다
	time.Sleep(100 * time.Millisecond)

	return app.ExitSuccess
}

// createServerForFile 은 파일에 대한 HTTP 서버를 생성하고 시작한다.
// 마크다운을 렌더링하고 서버에 초기 콘텐츠를 설정한다.
func createServerForFile(filePath string, cfg *config.Config) (*serverAdapter, error) {
	rendered, err := app.RenderMarkdown(filePath)
	if err != nil {
		return nil, err
	}

	srv, err := server.NewServer()
	if err != nil {
		return nil, fmt.Errorf("서버 생성 실패: %w", err)
	}

	filename := filepath.Base(filePath)
	srv.SetTitle(filename + " - WinMarkdownViewer")
	srv.SetFontSize(cfg.FontSize)
	srv.SetTheme(cfg.Theme)
	srv.SetContent(rendered)

	port, err := srv.Start()
	if err != nil {
		return nil, err
	}

	return &serverAdapter{srv: srv, port: port}, nil
}

// createWatcherForFile 은 파일 감시자를 생성하고 변경 이벤트 시 서버로 브로드캐스트하는 고루틴을 시작한다.
// @MX:NOTE: [AUTO] 각 윈도우별 독립적인 감시자와 재렌더링 고루틴을 생성한다
func createWatcherForFile(filePath string, srvAdapter *serverAdapter) (window.Closeable, error) {
	w, err := watcher.NewWatcher(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일 감시 시작 실패: %w", err)
	}

	ctx, watchCancel := context.WithCancel(context.Background())
	changes := w.Watch(ctx)

	// 파일 변경 이벤트 -> 재렌더링 -> 서버 브로드캐스트 고루틴
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-changes:
				if !ok {
					return
				}
				newHTML, renderErr := app.RenderMarkdown(filePath)
				if renderErr != nil {
					srvAdapter.Broadcast(fmt.Sprintf("<p style='color:red;'>렌더링 오류: %v</p>", renderErr))
					continue
				}
				srvAdapter.Broadcast(newHTML)
			}
		}
	}()

	return &watcherAdapter{w: w, cancel: watchCancel}, nil
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
