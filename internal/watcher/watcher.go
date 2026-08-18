// Package watcher 는 fsnotify를 사용하여 파일 변경을 감시하고 디바운스된 알림을 제공한다.
package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounceInterval 은 연속 파일 변경 이벤트를 병합하기 위한 디바운스 간격이다.
const debounceInterval = 100 * time.Millisecond

// pollingInterval 은 fsnotify 감시 실패 시 사용되는 폴링 간격이다.
const pollingInterval = 100 * time.Millisecond

// Watcher 는 단일 파일의 변경을 감시하는 구조체이다.
type Watcher struct {
	mu       sync.Mutex
	filePath string
	fsw      *fsnotify.Watcher
	closed   bool
}

// NewWatcher 는 지정된 파일 경로에 대한 새로운 Watcher를 생성한다.
// 파일이 존재하지 않으면 에러를 반환한다.
func NewWatcher(filePath string) (*Watcher, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("절대 경로 변환 실패: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("파일을 찾을 수 없습니다: %w", err)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify 워처 생성 실패: %w", err)
	}

	// 파일 자체를 감시한다
	if err := fsw.Add(absPath); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("파일 감시 등록 실패: %w", err)
	}

	// 부모 디렉토리도 감시한다 (파일 삭제 후 재생성 감지용)
	parentDir := filepath.Dir(absPath)
	if err := fsw.Add(parentDir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("디렉토리 감시 등록 실패: %w", err)
	}

	return &Watcher{
		filePath: absPath,
		fsw:      fsw,
	}, nil
}

// Watch 는 파일 변경을 감시하고 변경 알림 채널을 반환한다.
// 컨텍스트가 취소되면 채널이 닫힌다.
// Write 이벤트만 처리하며 디바운스를 적용한다.
func (w *Watcher) Watch(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	// done 은 메인 루프 종료 시 서브 고루틴들에게 알리기 위한 채널이다.
	done := make(chan struct{})
	// wg 는 서브 고루틴(waitForRecovery, debounceTimer)의 완료를 추적한다.
	var wg sync.WaitGroup

	go func() {
		var debounceTimer *time.Timer

		// 메인 이벤트 루프
		func() {
			for {
				select {
				case <-ctx.Done():
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					return

				case event, ok := <-w.fsw.Events:
					if !ok {
						return
					}

					w.mu.Lock()
					currentPath := w.filePath
					w.mu.Unlock()

					// 감시 중인 파일에 대한 이벤트만 처리한다
					eventPath, _ := filepath.Abs(event.Name)
					if eventPath != currentPath {
						continue
					}

					// Write 또는 Create(파일 복구) 이벤트만 처리한다
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
						// Create 이벤트 시 파일 감시를 다시 등록한다
						if event.Has(fsnotify.Create) {
							_ = w.fsw.Add(currentPath)
						}

						// 디바운스 타이머를 설정하거나 리셋한다
						if debounceTimer == nil {
							debounceTimer = time.AfterFunc(debounceInterval, func() {
								select {
								case ch <- struct{}{}:
								case <-done:
								default:
								}
							})
						} else {
							debounceTimer.Reset(debounceInterval)
						}
					}

					// 파일 삭제 시 폴링으로 복구를 기다린다
					if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						wg.Add(1)
						go w.waitForRecovery(ctx, currentPath, ch, done, &wg)
					}

				case _, ok := <-w.fsw.Errors:
					if !ok {
						return
					}
					// 에러를 무시하고 계속 감시한다
				}
			}
		}()

		// done을 닫아 모든 서브 고루틴에게 종료를 알린다
		close(done)
		// 모든 서브 고루틴이 종료될 때까지 기다린다
		wg.Wait()
		// 모든 서브 고루틴이 종료된 후 안전하게 채널을 닫는다
		close(ch)
	}()

	return ch
}

// waitForRecovery 는 삭제된 파일이 다시 생성될 때까지 폴링으로 대기한다.
// done 채널이 닫히면 즉시 종료하여 닫힌 채널에 쓰기를 방지한다.
func (w *Watcher) waitForRecovery(ctx context.Context, filePath string, ch chan<- struct{}, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if _, err := os.Stat(filePath); err == nil {
				// 파일이 복구되었다. 감시를 다시 등록한다.
				_ = w.fsw.Add(filePath)
				select {
				case ch <- struct{}{}:
				case <-done:
				default:
				}
				return
			}
		}
	}
}

// SwitchFile 은 감시 대상 파일을 변경한다.
// 새 파일이 존재하지 않으면 에러를 반환한다.
func (w *Watcher) SwitchFile(newPath string) error {
	absPath, err := filepath.Abs(newPath)
	if err != nil {
		return fmt.Errorf("절대 경로 변환 실패: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("파일을 찾을 수 없습니다: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// 이전 파일과 디렉토리의 감시를 해제한다
	oldDir := filepath.Dir(w.filePath)
	_ = w.fsw.Remove(w.filePath)
	_ = w.fsw.Remove(oldDir)

	// 새 파일과 디렉토리를 감시 등록한다
	if err := w.fsw.Add(absPath); err != nil {
		return fmt.Errorf("새 파일 감시 등록 실패: %w", err)
	}

	newDir := filepath.Dir(absPath)
	if err := w.fsw.Add(newDir); err != nil {
		return fmt.Errorf("새 디렉토리 감시 등록 실패: %w", err)
	}

	w.filePath = absPath
	return nil
}

// Close 는 Watcher의 리소스를 정리한다.
// 이미 닫힌 경우에도 안전하게 호출할 수 있다.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	return w.fsw.Close()
}
