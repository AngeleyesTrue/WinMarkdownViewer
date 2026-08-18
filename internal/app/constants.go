// Package app 에서 사용하는 공유 상수를 정의한다.
package app

// 단일 인스턴스 제어 및 IPC에 사용되는 이름 상수
const (
	// DefaultMutexName 은 단일 인스턴스 제어에 사용되는 Named Mutex 이름이다.
	DefaultMutexName = "WinMarkdownViewer_SingleInstance"

	// DefaultPipeName 은 인스턴스 간 파일 경로 전달에 사용되는 Named Pipe 이름이다.
	DefaultPipeName = `\\.\pipe\WinMarkdownViewer`

	// MaxPipeMessageSize 는 파이프를 통해 수신 가능한 최대 바이트 수이다.
	MaxPipeMessageSize = 4096

	// PipeTimeoutMs 는 파이프 연결/작업 타임아웃 (밀리초)이다.
	PipeTimeoutMs = 5000

	// PipeCommandOpen 은 파이프 프로토콜에서 파일 열기 명령 접두사이다.
	PipeCommandOpen = "OPEN:"
)
