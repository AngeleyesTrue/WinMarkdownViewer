// Package tray 는 시스템 트레이 아이콘과 메뉴를 관리한다.
// energye/systray 라이브러리를 사용하여 트레이 아이콘, 메뉴 항목,
// 더블클릭 이벤트를 처리한다.
package tray

import (
	"fmt"

	"github.com/energye/systray"
)

// Tray 는 시스템 트레이 아이콘과 메뉴를 관리하는 구조체이다.
type Tray struct {
	iconData []byte // 트레이 아이콘 데이터 (ICO 형식)
	tooltip  string // 트레이 아이콘 툴팁 텍스트
	onQuit   func() // 종료 메뉴 클릭 시 콜백
	quitFn   func() // systray.Quit 래퍼 (테스트에서 대체 가능)
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
	}, nil
}

// Run 은 시스템 트레이를 시작한다.
// @MX:WARN: [AUTO] systray.Run()은 블로킹 호출이므로 반드시 고루틴에서 실행해야 한다
// @MX:REASON: 메인 스레드에서 호출하면 WebView2 이벤트 루프를 차단한다
func (t *Tray) Run(onReady func(), onQuit func()) {
	t.onQuit = onQuit

	systray.Run(func() {
		systray.SetIcon(t.iconData)
		systray.SetTitle("WinMarkdownViewer")
		systray.SetTooltip(t.tooltip)

		// "열기" 메뉴 항목: 클릭 시 윈도우를 표시한다
		mOpen := systray.AddMenuItem("열기", "윈도우를 표시합니다")
		mOpen.Click(func() {
			if onReady != nil {
				onReady()
			}
		})

		// 구분선
		systray.AddSeparator()

		// "종료" 메뉴 항목: 클릭 시 프로그램을 종료한다
		mQuit := systray.AddMenuItem("종료", "프로그램을 종료합니다")
		mQuit.Click(func() {
			t.Quit()
		})

		if onReady != nil {
			onReady()
		}
	}, func() {
		// onExit 콜백: systray 종료 시 정리 작업
		if t.onQuit != nil {
			t.onQuit()
		}
	})
}

// SetTooltip 은 트레이 아이콘의 툴팁 텍스트를 설정한다.
func (t *Tray) SetTooltip(text string) {
	t.tooltip = text
}

// Quit 은 시스템 트레이를 종료한다.
func (t *Tray) Quit() {
	if t.onQuit != nil {
		t.onQuit()
	}
	if t.quitFn != nil {
		t.quitFn()
	}
}
