package window

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testServerHandle 은 테스트용 ServerHandle 구현이다.
type testServerHandle struct {
	port        int
	closeCalled atomic.Int32
}

func (s *testServerHandle) Close() error {
	s.closeCalled.Add(1)
	return nil
}

func (s *testServerHandle) Port() int {
	return s.port
}

// testCloseable 은 테스트용 Closeable 구현이다.
type testCloseable struct {
	closeCalled atomic.Int32
}

func (c *testCloseable) Close() error {
	c.closeCalled.Add(1)
	return nil
}

// nextPort 는 테스트에서 고유한 포트 번호를 생성한다.
var nextPort atomic.Int32

func init() {
	nextPort.Store(10000)
}

// newTestManager 는 모의 팩토리가 설정된 테스트용 WindowManager를 생성한다.
func newTestManager(opts ...ManagerOption) *WindowManager {
	defaultOpts := []ManagerOption{
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			p := int(nextPort.Add(1))
			return &testServerHandle{port: p}, nil
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			return &testCloseable{}, nil
		}),
	}
	allOpts := append(defaultOpts, opts...)
	return NewWindowManager(allOpts...)
}

// createTempFile 은 테스트용 임시 파일을 생성한다.
func createTempFile(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("# Test"), 0644); err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	return path
}

// TestNewWindowManager_기본설정 기본 설정으로 WindowManager가 생성되는지 검증한다.
func TestNewWindowManager_기본설정(t *testing.T) {
	t.Parallel()

	m := NewWindowManager()
	if m.maxWindows != defaultMaxWindows {
		t.Errorf("maxWindows: want %d, got %d", defaultMaxWindows, m.maxWindows)
	}
	if m.Count() != 0 {
		t.Errorf("Count: want 0, got %d", m.Count())
	}
}

// TestNewWindowManager_옵션적용 WithMaxWindows 옵션이 올바르게 적용되는지 검증한다.
func TestNewWindowManager_옵션적용(t *testing.T) {
	t.Parallel()

	m := NewWindowManager(WithMaxWindows(5))
	if m.maxWindows != 5 {
		t.Errorf("maxWindows: want 5, got %d", m.maxWindows)
	}
}

// TestOpenFile_새파일 새 파일을 열면 윈도우가 생성되는지 검증한다 (AC2.1).
func TestOpenFile_새파일(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path := createTempFile(t, "test.md")

	id, err := m.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() 오류: %v", err)
	}
	if id != 1 {
		t.Errorf("첫 번째 윈도우 ID: want 1, got %d", id)
	}
	if m.Count() != 1 {
		t.Errorf("Count: want 1, got %d", m.Count())
	}
}

// TestOpenFile_연속ID 여러 파일을 열면 순차 ID가 할당되는지 검증한다.
func TestOpenFile_연속ID(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path1 := createTempFile(t, "test1.md")
	path2 := createTempFile(t, "test2.md")

	id1, err := m.OpenFile(path1)
	if err != nil {
		t.Fatalf("OpenFile(1) 오류: %v", err)
	}
	id2, err := m.OpenFile(path2)
	if err != nil {
		t.Fatalf("OpenFile(2) 오류: %v", err)
	}

	if id1 != 1 || id2 != 2 {
		t.Errorf("ID 순서: want (1, 2), got (%d, %d)", id1, id2)
	}
}

// TestOpenFile_같은파일_중복방지 이미 열린 파일을 다시 열면 ErrFileAlreadyOpen을 반환하는지 검증한다 (AC2.4).
func TestOpenFile_같은파일_중복방지(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path := createTempFile(t, "test.md")

	id1, err := m.OpenFile(path)
	if err != nil {
		t.Fatalf("첫 번째 OpenFile() 오류: %v", err)
	}

	_, err = m.OpenFile(path)
	if err == nil {
		t.Fatal("두 번째 OpenFile()에서 에러를 기대했으나 nil 반환")
	}

	if !errors.Is(err, ErrFileAlreadyOpen) {
		t.Errorf("에러 타입: want ErrFileAlreadyOpen, got %v", err)
	}

	// 기존 윈도우 ID를 확인할 수 있어야 한다
	var alreadyOpen *FileAlreadyOpenError
	if errors.As(err, &alreadyOpen) {
		if alreadyOpen.WindowID != id1 {
			t.Errorf("기존 WindowID: want %d, got %d", id1, alreadyOpen.WindowID)
		}
	} else {
		t.Error("FileAlreadyOpenError로 변환할 수 없음")
	}

	// 윈도우 수는 여전히 1이어야 한다
	if m.Count() != 1 {
		t.Errorf("Count: want 1, got %d", m.Count())
	}
}

// TestOpenFile_최대윈도우_제한 최대 윈도우 수를 초과하면 ErrMaxWindowsReached를 반환하는지 검증한다 (AC2.6).
func TestOpenFile_최대윈도우_제한(t *testing.T) {
	t.Parallel()

	maxWin := 3
	m := newTestManager(WithMaxWindows(maxWin))

	// maxWin 개의 윈도우를 연다
	for i := 0; i < maxWin; i++ {
		path := createTempFile(t, filepath.Base(t.TempDir())+"_"+string(rune('a'+i))+".md")
		_, err := m.OpenFile(path)
		if err != nil {
			t.Fatalf("OpenFile(%d) 오류: %v", i+1, err)
		}
	}

	if m.Count() != maxWin {
		t.Errorf("Count: want %d, got %d", maxWin, m.Count())
	}

	// 초과 시도
	extraPath := createTempFile(t, "extra.md")
	_, err := m.OpenFile(extraPath)
	if !errors.Is(err, ErrMaxWindowsReached) {
		t.Errorf("에러 타입: want ErrMaxWindowsReached, got %v", err)
	}
}

// TestOpenFile_파일없음 존재하지 않는 파일을 열면 ErrFileNotFound를 반환하는지 검증한다 (AC2.7).
func TestOpenFile_파일없음(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	_, err := m.OpenFile("/nonexistent/path/file.md")
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("에러 타입: want ErrFileNotFound, got %v", err)
	}
}

// TestOpenFile_디렉토리 디렉토리를 열면 ErrFileNotFound를 반환하는지 검증한다.
func TestOpenFile_디렉토리(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	_, err := m.OpenFile(t.TempDir())
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("에러 타입: want ErrFileNotFound, got %v", err)
	}
}

// TestOpenFile_서버생성실패 서버 팩토리가 실패하면 에러를 반환하는지 검증한다.
func TestOpenFile_서버생성실패(t *testing.T) {
	t.Parallel()

	srvErr := errors.New("서버 생성 실패")
	m := NewWindowManager(
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			return nil, srvErr
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			return &testCloseable{}, nil
		}),
	)

	path := createTempFile(t, "test.md")
	_, err := m.OpenFile(path)
	if err == nil {
		t.Fatal("에러를 기대했으나 nil 반환")
	}
	if m.Count() != 0 {
		t.Errorf("실패 후 Count: want 0, got %d", m.Count())
	}
}

// TestOpenFile_감시자생성실패_서버정리 감시자 팩토리가 실패하면 서버도 정리되는지 검증한다.
func TestOpenFile_감시자생성실패_서버정리(t *testing.T) {
	t.Parallel()

	var createdServer *testServerHandle
	m := NewWindowManager(
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			createdServer = &testServerHandle{port: 8080}
			return createdServer, nil
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			return nil, errors.New("감시자 생성 실패")
		}),
	)

	path := createTempFile(t, "test.md")
	_, err := m.OpenFile(path)
	if err == nil {
		t.Fatal("에러를 기대했으나 nil 반환")
	}

	// 서버가 정리되었는지 확인
	if createdServer == nil {
		t.Fatal("서버가 생성되지 않음")
	}
	if createdServer.closeCalled.Load() != 1 {
		t.Errorf("서버 Close 호출 횟수: want 1, got %d", createdServer.closeCalled.Load())
	}
}

// TestCloseWindow_리소스정리 윈도우를 닫으면 리소스가 정리되는지 검증한다 (AC3.1).
func TestCloseWindow_리소스정리(t *testing.T) {
	t.Parallel()

	var lastSrv *testServerHandle
	var lastW *testCloseable

	m := NewWindowManager(
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			lastSrv = &testServerHandle{port: int(nextPort.Add(1))}
			return lastSrv, nil
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			lastW = &testCloseable{}
			return lastW, nil
		}),
	)

	path := createTempFile(t, "test.md")
	id, err := m.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() 오류: %v", err)
	}

	srvRef := lastSrv
	wRef := lastW

	err = m.CloseWindow(id)
	if err != nil {
		t.Fatalf("CloseWindow() 오류: %v", err)
	}

	if srvRef.closeCalled.Load() != 1 {
		t.Errorf("서버 Close 호출 횟수: want 1, got %d", srvRef.closeCalled.Load())
	}
	if wRef.closeCalled.Load() != 1 {
		t.Errorf("감시자 Close 호출 횟수: want 1, got %d", wRef.closeCalled.Load())
	}
	if m.Count() != 0 {
		t.Errorf("Count: want 0, got %d", m.Count())
	}
}

// TestCloseWindow_다른윈도우_영향없음 하나의 윈도우를 닫아도 다른 윈도우에 영향이 없는지 검증한다 (AC3.2).
func TestCloseWindow_다른윈도우_영향없음(t *testing.T) {
	t.Parallel()

	servers := make(map[int]*testServerHandle)
	watchers := make(map[int]*testCloseable)
	portCounter := atomic.Int32{}
	portCounter.Store(20000)

	m := NewWindowManager(
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			p := int(portCounter.Add(1))
			s := &testServerHandle{port: p}
			servers[p] = s
			return s, nil
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			p := int(portCounter.Load())
			w := &testCloseable{}
			watchers[p] = w
			return w, nil
		}),
	)

	path1 := createTempFile(t, "file1.md")
	path2 := createTempFile(t, "file2.md")

	id1, _ := m.OpenFile(path1)
	_, _ = m.OpenFile(path2)

	// 첫 번째 윈도우만 닫는다
	if err := m.CloseWindow(id1); err != nil {
		t.Fatalf("CloseWindow() 오류: %v", err)
	}

	// 두 번째 윈도우는 여전히 존재해야 한다
	if m.Count() != 1 {
		t.Errorf("Count: want 1, got %d", m.Count())
	}

	windows := m.GetWindows()
	if len(windows) != 1 {
		t.Fatalf("GetWindows 길이: want 1, got %d", len(windows))
	}
}

// TestCloseWindow_존재하지않는ID ErrWindowNotFound를 반환하는지 검증한다.
func TestCloseWindow_존재하지않는ID(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	err := m.CloseWindow(999)
	if !errors.Is(err, ErrWindowNotFound) {
		t.Errorf("에러 타입: want ErrWindowNotFound, got %v", err)
	}
}

// TestGetWindows_목록조회 열려있는 윈도우 목록을 올바르게 반환하는지 검증한다.
func TestGetWindows_목록조회(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path1 := createTempFile(t, "a.md")
	path2 := createTempFile(t, "b.md")

	m.OpenFile(path1)
	m.OpenFile(path2)

	windows := m.GetWindows()
	if len(windows) != 2 {
		t.Fatalf("GetWindows 길이: want 2, got %d", len(windows))
	}

	// ID 순 정렬 확인
	if windows[0].ID >= windows[1].ID {
		t.Errorf("ID 정렬: windows[0].ID(%d) >= windows[1].ID(%d)", windows[0].ID, windows[1].ID)
	}
}

// TestGetWindows_빈목록 윈도우가 없을 때 빈 슬라이스를 반환하는지 검증한다.
func TestGetWindows_빈목록(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	windows := m.GetWindows()
	if len(windows) != 0 {
		t.Errorf("GetWindows 길이: want 0, got %d", len(windows))
	}
}

// TestFindByPath_존재하는파일 열려있는 파일을 경로로 찾을 수 있는지 검증한다.
func TestFindByPath_존재하는파일(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path := createTempFile(t, "test.md")

	m.OpenFile(path)

	info, found := m.FindByPath(path)
	if !found {
		t.Fatal("FindByPath: want found, got not found")
	}
	absPath, _ := filepath.Abs(path)
	if info.FilePath != absPath {
		t.Errorf("FilePath: want %s, got %s", absPath, info.FilePath)
	}
}

// TestFindByPath_존재하지않는파일 열려있지 않은 파일에 대해 false를 반환하는지 검증한다.
func TestFindByPath_존재하지않는파일(t *testing.T) {
	t.Parallel()

	m := newTestManager()

	_, found := m.FindByPath("/not/open.md")
	if found {
		t.Error("FindByPath: want not found, got found")
	}
}

// TestShutdown_전체종료 Shutdown이 모든 윈도우를 닫는지 검증한다 (AC3.4).
func TestShutdown_전체종료(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path1 := createTempFile(t, "file1.md")
	path2 := createTempFile(t, "file2.md")
	path3 := createTempFile(t, "file3.md")

	m.OpenFile(path1)
	m.OpenFile(path2)
	m.OpenFile(path3)

	if m.Count() != 3 {
		t.Fatalf("Shutdown 전 Count: want 3, got %d", m.Count())
	}

	m.Shutdown()

	if m.Count() != 0 {
		t.Errorf("Shutdown 후 Count: want 0, got %d", m.Count())
	}
}

// TestOpenFile_콜백호출 OnWindowOpened 콜백이 호출되는지 검증한다.
func TestOpenFile_콜백호출(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path := createTempFile(t, "callback.md")

	var received atomic.Value
	done := make(chan struct{})
	m.OnWindowOpened(func(info WindowInfo) {
		received.Store(info)
		close(done)
	})

	_, err := m.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() 오류: %v", err)
	}

	select {
	case <-done:
		info := received.Load().(WindowInfo)
		if info.ID != 1 {
			t.Errorf("콜백 ID: want 1, got %d", info.ID)
		}
	case <-time.After(time.Second):
		t.Error("OnWindowOpened 콜백이 호출되지 않음")
	}
}

// TestCloseWindow_콜백호출 OnWindowClosed 콜백이 호출되는지 검증한다.
func TestCloseWindow_콜백호출(t *testing.T) {
	t.Parallel()

	m := newTestManager()
	path := createTempFile(t, "callback.md")

	id, _ := m.OpenFile(path)

	var received atomic.Value
	done := make(chan struct{})
	m.OnWindowClosed(func(info WindowInfo) {
		received.Store(info)
		close(done)
	})

	m.CloseWindow(id)

	select {
	case <-done:
		info := received.Load().(WindowInfo)
		if info.ID != id {
			t.Errorf("콜백 ID: want %d, got %d", id, info.ID)
		}
	case <-time.After(time.Second):
		t.Error("OnWindowClosed 콜백이 호출되지 않음")
	}
}

// TestOpenFile_동시접근 여러 고루틴에서 동시에 OpenFile을 호출해도 안전한지 검증한다.
func TestOpenFile_동시접근(t *testing.T) {
	t.Parallel()

	m := newTestManager(WithMaxWindows(100))

	const goroutines = 20
	paths := make([]string, goroutines)
	for i := 0; i < goroutines; i++ {
		paths[i] = createTempFile(t, filepath.Base(t.TempDir())+"_concurrent_"+string(rune('a'+i))+".md")
	}

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := m.OpenFile(paths[idx])
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("동시 OpenFile 오류: %v", err)
	}

	if m.Count() != goroutines {
		t.Errorf("Count: want %d, got %d", goroutines, m.Count())
	}
}

// TestOpenFile_최대윈도우10개_11번째거부 기본 최대 10개 윈도우 제한을 검증한다 (AC2.6).
func TestOpenFile_최대윈도우10개_11번째거부(t *testing.T) {
	t.Parallel()

	m := newTestManager() // 기본 최대 10개

	// 10개 윈도우를 연다
	for i := 0; i < 10; i++ {
		path := createTempFile(t, filepath.Base(t.TempDir())+"_max_"+string(rune('a'+i))+".md")
		_, err := m.OpenFile(path)
		if err != nil {
			t.Fatalf("OpenFile(%d) 오류: %v", i+1, err)
		}
	}

	if m.Count() != 10 {
		t.Fatalf("Count: want 10, got %d", m.Count())
	}

	// 11번째 시도
	path := createTempFile(t, "eleventh.md")
	_, err := m.OpenFile(path)
	if !errors.Is(err, ErrMaxWindowsReached) {
		t.Errorf("에러 타입: want ErrMaxWindowsReached, got %v", err)
	}
}

// TestOpenFile_독립감시자 각 윈도우가 독립적인 감시자를 가지는지 검증한다 (AC4.1).
func TestOpenFile_독립감시자(t *testing.T) {
	t.Parallel()

	watcherCount := atomic.Int32{}
	m := NewWindowManager(
		WithServerFactory(func(filePath string) (ServerHandle, error) {
			return &testServerHandle{port: int(nextPort.Add(1))}, nil
		}),
		WithWatcherFactory(func(filePath string) (Closeable, error) {
			watcherCount.Add(1)
			return &testCloseable{}, nil
		}),
	)

	path1 := createTempFile(t, "watch1.md")
	path2 := createTempFile(t, "watch2.md")

	m.OpenFile(path1)
	m.OpenFile(path2)

	if watcherCount.Load() != 2 {
		t.Errorf("생성된 감시자 수: want 2, got %d", watcherCount.Load())
	}
}

// TestOpenFile_팩토리미설정 팩토리가 없으면 에러를 반환하는지 검증한다.
func TestOpenFile_팩토리미설정(t *testing.T) {
	t.Parallel()

	m := NewWindowManager()
	path := createTempFile(t, "test.md")

	_, err := m.OpenFile(path)
	if err == nil {
		t.Fatal("팩토리 미설정 시 에러를 기대했으나 nil 반환")
	}
}
