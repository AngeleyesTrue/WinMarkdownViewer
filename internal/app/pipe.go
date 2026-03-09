package app

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// @MX:NOTE: [AUTO] ListenPipe는 context 취소 시 CancelIoEx로 대기 중인 ConnectNamedPipe를 중단한다.

// maxConsecutiveErrors 는 연속 에러 발생 시 최대 허용 횟수이다.
const maxConsecutiveErrors = 10

// ListenPipe 는 Named Pipe 서버를 시작하여 클라이언트로부터 파일 경로를 수신한다.
// ctx 취소 시 서버를 종료한다.
// handler 는 유효한 파일 경로를 수신할 때마다 호출된다.
// 유효하지 않은 데이터 (빈 문자열, 존재하지 않는 파일)는 무시한다 (ACC-011).
func ListenPipe(ctx context.Context, handler func(filePath string)) error {
	var backoff time.Duration = 100 * time.Millisecond
	var consecutiveErrors int

	for {
		// context 확인
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := listenOnce(ctx, handler); err != nil {
			// context 취소인 경우 정상 종료
			if ctx.Err() != nil {
				return ctx.Err()
			}

			consecutiveErrors++
			log.Printf("파이프 리슨 에러 (%d/%d): %v", consecutiveErrors, maxConsecutiveErrors, err)

			if consecutiveErrors >= maxConsecutiveErrors {
				return fmt.Errorf("파이프 복구 실패 (연속 %d회 에러): %w", consecutiveErrors, err)
			}

			// 지수 백오프 + 지터
			jitter := time.Duration(rand.IntN(int(backoff / 2)))
			time.Sleep(backoff + jitter)
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		// 성공 시 리셋
		consecutiveErrors = 0
		backoff = 100 * time.Millisecond
	}
}

// listenOnce 는 하나의 클라이언트 연결을 처리한다.
func listenOnce(ctx context.Context, handler func(filePath string)) error {
	pipeName, err := syscall.UTF16PtrFromString(DefaultPipeName)
	if err != nil {
		return fmt.Errorf("파이프 이름 변환 실패: %w", err)
	}

	// Named Pipe 생성
	handle, err := windows.CreateNamedPipe(
		pipeName,
		windows.PIPE_ACCESS_INBOUND|windows.FILE_FLAG_OVERLAPPED,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
		1,                   // 최대 인스턴스 수
		MaxPipeMessageSize,  // 출력 버퍼 크기
		MaxPipeMessageSize,  // 입력 버퍼 크기
		PipeTimeoutMs,         // 기본 타임아웃
		nil,                 // 보안 속성
	)
	if err != nil {
		return fmt.Errorf("Named Pipe 생성 실패: %w", err)
	}
	defer windows.CloseHandle(handle)

	// Overlapped I/O를 위한 이벤트 생성
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fmt.Errorf("이벤트 생성 실패: %w", err)
	}
	defer windows.CloseHandle(event)

	// Overlapped 구조체 준비
	overlapped := &windows.Overlapped{
		HEvent: event,
	}

	// 비동기 ConnectNamedPipe 호출
	err = windows.ConnectNamedPipe(handle, overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING && err != windows.ERROR_PIPE_CONNECTED {
		return fmt.Errorf("ConnectNamedPipe 실패: %w", err)
	}

	// 클라이언트 연결 또는 context 취소 대기
	if err != windows.ERROR_PIPE_CONNECTED {
		if waitErr := waitForEventOrContext(ctx, event, handle); waitErr != nil {
			return waitErr
		}
	}

	// 데이터 읽기
	buf := make([]byte, MaxPipeMessageSize)
	var bytesRead uint32

	// 읽기도 overlapped로 처리
	readEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fmt.Errorf("읽기 이벤트 생성 실패: %w", err)
	}
	defer windows.CloseHandle(readEvent)

	readOverlapped := &windows.Overlapped{
		HEvent: readEvent,
	}

	err = windows.ReadFile(handle, buf, &bytesRead, readOverlapped)
	if err == windows.ERROR_IO_PENDING {
		// 읽기 완료 대기
		if waitErr := waitForEventOrContext(ctx, readEvent, handle); waitErr != nil {
			return waitErr
		}
		// 실제 읽은 바이트 수 가져오기
		err = windows.GetOverlappedResult(handle, readOverlapped, &bytesRead, false)
		if err != nil {
			return fmt.Errorf("읽기 결과 가져오기 실패: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("파이프 읽기 실패: %w", err)
	}

	// 연결 해제
	windows.DisconnectNamedPipe(handle)

	if bytesRead == 0 {
		return nil // 빈 데이터 무시
	}

	// 널 바이트 제거 (악의적 입력 방어)
	rawData := buf[:bytesRead]
	if idx := bytes.IndexByte(rawData, 0); idx >= 0 {
		rawData = rawData[:idx]
	}

	filePath := string(rawData)

	// 유효성 검증: 빈 문자열 무시
	if len(filePath) == 0 {
		return nil
	}

	// 경로 정규화 및 절대 경로 검증
	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		return nil // 상대 경로 무시
	}

	// 유효성 검증: 파일 존재 여부
	if _, err := os.Stat(filePath); err != nil {
		return nil // 존재하지 않는 파일 무시 (ACC-011)
	}

	handler(filePath)
	return nil
}

// waitForEventOrContext 는 이벤트 시그널 또는 context 취소를 대기한다.
func waitForEventOrContext(ctx context.Context, event, pipeHandle windows.Handle) error {
	for {
		select {
		case <-ctx.Done():
			// CancelIoEx로 대기 중인 I/O 취소
			cancelIoEx(pipeHandle)
			return ctx.Err()
		default:
		}

		// 짧은 대기로 이벤트 확인
		result, err := windows.WaitForSingleObject(event, 100)
		if err != nil {
			return fmt.Errorf("이벤트 대기 실패: %w", err)
		}
		if result == windows.WAIT_OBJECT_0 {
			return nil // 이벤트 시그널됨
		}
		// WAIT_TIMEOUT이면 context 확인 후 재시도
	}
}

// cancelIoEx 는 진행 중인 I/O 작업을 취소한다.
func cancelIoEx(handle windows.Handle) {
	// CancelIoEx 호출
	modkernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procCancelIoEx := modkernel32.NewProc("CancelIoEx")
	procCancelIoEx.Call(uintptr(handle), 0)
}

// SendPath 는 Named Pipe를 통해 파일 경로를 전송한다.
// 빈 문자열이나 존재하지 않는 파일은 에러를 반환한다.
func SendPath(filePath string) error {
	// 입력 유효성 검증
	if len(filePath) == 0 {
		return fmt.Errorf("파일 경로가 비어 있습니다")
	}

	// 파일 존재 여부 확인
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("파일을 찾을 수 없습니다: %s", filePath)
	}

	pipeName, err := syscall.UTF16PtrFromString(DefaultPipeName)
	if err != nil {
		return fmt.Errorf("파이프 이름 변환 실패: %w", err)
	}

	// Named Pipe에 연결 (타임아웃 포함)
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
		return fmt.Errorf("파이프 연결 실패: %w", err)
	}
	defer windows.CloseHandle(handle)

	// 데이터 전송
	data := []byte(filePath)
	var bytesWritten uint32
	err = windows.WriteFile(handle, data, &bytesWritten, nil)
	if err != nil {
		return fmt.Errorf("파이프 쓰기 실패: %w", err)
	}

	return nil
}

