package window

import (
	"path/filepath"
	"sync"
)

// Closeable 은 리소스 정리를 위한 인터페이스이다.
// server.Server와 watcher.Watcher를 추상화하여 테스트에서 모의 객체로 대체할 수 있다.
type Closeable interface {
	Close() error
}

// ServerHandle 은 윈도우가 소유하는 서버의 인터페이스이다.
// Closeable을 포함하고, 포트 정보를 제공한다.
type ServerHandle interface {
	Closeable
	Port() int
}

// WindowInfo 는 외부에서 조회할 수 있는 윈도우 읽기 전용 정보이다.
type WindowInfo struct {
	ID       int    // 윈도우 고유 ID
	FilePath string // 파일 절대 경로
	Title    string // 표시용 파일명
	Port     int    // 서버 포트
}

// Window 는 개별 마크다운 뷰어 윈도우를 나타낸다.
// 각 윈도우는 독립적인 Server와 Watcher 인스턴스를 소유한다.
type Window struct {
	mu       sync.Mutex
	id       int
	filePath string
	title    string
	port     int
	server   ServerHandle
	watcher  Closeable
	closed   bool
}

// NewWindow 는 새로운 Window 인스턴스를 생성한다.
func NewWindow(id int, filePath string, srv ServerHandle, w Closeable) *Window {
	return &Window{
		id:       id,
		filePath: filePath,
		title:    filepath.Base(filePath),
		port:     srv.Port(),
		server:   srv,
		watcher:  w,
	}
}

// Close 는 윈도우의 리소스를 정리한다.
// Watcher를 먼저 닫고 Server를 닫는다 (순서 중요).
// 이미 닫힌 경우에도 안전하게 호출할 수 있다.
func (w *Window) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	// Watcher를 먼저 닫는다 (파일 변경 이벤트 중단)
	var firstErr error
	if w.watcher != nil {
		if err := w.watcher.Close(); err != nil {
			firstErr = err
		}
	}

	// Server를 닫는다
	if w.server != nil {
		if err := w.server.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Info 는 윈도우의 읽기 전용 정보를 반환한다.
func (w *Window) Info() WindowInfo {
	return WindowInfo{
		ID:       w.id,
		FilePath: w.filePath,
		Title:    w.title,
		Port:     w.port,
	}
}

// ID 는 윈도우 ID를 반환한다.
func (w *Window) ID() int {
	return w.id
}

// FilePath 는 윈도우의 파일 절대 경로를 반환한다.
func (w *Window) FilePath() string {
	return w.filePath
}
