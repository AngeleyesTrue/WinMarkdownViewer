package app_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/app"
	"golang.org/x/sys/windows"
)

// TestPipe_서버클라이언트통신 은 ListenPipe 서버와 SendPath 클라이언트 간
// 정상적인 라운드트립 통신을 검증한다.
func TestPipe_서버클라이언트통신(t *testing.T) {
	// 임시 파일 생성 (파일 존재 검증용)
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var received string
	var mu sync.Mutex
	done := make(chan struct{})

	handler := func(filePath string) {
		mu.Lock()
		received = filePath
		mu.Unlock()
		close(done)
	}

	// 서버 시작
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.ListenPipe(ctx, handler)
	}()

	// 서버 준비 대기
	time.Sleep(200 * time.Millisecond)

	// 클라이언트로 파일 경로 전송
	if err := app.SendPath(tmpFile); err != nil {
		t.Fatalf("SendPath() 오류: %v", err)
	}

	// 수신 대기
	select {
	case <-done:
		mu.Lock()
		if received != tmpFile {
			t.Errorf("수신된 경로 = %q, want %q", received, tmpFile)
		}
		mu.Unlock()
	case <-time.After(3 * time.Second):
		t.Fatal("핸들러가 호출되지 않았다 (타임아웃)")
	}

	cancel()
}

// TestPipe_순차적다중연결 은 여러 번의 순차적 연결이 모두 성공하는지 검증한다.
func TestPipe_순차적다중연결(t *testing.T) {
	tmpDir := t.TempDir()
	files := []string{"a.md", "b.md", "c.md"}
	for _, name := range files {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte("# "+name), 0644); err != nil {
			t.Fatalf("테스트 파일 생성 실패: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var receivedPaths []string
	var mu sync.Mutex
	count := 0
	allDone := make(chan struct{})

	handler := func(filePath string) {
		mu.Lock()
		receivedPaths = append(receivedPaths, filePath)
		count++
		if count == len(files) {
			close(allDone)
		}
		mu.Unlock()
	}

	// 서버 시작
	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 순차 전송
	for _, name := range files {
		p := filepath.Join(tmpDir, name)
		if err := app.SendPath(p); err != nil {
			t.Fatalf("SendPath(%s) 오류: %v", name, err)
		}
		// Named Pipe는 연결-전송-종료 사이클마다 간격 필요
		time.Sleep(100 * time.Millisecond)
	}

	select {
	case <-allDone:
		mu.Lock()
		if len(receivedPaths) != len(files) {
			t.Errorf("수신된 경로 수 = %d, want %d", len(receivedPaths), len(files))
		}
		mu.Unlock()
	case <-time.After(5 * time.Second):
		mu.Lock()
		t.Fatalf("모든 파일을 수신하지 못함: %d/%d", len(receivedPaths), len(files))
		mu.Unlock()
	}

	cancel()
}

// TestPipe_컨텍스트취소로서버중단 은 context 취소 시 서버가 정상 종료하는지 검증한다.
func TestPipe_컨텍스트취소로서버중단(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- app.ListenPipe(ctx, func(string) {})
	}()
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case err := <-serverDone:
		// 컨텍스트 취소로 인한 종료는 정상
		if err != nil && err != context.Canceled {
			t.Logf("서버 종료 에러 (허용됨): %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("서버가 context 취소 후 종료되지 않았다 (타임아웃)")
	}
}

// TestPipe_동시연결 은 여러 클라이언트가 동시에 연결해도 처리되는지 검증한다 (ACC-017).
func TestPipe_동시연결(t *testing.T) {
	tmpDir := t.TempDir()
	numClients := 5
	for i := 0; i < numClients; i++ {
		p := filepath.Join(tmpDir, fmt.Sprintf("concurrent_%d.md", i))
		if err := os.WriteFile(p, []byte("# Concurrent"), 0644); err != nil {
			t.Fatalf("파일 생성 실패: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var receivedCount int
	var mu sync.Mutex
	allDone := make(chan struct{})

	handler := func(filePath string) {
		mu.Lock()
		receivedCount++
		if receivedCount == numClients {
			close(allDone)
		}
		mu.Unlock()
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 동시 전송
	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p := filepath.Join(tmpDir, fmt.Sprintf("concurrent_%d.md", idx))
			// 동시 연결 시 Named Pipe 인스턴스가 1개이므로 순차적으로 처리됨
			// 재시도 로직 포함
			for retry := 0; retry < 5; retry++ {
				if err := app.SendPath(p); err == nil {
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			t.Errorf("SendPath 실패 (idx=%d)", idx)
		}(i)
	}
	wg.Wait()

	select {
	case <-allDone:
		// 성공
	case <-time.After(5 * time.Second):
		mu.Lock()
		t.Fatalf("동시 연결 처리 실패: %d/%d", receivedCount, numClients)
		mu.Unlock()
	}

	cancel()
}

// TestPipe_빈문자열무시 는 빈 문자열이 전송되면 핸들러가 호출되지 않는지 검증한다 (ACC-011).
func TestPipe_빈문자열무시(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handlerCalled := false
	handler := func(filePath string) {
		handlerCalled = true
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 빈 문자열 전송 시도 - SendPath에서 거부되어야 함
	err := app.SendPath("")
	if err == nil {
		t.Error("빈 문자열 SendPath는 에러를 반환해야 한다")
	}

	time.Sleep(200 * time.Millisecond)

	if handlerCalled {
		t.Error("빈 문자열에 대해 핸들러가 호출되었다")
	}

	cancel()
}

// TestPipe_존재하지않는파일거부 는 존재하지 않는 파일 경로가 서버에서 무시되는지 검증한다.
func TestPipe_존재하지않는파일거부(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handlerCalled := false
	handler := func(filePath string) {
		handlerCalled = true
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 존재하지 않는 파일 전송 시도 - SendPath에서 거부되어야 함
	err := app.SendPath(`C:\nonexistent\path\file_99999.md`)
	if err == nil {
		t.Error("존재하지 않는 파일 SendPath는 에러를 반환해야 한다")
	}

	time.Sleep(200 * time.Millisecond)

	if handlerCalled {
		t.Error("존재하지 않는 파일에 대해 핸들러가 호출되었다")
	}

	cancel()
}

// TestSendPath_빈문자열에러메시지 는 빈 문자열 전송 시 적절한 에러 메시지를 반환하는지 검증한다.
func TestSendPath_빈문자열에러메시지(t *testing.T) {
	err := app.SendPath("")
	if err == nil {
		t.Fatal("빈 문자열 SendPath()는 에러를 반환해야 한다")
	}
	if !strings.Contains(err.Error(), "비어 있습니다") {
		t.Errorf("에러 메시지에 '비어 있습니다'가 포함되어야 함: %v", err)
	}
}

// TestSendPath_존재하지않는파일에러메시지 는 존재하지 않는 파일 전송 시
// 적절한 에러 메시지를 반환하는지 검증한다.
func TestSendPath_존재하지않는파일에러메시지(t *testing.T) {
	err := app.SendPath(`C:\nonexistent_path_99999\file.md`)
	if err == nil {
		t.Fatal("존재하지 않는 파일 SendPath()는 에러를 반환해야 한다")
	}
	if !strings.Contains(err.Error(), "찾을 수 없습니다") {
		t.Errorf("에러 메시지에 '찾을 수 없습니다'가 포함되어야 함: %v", err)
	}
}

// TestListenPipe_즉시취소 는 이미 취소된 context로 ListenPipe를 호출하면
// 즉시 종료되는지 검증한다.
func TestListenPipe_즉시취소(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	done := make(chan error, 1)
	go func() {
		done <- app.ListenPipe(ctx, func(string) {})
	}()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("즉시 취소된 context에서 err = %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("이미 취소된 context에서 ListenPipe가 종료되지 않았다")
	}
}

// rawWriteToPipe 는 SendPath의 유효성 검사를 우회하여 Named Pipe에
// 원시 바이트를 직접 전송하는 헬퍼 함수이다.
func rawWriteToPipe(t *testing.T, data []byte) error {
	t.Helper()

	pipeName, err := syscall.UTF16PtrFromString(app.DefaultPipeName)
	if err != nil {
		return err
	}

	var handle windows.Handle
	for retry := 0; retry < 10; retry++ {
		handle, err = windows.CreateFile(
			pipeName,
			windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	var bytesWritten uint32
	return windows.WriteFile(handle, data, &bytesWritten, nil)
}

// TestPipe_존재하지않는파일서버측무시 는 유효성 검사를 우회하여 존재하지 않는 파일 경로를
// 직접 파이프에 전송했을 때 서버가 핸들러를 호출하지 않는지 검증한다 (ACC-011).
func TestPipe_존재하지않는파일서버측무시(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handlerCalled := false
	handler := func(filePath string) {
		handlerCalled = true
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 존재하지 않는 파일 경로를 직접 파이프에 전송
	if err := rawWriteToPipe(t, []byte(`C:\nonexistent_99999\file.md`)); err != nil {
		t.Fatalf("rawWriteToPipe 오류: %v", err)
	}

	// 핸들러가 호출되지 않아야 함
	time.Sleep(500 * time.Millisecond)

	if handlerCalled {
		t.Error("존재하지 않는 파일에 대해 서버 측 핸들러가 호출되었다")
	}

	cancel()
}

// TestPipe_널바이트경로서버측무시 는 널 바이트가 포함된 경로를 직접 파이프에 전송했을 때
// 서버가 널 바이트 이후를 무시하고 유효한 경로 부분만 처리하는지 검증한다.
func TestPipe_널바이트경로서버측무시(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handlerCalled := false
	handler := func(filePath string) {
		handlerCalled = true
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 널 바이트가 포함된 경로를 직접 파이프에 전송 (존재하지 않는 파일)
	if err := rawWriteToPipe(t, []byte("C:\\test\x00evil\\file.md")); err != nil {
		t.Fatalf("rawWriteToPipe 오류: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if handlerCalled {
		t.Error("널 바이트가 포함된 존재하지 않는 경로에 대해 핸들러가 호출되었다")
	}

	cancel()
}

// TestPipe_상대경로서버측무시 는 상대 경로를 직접 파이프에 전송했을 때
// 서버가 핸들러를 호출하지 않는지 검증한다.
func TestPipe_상대경로서버측무시(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handlerCalled := false
	handler := func(filePath string) {
		handlerCalled = true
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 상대 경로를 직접 파이프에 전송
	if err := rawWriteToPipe(t, []byte("relative/path/file.md")); err != nil {
		t.Fatalf("rawWriteToPipe 오류: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if handlerCalled {
		t.Error("상대 경로에 대해 핸들러가 호출되었다")
	}

	cancel()
}

// TestPipe_유효한파일서버측전달 은 직접 파이프에 유효한 파일 경로를 전송했을 때
// 서버가 핸들러를 정상적으로 호출하는지 검증한다.
func TestPipe_유효한파일서버측전달(t *testing.T) {
	// 테스트용 임시 파일 생성
	tmpFile := filepath.Join(t.TempDir(), "raw_test.md")
	if err := os.WriteFile(tmpFile, []byte("# Raw Test"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var received string
	var mu sync.Mutex
	done := make(chan struct{})

	handler := func(filePath string) {
		mu.Lock()
		received = filePath
		mu.Unlock()
		close(done)
	}

	go func() {
		_ = app.ListenPipe(ctx, handler)
	}()
	time.Sleep(200 * time.Millisecond)

	// 유효한 파일 경로를 직접 파이프에 전송
	if err := rawWriteToPipe(t, []byte(tmpFile)); err != nil {
		t.Fatalf("rawWriteToPipe 오류: %v", err)
	}

	select {
	case <-done:
		mu.Lock()
		if received != tmpFile {
			t.Errorf("수신된 경로 = %q, want %q", received, tmpFile)
		}
		mu.Unlock()
	case <-time.After(3 * time.Second):
		t.Fatal("핸들러가 호출되지 않았다 (타임아웃)")
	}

	cancel()
}
