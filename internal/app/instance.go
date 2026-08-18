package app

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

// InstanceLock 은 Named Mutex를 통한 단일 인스턴스 잠금을 관리한다.
type InstanceLock struct {
	handle windows.Handle
}

// NewInstanceLock 은 새 InstanceLock을 생성한다.
func NewInstanceLock() *InstanceLock {
	return &InstanceLock{}
}

// TryLock 은 Named Mutex 획득을 시도한다.
// 성공 시 true (첫 번째 인스턴스), 실패 시 false (이미 실행 중)를 반환한다.
func (il *InstanceLock) TryLock() (bool, error) {
	name, err := syscall.UTF16PtrFromString(DefaultMutexName)
	if err != nil {
		return false, fmt.Errorf("뮤텍스 이름 변환 실패: %w", err)
	}

	// CreateMutexW 호출 (initialOwner = false로 생성 후 즉시 소유하지 않음)
	handle, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			// 뮤텍스가 이미 존재 = 다른 인스턴스가 실행 중
			// handle은 유효하지만 소유권 없음, 닫아야 함
			if handle != 0 {
				windows.CloseHandle(handle)
			}
			return false, nil
		}
		return false, fmt.Errorf("뮤텍스 생성 실패: %w", err)
	}

	il.handle = handle
	return true, nil
}

// Unlock 은 Mutex를 해제한다.
func (il *InstanceLock) Unlock() error {
	if il.handle == 0 {
		return nil
	}

	// ReleaseMutex로 소유권 해제
	err := windows.ReleaseMutex(il.handle)
	if err != nil {
		// ReleaseMutex 실패는 소유하지 않은 경우일 수 있음, 핸들은 닫아야 함
		_ = err
	}

	if err := windows.CloseHandle(il.handle); err != nil {
		return fmt.Errorf("뮤텍스 핸들 닫기 실패: %w", err)
	}
	il.handle = 0
	return nil
}

