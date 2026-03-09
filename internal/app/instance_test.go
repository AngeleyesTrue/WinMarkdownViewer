package app_test

import (
	"testing"

	"github.com/AngeleyesTrue/WinMarkdownViewer/internal/app"
)

// TestTryLock_첫번째인스턴스성공 은 첫 번째 TryLock 호출이 성공하는지 검증한다.
func TestTryLock_첫번째인스턴스성공(t *testing.T) {
	inst := app.NewInstanceLock()
	t.Cleanup(func() {
		_ = inst.Unlock()
	})

	got, err := inst.TryLock()
	if err != nil {
		t.Fatalf("TryLock() 오류: %v", err)
	}
	if !got {
		t.Error("첫 번째 TryLock() = false, want true")
	}
}

// TestTryLock_두번째인스턴스실패 는 동일 프로세스에서 두 번째 TryLock이
// 이미 실행 중임을 감지하는지 검증한다.
func TestTryLock_두번째인스턴스실패(t *testing.T) {
	inst1 := app.NewInstanceLock()
	t.Cleanup(func() {
		_ = inst1.Unlock()
	})

	got1, err := inst1.TryLock()
	if err != nil {
		t.Fatalf("첫 번째 TryLock() 오류: %v", err)
	}
	if !got1 {
		t.Fatal("첫 번째 TryLock() = false, want true")
	}

	// 같은 뮤텍스 이름으로 두 번째 인스턴스 시도
	inst2 := app.NewInstanceLock()
	t.Cleanup(func() {
		_ = inst2.Unlock()
	})

	got2, err := inst2.TryLock()
	if err != nil {
		t.Fatalf("두 번째 TryLock() 오류: %v", err)
	}
	if got2 {
		t.Error("두 번째 TryLock() = true, want false (이미 실행 중)")
	}
}

// TestUnlock후재TryLock성공 은 Unlock 후 다시 TryLock이 성공하는지 검증한다 (ACC-012).
func TestUnlock후재TryLock성공(t *testing.T) {
	inst1 := app.NewInstanceLock()

	got, err := inst1.TryLock()
	if err != nil {
		t.Fatalf("첫 번째 TryLock() 오류: %v", err)
	}
	if !got {
		t.Fatal("첫 번째 TryLock() = false, want true")
	}

	// Unlock
	if err := inst1.Unlock(); err != nil {
		t.Fatalf("Unlock() 오류: %v", err)
	}

	// 새 인스턴스로 다시 TryLock
	inst2 := app.NewInstanceLock()
	t.Cleanup(func() {
		_ = inst2.Unlock()
	})

	got2, err := inst2.TryLock()
	if err != nil {
		t.Fatalf("재 TryLock() 오류: %v", err)
	}
	if !got2 {
		t.Error("Unlock 후 재 TryLock() = false, want true")
	}
}

// TestUnlock_잠금없이호출 은 TryLock을 호출하지 않은 상태에서 Unlock이
// 에러 없이 반환되는지 검증한다 (handle == 0 분기).
func TestUnlock_잠금없이호출(t *testing.T) {
	inst := app.NewInstanceLock()

	// TryLock 없이 바로 Unlock - handle이 0이므로 nil 반환
	err := inst.Unlock()
	if err != nil {
		t.Errorf("잠금 없이 Unlock() 오류: %v", err)
	}
}

// TestUnlock_중복호출 은 Unlock을 두 번 호출해도 안전한지 검증한다.
func TestUnlock_중복호출(t *testing.T) {
	inst := app.NewInstanceLock()

	got, err := inst.TryLock()
	if err != nil {
		t.Fatalf("TryLock() 오류: %v", err)
	}
	if !got {
		t.Fatal("TryLock() = false, want true")
	}

	// 첫 번째 Unlock
	if err := inst.Unlock(); err != nil {
		t.Fatalf("첫 번째 Unlock() 오류: %v", err)
	}

	// 두 번째 Unlock - handle이 이미 0이므로 nil 반환
	err = inst.Unlock()
	if err != nil {
		t.Errorf("두 번째 Unlock() 오류: %v", err)
	}
}
