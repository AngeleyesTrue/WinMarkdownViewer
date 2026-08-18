// Package tray 는 시스템 트레이 아이콘과 메뉴를 관리한다.
// energye/systray 라이브러리를 사용하여 트레이 아이콘, 메뉴 항목,
// 더블클릭 이벤트를 처리한다.
package tray

import (
	"fmt"
	"sync/atomic"

	"github.com/energye/systray"
)

// WindowListItem 은 트레이 메뉴에 표시할 윈도우 정보이다.
type WindowListItem struct {
	ID    int    // 윈도우 고유 ID
	Title string // 표시용 파일명 (예: "readme.md")
}

// WindowListProvider 는 현재 열린 윈도우 목록을 반환하는 콜백 타입이다.
type WindowListProvider func() []WindowListItem

// WindowActivator 는 트레이 메뉴에서 윈도우 항목 클릭 시 호출되는 콜백 타입이다.
type WindowActivator func(windowID int)

// Tray 는 시스템 트레이 아이콘과 메뉴를 관리하는 구조체이다.
type Tray struct {
	iconData           []byte             // 트레이 아이콘 데이터 (ICO 형식)
	tooltip            string             // 트레이 아이콘 툴팁 텍스트
	onQuit             func()             // 종료 메뉴 클릭 시 콜백
	quitFn             func()             // systray.Quit 래퍼 (테스트에서 대체 가능)
	windowListProvider WindowListProvider // 윈도우 목록 제공자
	windowActivator    WindowActivator    // 윈도우 활성화 콜백
	running            bool               // systray 실행 상태
	quitting           atomic.Bool        // Quit() 이중 호출 방지용 (CAS 기반)
	resetMenuFn        func()             // systray.ResetMenu 래퍼 (테스트에서 대체 가능)
	addMenuItemFn      func(title, tooltip string) menuItem // systray.AddMenuItem 래퍼
	addSeparatorFn     func()             // systray.AddSeparator 래퍼
}

// menuItem 은 systray.MenuItem를 추상화한 인터페이스이다.
// 테스트에서 모의 객체로 대체할 수 있다.
type menuItem interface {
	Click(fn func())
}

// NewTray 는 아이콘 데이터를 사용하여 새로운 Tray 인스턴스를 생성한다.
// 아이콘 데이터가 nil이거나 비어있으면 에러를 반환한다.
func NewTray(iconData []byte) (*Tray, error) {
	if len(iconData) == 0 {
		return nil, fmt.Errorf("아이콘 데이터가 비어있습니다")
	}

	return &Tray{
		iconData: iconData,
		tooltip:  "WinMarkdownViewer",
		quitFn:   systray.Quit,
		resetMenuFn: systray.ResetMenu,
		addMenuItemFn: func(title, tooltip string) menuItem {
			return systray.AddMenuItem(title, tooltip)
		},
		addSeparatorFn: systray.AddSeparator,
	}, nil
}

// Run 은 시스템 트레이를 시작한다.
func (t *Tray) Run(onReady func(), onQuit func()) {
	t.onQuit = onQuit

	systray.Run(func() {
		t.running = true
		systray.SetIcon(t.iconData)
		systray.SetTitle("WinMarkdownViewer")
		systray.SetTooltip(t.tooltip)

		t.populateMenu()

		if onReady != nil {
			onReady()
		}
	}, func() {
		// onExit 콜백: systray 종료 시 정리 작업
		// Quit()의 atomic CAS가 이중 호출을 방지하므로 안전하게 호출 가능
		t.running = false
		t.Quit()
	})
}

// SetTooltip 은 트레이 아이콘의 툴팁 텍스트를 설정한다.
func (t *Tray) SetTooltip(text string) {
	t.tooltip = text
}

// SetWindowList 는 윈도우 목록 제공자와 활성화 콜백을 설정한다.
// 트레이 메뉴에 동적 윈도우 목록을 표시하려면 이 메서드로 콜백을 등록한다.
func (t *Tray) SetWindowList(provider WindowListProvider, activator WindowActivator) {
	t.windowListProvider = provider
	t.windowActivator = activator
}

// buildWindowMenuItems 는 현재 윈도우 목록에서 메뉴 항목 정보를 빌드한다.
// provider가 nil이면 빈 슬라이스를 반환한다.
func (t *Tray) buildWindowMenuItems() []WindowListItem {
	if t.windowListProvider == nil {
		return nil
	}
	return t.windowListProvider()
}

// activateWindow 는 지정된 ID의 윈도우를 활성화한다.
// windowActivator가 nil이면 아무 동작도 하지 않는다.
func (t *Tray) activateWindow(windowID int) {
	if t.windowActivator != nil {
		t.windowActivator(windowID)
	}
}

// populateMenu 는 현재 윈도우 목록으로 트레이 메뉴 항목을 구성한다.
// Run의 초기 메뉴 구성과 RefreshMenu의 메뉴 재구성에서 공통으로 사용한다.
func (t *Tray) populateMenu() {
	// 윈도우 목록 메뉴 항목 추가
	items := t.buildWindowMenuItems()
	for _, item := range items {
		id := item.ID
		m := t.addMenuItemFn(item.Title, item.Title)
		m.Click(func() {
			t.activateWindow(id)
		})
	}

	// 구분선
	t.addSeparatorFn()

	// "종료" 메뉴 항목
	mQuit := t.addMenuItemFn("종료", "프로그램을 종료합니다")
	mQuit.Click(func() {
		t.Quit()
	})
}

// RefreshMenu 는 트레이 메뉴를 현재 윈도우 목록으로 재구성한다.
// systray가 실행 중일 때만 동작한다.
func (t *Tray) RefreshMenu() {
	if !t.running {
		return
	}

	t.resetMenuFn()
	t.populateMenu()
}

// Quit 은 시스템 트레이를 종료한다.
func (t *Tray) Quit() {
	if !t.quitting.CompareAndSwap(false, true) {
		return
	}
	if t.onQuit != nil {
		t.onQuit()
	}
	if t.quitFn != nil {
		t.quitFn()
	}
}
