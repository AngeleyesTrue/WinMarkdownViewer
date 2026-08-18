package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewWatcher_정상생성 유효한 파일 경로로 Watcher를 생성할 수 있는지 검증한다.
func TestNewWatcher_정상생성(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()
}

// TestNewWatcher_존재하지않는파일 존재하지 않는 파일 경로로 생성 시 에러를 반환하는지 검증한다.
func TestNewWatcher_존재하지않는파일(t *testing.T) {
	t.Parallel()

	_, err := NewWatcher("/nonexistent/path/file.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일에 대해 에러가 반환되어야 한다")
	}
}

// TestWatch_파일변경감지 파일 내용이 변경될 때 알림 채널로 이벤트가 전달되는지 검증한다.
func TestWatch_파일변경감지(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "watch.md")
	if err := os.WriteFile(tmpFile, []byte("# 초기 내용"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := w.Watch(ctx)

	// 파일 변경을 트리거한다
	time.Sleep(200 * time.Millisecond)
	if err := os.WriteFile(tmpFile, []byte("# 변경된 내용"), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	select {
	case <-ch:
		// 성공: 변경 알림을 수신했다
	case <-ctx.Done():
		t.Fatal("파일 변경 알림이 타임아웃 내에 수신되지 않았다")
	}
}

// TestWatch_디바운스 짧은 시간 내 연속 저장 시 하나의 이벤트만 발생하는지 검증한다.
func TestWatch_디바운스(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "debounce.md")
	if err := os.WriteFile(tmpFile, []byte("# 초기"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := w.Watch(ctx)

	// 짧은 간격으로 여러 번 파일을 쓴다 (디바운스 윈도우 100ms 이내)
	time.Sleep(200 * time.Millisecond)
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(tmpFile, []byte("# 변경 "+string(rune('A'+i))), 0644); err != nil {
			t.Fatalf("파일 쓰기 실패: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 디바운스 후 첫 번째 이벤트를 수신한다
	var eventCount atomic.Int32
	done := make(chan struct{})

	go func() {
		defer close(done)
		timer := time.NewTimer(500 * time.Millisecond)
		defer timer.Stop()
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return
				}
				eventCount.Add(1)
				// 타이머를 리셋하여 추가 이벤트를 기다린다
				timer.Reset(500 * time.Millisecond)
			case <-timer.C:
				return
			}
		}
	}()

	<-done
	count := eventCount.Load()
	// 디바운스로 인해 연속 저장이 1~2개의 이벤트로 병합되어야 한다
	if count == 0 {
		t.Fatal("최소 1개의 디바운스된 이벤트가 발생해야 한다")
	}
	if count > 2 {
		t.Errorf("디바운스 실패: %d개의 이벤트 발생 (최대 2개 기대)", count)
	}
}

// TestWatch_컨텍스트취소 컨텍스트 취소 시 채널이 닫히는지 검증한다.
func TestWatch_컨텍스트취소(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "cancel.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ch := w.Watch(ctx)

	// 컨텍스트를 취소한다
	cancel()

	// 채널이 닫혀야 한다
	select {
	case _, ok := <-ch:
		if ok {
			// 이벤트를 받은 경우 한 번 더 시도한다
			select {
			case _, ok := <-ch:
				if ok {
					t.Log("채널 닫힘 대기 중...")
				}
			case <-time.After(1 * time.Second):
				t.Fatal("컨텍스트 취소 후 채널이 닫혀야 한다")
			}
		}
		// ok == false: 채널이 정상적으로 닫혔다
	case <-time.After(2 * time.Second):
		t.Fatal("컨텍스트 취소 후 채널이 닫혀야 한다")
	}
}

// TestSwitchFile_정상전환 감시 중인 파일을 다른 파일로 전환할 수 있는지 검증한다.
func TestSwitchFile_정상전환(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file1 := filepath.Join(dir, "file1.md")
	file2 := filepath.Join(dir, "file2.md")

	if err := os.WriteFile(file1, []byte("# 파일1"), 0644); err != nil {
		t.Fatalf("파일1 생성 실패: %v", err)
	}
	if err := os.WriteFile(file2, []byte("# 파일2"), 0644); err != nil {
		t.Fatalf("파일2 생성 실패: %v", err)
	}

	w, err := NewWatcher(file1)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	if err := w.SwitchFile(file2); err != nil {
		t.Fatalf("SwitchFile() 오류: %v", err)
	}
}

// TestSwitchFile_존재하지않는파일 존재하지 않는 파일로 전환 시 에러를 반환하는지 검증한다.
func TestSwitchFile_존재하지않는파일(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	err = w.SwitchFile("/nonexistent/path/file.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일로 전환 시 에러가 반환되어야 한다")
	}
}

// TestClose_정리 Close 호출 시 리소스가 정리되는지 검증한다.
func TestClose_정리(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "close.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() 오류: %v", err)
	}

	// 이미 닫힌 후 다시 Close를 호출해도 패닉이 발생하지 않아야 한다
	_ = w.Close()
}

// TestSwitchFile_변경후감지 파일 전환 후 새 파일의 변경을 감지하는지 검증한다.
func TestSwitchFile_변경후감지(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file1 := filepath.Join(dir, "file1.md")
	file2 := filepath.Join(dir, "file2.md")

	if err := os.WriteFile(file1, []byte("# 파일1"), 0644); err != nil {
		t.Fatalf("파일1 생성 실패: %v", err)
	}
	if err := os.WriteFile(file2, []byte("# 파일2"), 0644); err != nil {
		t.Fatalf("파일2 생성 실패: %v", err)
	}

	w, err := NewWatcher(file1)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := w.Watch(ctx)

	// file2로 전환한다
	time.Sleep(200 * time.Millisecond)
	if err := w.SwitchFile(file2); err != nil {
		t.Fatalf("SwitchFile() 오류: %v", err)
	}

	// file2를 수정한다
	time.Sleep(200 * time.Millisecond)
	if err := os.WriteFile(file2, []byte("# 변경된 파일2"), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	select {
	case <-ch:
		// 성공: 전환된 파일의 변경 알림을 수신했다
	case <-ctx.Done():
		t.Fatal("전환된 파일의 변경 알림이 타임아웃 내에 수신되지 않았다")
	}
}

// TestWatch_fsnotify에러채널 fsnotify 에러 채널이 닫힐 때 안전하게 처리되는지 검증한다.
func TestWatch_Close후채널닫힘(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := w.Watch(ctx)

	// Close를 호출하면 fsnotify 채널이 닫히고 Watch 루프가 종료된다
	time.Sleep(100 * time.Millisecond)
	w.Close()

	// 채널이 닫혀야 한다
	select {
	case _, ok := <-ch:
		if ok {
			// 이벤트 하나 더 소비
			<-ch
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close 후 Watch 채널이 닫혀야 한다")
	}
}

// TestWatch_파일삭제후컨텍스트취소 파일 삭제 후 복구 전 컨텍스트가 취소되면 안전하게 종료되는지 검증한다.
func TestWatch_파일삭제후컨텍스트취소(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "delete_cancel.md")
	if err := os.WriteFile(tmpFile, []byte("# 테스트"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ch := w.Watch(ctx)

	// 파일을 삭제한다
	time.Sleep(200 * time.Millisecond)
	if err := os.Remove(tmpFile); err != nil {
		t.Fatalf("파일 삭제 실패: %v", err)
	}

	// 복구하지 않고 컨텍스트를 취소한다 (waitForRecovery의 ctx.Done 경로)
	time.Sleep(200 * time.Millisecond)
	cancel()

	// 채널이 닫혀야 한다
	select {
	case _, ok := <-ch:
		if ok {
			<-ch // 추가 이벤트 소비
		}
	case <-time.After(3 * time.Second):
		t.Fatal("컨텍스트 취소 후 채널이 닫혀야 한다")
	}
}

// TestWatch_파일삭제후복구 파일이 삭제된 후 다시 생성되면 감시를 재개하는지 검증한다.
func TestWatch_파일삭제후복구(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "recover.md")
	if err := os.WriteFile(tmpFile, []byte("# 초기"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	w, err := NewWatcher(tmpFile)
	if err != nil {
		t.Fatalf("NewWatcher() 오류: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := w.Watch(ctx)

	// 파일을 삭제한다
	time.Sleep(200 * time.Millisecond)
	if err := os.Remove(tmpFile); err != nil {
		t.Fatalf("파일 삭제 실패: %v", err)
	}

	// 파일을 다시 생성한다
	time.Sleep(300 * time.Millisecond)
	if err := os.WriteFile(tmpFile, []byte("# 복구된 내용"), 0644); err != nil {
		t.Fatalf("파일 재생성 실패: %v", err)
	}

	// 파일 복구 후 변경 알림을 수신해야 한다
	select {
	case <-ch:
		// 성공: 파일 복구 후 알림을 수신했다
	case <-ctx.Done():
		t.Fatal("파일 삭제 후 복구 알림이 타임아웃 내에 수신되지 않았다")
	}
}
