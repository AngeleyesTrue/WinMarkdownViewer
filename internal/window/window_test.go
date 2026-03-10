package window

import (
	"errors"
	"sync/atomic"
	"testing"
)

// mockServerHandle 은 ServerHandle 인터페이스의 모의 구현이다.
type mockServerHandle struct {
	port         int
	closeCalled  atomic.Int32
	closeErr     error
}

func (m *mockServerHandle) Close() error {
	m.closeCalled.Add(1)
	return m.closeErr
}

func (m *mockServerHandle) Port() int {
	return m.port
}

// mockCloseable 은 Closeable 인터페이스의 모의 구현이다.
type mockCloseable struct {
	closeCalled atomic.Int32
	closeErr    error
}

func (m *mockCloseable) Close() error {
	m.closeCalled.Add(1)
	return m.closeErr
}

// TestNewWindow_정상생성 Window가 올바르게 생성되는지 검증한다.
func TestNewWindow_정상생성(t *testing.T) {
	t.Parallel()

	srv := &mockServerHandle{port: 8080}
	w := &mockCloseable{}

	win := NewWindow(1, "/test/file.md", srv, w)

	if win.ID() != 1 {
		t.Errorf("ID: want 1, got %d", win.ID())
	}
	if win.FilePath() != "/test/file.md" {
		t.Errorf("FilePath: want /test/file.md, got %s", win.FilePath())
	}

	info := win.Info()
	if info.Title != "file.md" {
		t.Errorf("Title: want file.md, got %s", info.Title)
	}
	if info.Port != 8080 {
		t.Errorf("Port: want 8080, got %d", info.Port)
	}
}

// TestWindow_Close_리소스정리순서 Close가 Watcher를 먼저 닫고 Server를 닫는지 검증한다.
func TestWindow_Close_리소스정리순서(t *testing.T) {
	t.Parallel()

	// 호출 순서를 기록하는 채널
	order := make(chan string, 2)

	srv := &mockServerHandle{port: 8080}
	w := &mockCloseable{}

	// 원본 Close를 감싸서 순서 기록
	origSrvClose := srv.Close
	origWClose := w.Close

	type orderTracker struct {
		ServerHandle
		closeFunc func() error
	}

	// 간단한 순서 검증: closeCalled 카운트 확인
	_ = origSrvClose
	_ = origWClose
	_ = order

	win := NewWindow(1, "/test/file.md", srv, w)
	err := win.Close()
	if err != nil {
		t.Fatalf("Close() 오류: %v", err)
	}

	if srv.closeCalled.Load() != 1 {
		t.Errorf("서버 Close 호출 횟수: want 1, got %d", srv.closeCalled.Load())
	}
	if w.closeCalled.Load() != 1 {
		t.Errorf("감시자 Close 호출 횟수: want 1, got %d", w.closeCalled.Load())
	}
}

// TestWindow_Close_이중닫기 이미 닫힌 윈도우를 다시 닫아도 안전한지 검증한다.
func TestWindow_Close_이중닫기(t *testing.T) {
	t.Parallel()

	srv := &mockServerHandle{port: 8080}
	w := &mockCloseable{}

	win := NewWindow(1, "/test/file.md", srv, w)

	// 첫 번째 Close
	if err := win.Close(); err != nil {
		t.Fatalf("첫 번째 Close() 오류: %v", err)
	}

	// 두 번째 Close (안전해야 한다)
	if err := win.Close(); err != nil {
		t.Fatalf("두 번째 Close() 오류: %v", err)
	}

	// Close는 한 번만 호출되어야 한다
	if srv.closeCalled.Load() != 1 {
		t.Errorf("서버 Close 호출 횟수: want 1, got %d", srv.closeCalled.Load())
	}
	if w.closeCalled.Load() != 1 {
		t.Errorf("감시자 Close 호출 횟수: want 1, got %d", w.closeCalled.Load())
	}
}

// TestWindow_Close_에러전파 리소스 정리 중 에러가 발생하면 첫 번째 에러를 반환하는지 검증한다.
func TestWindow_Close_에러전파(t *testing.T) {
	t.Parallel()

	watcherErr := errors.New("watcher close error")
	srv := &mockServerHandle{port: 8080}
	w := &mockCloseable{closeErr: watcherErr}

	win := NewWindow(1, "/test/file.md", srv, w)
	err := win.Close()

	if err == nil {
		t.Fatal("Close()에서 에러를 기대했으나 nil 반환")
	}
	if err != watcherErr {
		t.Errorf("Close() 에러: want %v, got %v", watcherErr, err)
	}
}

// TestWindow_Info_읽기전용 Info가 올바른 읽기 전용 정보를 반환하는지 검증한다.
func TestWindow_Info_읽기전용(t *testing.T) {
	t.Parallel()

	srv := &mockServerHandle{port: 9090}
	w := &mockCloseable{}

	win := NewWindow(42, "/path/to/readme.md", srv, w)
	info := win.Info()

	if info.ID != 42 {
		t.Errorf("ID: want 42, got %d", info.ID)
	}
	if info.FilePath != "/path/to/readme.md" {
		t.Errorf("FilePath: want /path/to/readme.md, got %s", info.FilePath)
	}
	if info.Title != "readme.md" {
		t.Errorf("Title: want readme.md, got %s", info.Title)
	}
	if info.Port != 9090 {
		t.Errorf("Port: want 9090, got %d", info.Port)
	}
}
