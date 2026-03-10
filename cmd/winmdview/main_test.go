package main

import (
	"errors"
	"sync"
	"testing"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/window"
)

// TestParseFlagsRegister --register 플래그 파싱을 검증한다.
func TestParseFlagsRegister(t *testing.T) {
	flags := parseFlags([]string{"--register"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	if flags.unregister {
		t.Error("--unregister 플래그가 false여야 한다")
	}
	if flags.filePath != "" {
		t.Errorf("filePath가 비어있어야 한다: got %q", flags.filePath)
	}
}

// TestParseFlagsUnregister --unregister 플래그 파싱을 검증한다.
func TestParseFlagsUnregister(t *testing.T) {
	flags := parseFlags([]string{"--unregister"})

	if !flags.unregister {
		t.Error("--unregister 플래그가 true여야 한다")
	}
	if flags.register {
		t.Error("--register 플래그가 false여야 한다")
	}
}

// TestParseFlagsSetDefault --set-default 플래그 파싱을 검증한다.
func TestParseFlagsSetDefault(t *testing.T) {
	flags := parseFlags([]string{"--set-default"})

	if !flags.setDefault {
		t.Error("--set-default 플래그가 true여야 한다")
	}
}

// TestParseFlagsFilePath 파일 경로 인자 파싱을 검증한다.
func TestParseFlagsFilePath(t *testing.T) {
	flags := parseFlags([]string{"test.md"})

	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
	if flags.register {
		t.Error("register 플래그가 false여야 한다")
	}
}

// TestParseFlagsRegisterWithFilePath --register와 파일 경로가 함께 전달된 경우를 검증한다 (ACC-014).
// 등록만 수행하고 파일을 열지 않아야 한다.
func TestParseFlagsRegisterWithFilePath(t *testing.T) {
	flags := parseFlags([]string{"--register", "test.md"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	// filePath도 파싱되지만, register가 true이면 파일 열기는 수행하지 않는다
	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
}

// TestParseFlagsNoArgs 인자 없이 호출된 경우를 검증한다.
func TestParseFlagsNoArgs(t *testing.T) {
	flags := parseFlags([]string{})

	if flags.register || flags.unregister || flags.setDefault {
		t.Error("플래그가 모두 false여야 한다")
	}
	if flags.filePath != "" {
		t.Errorf("filePath가 비어있어야 한다: got %q", flags.filePath)
	}
}

// TestParseFlagsMultipleFlags 여러 플래그 조합을 검증한다.
func TestParseFlagsMultipleFlags(t *testing.T) {
	flags := parseFlags([]string{"--register", "--set-default"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	if !flags.setDefault {
		t.Error("--set-default 플래그가 true여야 한다")
	}
}

// TestParseFlagsUnknownFlagIgnored 알 수 없는 플래그는 무시된다.
func TestParseFlagsUnknownFlagIgnored(t *testing.T) {
	flags := parseFlags([]string{"--unknown", "test.md"})

	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
}

// TestParseFlagsWindowsPath 윈도우 절대 경로를 올바르게 파싱하는지 검증한다.
func TestParseFlagsWindowsPath(t *testing.T) {
	flags := parseFlags([]string{`C:\Users\test\doc.md`})

	if flags.filePath != `C:\Users\test\doc.md` {
		t.Errorf("filePath: want %q, got %q", `C:\Users\test\doc.md`, flags.filePath)
	}
}

// --- Phase 4: handlePipeMessage 헬퍼 함수 테스트 ---

// mockWindowOpener 는 handlePipeMessage 테스트를 위한 WindowManager 동작을 모의한다.
type mockWindowOpener struct {
	openFileFn func(path string) (int, error)
}

func (m *mockWindowOpener) OpenFile(path string) (int, error) {
	return m.openFileFn(path)
}

// TestHandlePipeMessage_새파일열기 새 파일이 파이프로 수신되면 WindowManager.OpenFile이 호출되는지 검증한다.
func TestHandlePipeMessage_새파일열기(t *testing.T) {
	t.Parallel()

	var openedPath string
	opener := &mockWindowOpener{
		openFileFn: func(path string) (int, error) {
			openedPath = path
			return 1, nil
		},
	}

	var activatedID int
	result := handlePipeMessage(opener, `C:\docs\test.md`, func(id int) {
		activatedID = id
	})

	if result != pipeResultNewWindow {
		t.Errorf("result: want %d (pipeResultNewWindow), got %d", pipeResultNewWindow, result)
	}
	if openedPath != `C:\docs\test.md` {
		t.Errorf("OpenFile 호출 경로: want %q, got %q", `C:\docs\test.md`, openedPath)
	}
	if activatedID != 0 {
		t.Errorf("활성화 콜백이 호출되지 않아야 함: got id=%d", activatedID)
	}
}

// TestHandlePipeMessage_이미열린파일_활성화 이미 열린 파일이면 기존 윈도우를 활성화하는지 검증한다.
func TestHandlePipeMessage_이미열린파일_활성화(t *testing.T) {
	t.Parallel()

	opener := &mockWindowOpener{
		openFileFn: func(path string) (int, error) {
			return 0, &window.FileAlreadyOpenError{WindowID: 42}
		},
	}

	var activatedID int
	result := handlePipeMessage(opener, `C:\docs\test.md`, func(id int) {
		activatedID = id
	})

	if result != pipeResultActivated {
		t.Errorf("result: want %d (pipeResultActivated), got %d", pipeResultActivated, result)
	}
	if activatedID != 42 {
		t.Errorf("활성화된 윈도우 ID: want 42, got %d", activatedID)
	}
}

// TestHandlePipeMessage_최대윈도우초과 최대 윈도우 수 초과 시 에러 결과를 반환하는지 검증한다.
func TestHandlePipeMessage_최대윈도우초과(t *testing.T) {
	t.Parallel()

	opener := &mockWindowOpener{
		openFileFn: func(path string) (int, error) {
			return 0, window.ErrMaxWindowsReached
		},
	}

	result := handlePipeMessage(opener, `C:\docs\test.md`, func(id int) {})

	if result != pipeResultError {
		t.Errorf("result: want %d (pipeResultError), got %d", pipeResultError, result)
	}
}

// TestHandlePipeMessage_파일없음 파일이 없을 때 에러 결과를 반환하는지 검증한다.
func TestHandlePipeMessage_파일없음(t *testing.T) {
	t.Parallel()

	opener := &mockWindowOpener{
		openFileFn: func(path string) (int, error) {
			return 0, window.ErrFileNotFound
		},
	}

	result := handlePipeMessage(opener, `C:\nonexistent.md`, func(id int) {})

	if result != pipeResultError {
		t.Errorf("result: want %d (pipeResultError), got %d", pipeResultError, result)
	}
}

// TestHandlePipeMessage_기타에러 기타 에러도 에러 결과를 반환하는지 검증한다.
func TestHandlePipeMessage_기타에러(t *testing.T) {
	t.Parallel()

	opener := &mockWindowOpener{
		openFileFn: func(path string) (int, error) {
			return 0, errors.New("알 수 없는 오류")
		},
	}

	result := handlePipeMessage(opener, `C:\docs\test.md`, func(id int) {})

	if result != pipeResultError {
		t.Errorf("result: want %d (pipeResultError), got %d", pipeResultError, result)
	}
}

// --- Phase 5: windowTracker 테스트 ---

// TestWindowTracker_AddAndGet 윈도우 항목 추가 및 조회를 검증한다.
func TestWindowTracker_AddAndGet(t *testing.T) {
	t.Parallel()

	tracker := newWindowTracker()
	tracker.add(1, nil, 0) // viewer와 hwnd는 nil/0으로 테스트

	entry, ok := tracker.get(1)
	if !ok {
		t.Fatal("windowID=1 항목이 존재해야 한다")
	}
	if entry.viewer != nil {
		t.Error("viewer는 nil이어야 한다 (테스트용)")
	}
}

// TestWindowTracker_GetNotFound 존재하지 않는 윈도우 조회 시 false를 반환하는지 검증한다.
func TestWindowTracker_GetNotFound(t *testing.T) {
	t.Parallel()

	tracker := newWindowTracker()
	_, ok := tracker.get(999)
	if ok {
		t.Error("존재하지 않는 windowID에 대해 false를 반환해야 한다")
	}
}

// TestWindowTracker_Remove 윈도우 항목 제거와 allClosed 채널 동작을 검증한다.
func TestWindowTracker_Remove(t *testing.T) {
	t.Parallel()

	tracker := newWindowTracker()
	tracker.add(1, nil, 0)
	tracker.add(2, nil, 0)

	// 첫 번째 윈도우 제거 -> 아직 allClosed가 닫히지 않아야 한다
	tracker.remove(1)
	select {
	case <-tracker.allClosed:
		t.Fatal("아직 윈도우가 남아있으므로 allClosed가 닫히면 안 된다")
	default:
		// 정상
	}

	// 두 번째 윈도우 제거 -> allClosed가 닫혀야 한다
	tracker.remove(2)
	select {
	case <-tracker.allClosed:
		// 정상
	default:
		t.Fatal("모든 윈도우가 닫힌 후 allClosed 채널이 닫혀야 한다")
	}
}

// TestWindowTracker_Wait WaitGroup 기반 대기가 올바르게 동작하는지 검증한다.
func TestWindowTracker_Wait(t *testing.T) {
	t.Parallel()

	tracker := newWindowTracker()
	tracker.add(1, nil, 0)

	done := make(chan struct{})
	go func() {
		tracker.wait()
		close(done)
	}()

	// 아직 대기 중이어야 한다
	select {
	case <-done:
		t.Fatal("윈도우가 아직 열려 있으므로 wait()가 반환되면 안 된다")
	default:
	}

	tracker.remove(1)

	// wait()가 반환되어야 한다
	<-done
}

// TestWindowTracker_ConcurrentAccess 동시 접근 시 데이터 레이스가 없는지 검증한다.
func TestWindowTracker_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tracker := newWindowTracker()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			tracker.add(id, nil, 0)
			tracker.get(id)
			tracker.remove(id)
		}()
	}
	wg.Wait()
}

