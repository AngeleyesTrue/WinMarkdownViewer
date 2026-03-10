// Package window 는 다중 윈도우 관리를 위한 WindowManager와 Window를 제공한다.
package window

import "errors"

// 윈도우 관리에서 발생할 수 있는 에러 타입들이다.
var (
	// ErrMaxWindowsReached 는 최대 윈도우 수에 도달했을 때 반환된다.
	ErrMaxWindowsReached = errors.New("최대 윈도우 수에 도달했습니다")

	// ErrFileNotFound 는 파일이 존재하지 않을 때 반환된다.
	ErrFileNotFound = errors.New("파일을 찾을 수 없습니다")

	// ErrFileNotReadable 는 파일을 읽을 수 없을 때 반환된다.
	ErrFileNotReadable = errors.New("파일을 읽을 수 없습니다")

	// ErrWindowNotFound 는 지정된 ID의 윈도우가 없을 때 반환된다.
	ErrWindowNotFound = errors.New("윈도우를 찾을 수 없습니다")

	// ErrFileAlreadyOpen 은 이미 열려있는 파일을 다시 열려고 할 때 반환된다.
	ErrFileAlreadyOpen = errors.New("파일이 이미 열려있습니다")
)

// FileAlreadyOpenError 는 이미 열린 파일에 대한 에러이다.
// 기존 윈도우 ID를 함께 제공한다.
type FileAlreadyOpenError struct {
	WindowID int
}

// Error 는 error 인터페이스를 구현한다.
func (e *FileAlreadyOpenError) Error() string {
	return ErrFileAlreadyOpen.Error()
}

// Is 는 errors.Is 비교를 지원한다.
func (e *FileAlreadyOpenError) Is(target error) bool {
	return target == ErrFileAlreadyOpen
}
