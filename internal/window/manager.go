package window

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// defaultMaxWindows 는 기본 최대 윈도우 수이다.
const defaultMaxWindows = 10

// warningThreshold 는 윈도우 수 경고 임계값이다.
const warningThreshold = 8

// ServerFactory 는 서버를 생성하고 시작하는 팩토리 함수 타입이다.
// content 는 초기 렌더링된 HTML 콘텐츠이다.
type ServerFactory func(filePath string) (ServerHandle, error)

// WatcherFactory 는 파일 감시자를 생성하고 시작하는 팩토리 함수 타입이다.
type WatcherFactory func(filePath string) (Closeable, error)

// ManagerOption 은 WindowManager의 옵션을 설정하는 함수 타입이다.
type ManagerOption func(*WindowManager)

// WithMaxWindows 는 최대 윈도우 수를 설정하는 옵션이다.
func WithMaxWindows(max int) ManagerOption {
	return func(m *WindowManager) {
		if max > 0 {
			m.maxWindows = max
		}
	}
}

// WithServerFactory 는 서버 팩토리를 설정하는 옵션이다.
func WithServerFactory(f ServerFactory) ManagerOption {
	return func(m *WindowManager) {
		m.serverFactory = f
	}
}

// WithWatcherFactory 는 감시자 팩토리를 설정하는 옵션이다.
func WithWatcherFactory(f WatcherFactory) ManagerOption {
	return func(m *WindowManager) {
		m.watcherFactory = f
	}
}

// WindowManager 는 다중 윈도우를 관리하는 매니저이다.
// 스레드 안전하게 윈도우를 생성, 조회, 삭제한다.
type WindowManager struct {
	mu             sync.RWMutex
	windows        map[int]*Window
	nextID         int
	maxWindows     int
	serverFactory  ServerFactory
	watcherFactory WatcherFactory
	onOpened       func(WindowInfo)
	onClosed       func(WindowInfo)
}

// NewWindowManager 는 새로운 WindowManager를 생성한다.
func NewWindowManager(opts ...ManagerOption) *WindowManager {
	m := &WindowManager{
		windows:    make(map[int]*Window),
		nextID:     1,
		maxWindows: defaultMaxWindows,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// OpenFile 은 파일을 새 윈도우에서 연다.
// 이미 열린 파일이면 기존 윈도우 ID와 함께 ErrFileAlreadyOpen을 반환한다.
// 최대 윈도우 수에 도달하면 ErrMaxWindowsReached를 반환한다.
// @MX:ANCHOR: [AUTO] 윈도우 생성의 핵심 진입점으로 파일 검증, 중복 확인, 리소스 생성을 조율한다
// @MX:REASON: 다중 윈도우 아키텍처의 핵심 함수 (fan_in >= 3)
func (m *WindowManager) OpenFile(filePath string) (int, error) {
	// 1. 절대 경로로 정규화
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return 0, fmt.Errorf("경로 변환 실패: %w", err)
	}

	// 2. 파일 존재 확인
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return 0, ErrFileNotFound
		}
		return 0, fmt.Errorf("%w: %v", ErrFileNotReadable, statErr)
	}
	if info.IsDir() {
		return 0, ErrFileNotFound
	}

	// 3. 파일 읽기 가능 확인
	f, openErr := os.Open(absPath)
	if openErr != nil {
		return 0, fmt.Errorf("%w: %v", ErrFileNotReadable, openErr)
	}
	f.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 4. 이미 열린 파일인지 확인
	for _, w := range m.windows {
		if w.filePath == absPath {
			return 0, &FileAlreadyOpenError{WindowID: w.id}
		}
	}

	// 5. 최대 윈도우 수 확인
	if len(m.windows) >= m.maxWindows {
		return 0, ErrMaxWindowsReached
	}

	// 6. 팩토리가 설정되어 있는지 확인
	if m.serverFactory == nil {
		return 0, fmt.Errorf("ServerFactory가 설정되지 않았습니다")
	}
	if m.watcherFactory == nil {
		return 0, fmt.Errorf("WatcherFactory가 설정되지 않았습니다")
	}

	// 7. 서버 생성 및 시작
	srv, err := m.serverFactory(absPath)
	if err != nil {
		return 0, fmt.Errorf("서버 생성 실패: %w", err)
	}

	// 8. 감시자 생성 및 시작
	w, err := m.watcherFactory(absPath)
	if err != nil {
		srv.Close()
		return 0, fmt.Errorf("감시자 생성 실패: %w", err)
	}

	// 9. 윈도우 생성 및 등록
	id := m.nextID
	m.nextID++

	win := NewWindow(id, absPath, srv, w)
	m.windows[id] = win

	// 10. 콜백 실행
	if m.onOpened != nil {
		info := win.Info()
		go m.onOpened(info)
	}

	return id, nil
}

// CloseWindow 는 지정된 ID의 윈도우를 닫고 리소스를 정리한다.
func (m *WindowManager) CloseWindow(id int) error {
	m.mu.Lock()
	win, ok := m.windows[id]
	if !ok {
		m.mu.Unlock()
		return ErrWindowNotFound
	}
	delete(m.windows, id)
	m.mu.Unlock()

	info := win.Info()
	err := win.Close()

	if m.onClosed != nil {
		go m.onClosed(info)
	}

	return err
}

// GetWindows 는 열려있는 모든 윈도우의 정보를 반환한다.
// ID 기준으로 정렬된다.
func (m *WindowManager) GetWindows() []WindowInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]WindowInfo, 0, len(m.windows))
	for _, w := range m.windows {
		infos = append(infos, w.Info())
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ID < infos[j].ID
	})

	return infos
}

// FindByPath 는 파일 경로로 윈도우를 찾아 정보를 반환한다.
// 해당 파일이 열려있지 않으면 false를 반환한다.
func (m *WindowManager) FindByPath(path string) (*WindowInfo, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, w := range m.windows {
		if w.filePath == absPath {
			info := w.Info()
			return &info, true
		}
	}

	return nil, false
}

// Count 는 현재 열려있는 윈도우 수를 반환한다.
func (m *WindowManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.windows)
}

// Shutdown 은 모든 윈도우를 닫고 리소스를 정리한다.
func (m *WindowManager) Shutdown() {
	m.mu.Lock()
	windows := make([]*Window, 0, len(m.windows))
	for _, w := range m.windows {
		windows = append(windows, w)
	}
	m.windows = make(map[int]*Window)
	m.mu.Unlock()

	for _, w := range windows {
		info := w.Info()
		w.Close()
		if m.onClosed != nil {
			go m.onClosed(info)
		}
	}
}

// GetWindow 는 지정된 ID의 윈도우 정보를 반환한다.
// 해당 ID의 윈도우가 없으면 nil, false를 반환한다.
func (m *WindowManager) GetWindow(id int) (*WindowInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w, ok := m.windows[id]
	if !ok {
		return nil, false
	}
	info := w.Info()
	return &info, true
}

// OnWindowOpened 는 윈도우가 열릴 때 호출될 콜백을 설정한다.
func (m *WindowManager) OnWindowOpened(fn func(WindowInfo)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onOpened = fn
}

// OnWindowClosed 는 윈도우가 닫힐 때 호출될 콜백을 설정한다.
func (m *WindowManager) OnWindowClosed(fn func(WindowInfo)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onClosed = fn
}
