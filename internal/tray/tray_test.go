package tray

import (
	"sync/atomic"
	"testing"
)

// newTestTray 는 테스트용 Tray 인스턴스를 생성하는 헬퍼이다.
func newTestTray(t *testing.T) *Tray {
	t.Helper()
	trayInst, err := NewTray([]byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00})
	if err != nil {
		t.Fatalf("NewTray() 오류: %v", err)
	}
	trayInst.quitFn = func() {} // systray.Quit 대체
	return trayInst
}

// TestNewTrayWithValidIcon 유효한 아이콘 데이터로 Tray 생성을 검증한다.
func TestNewTrayWithValidIcon(t *testing.T) {
	// 최소 유효 아이콘 데이터 (빈 데이터가 아닌 바이트 슬라이스)
	iconData := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00}

	trayInst, err := NewTray(iconData)
	if err != nil {
		t.Fatalf("NewTray() 오류: %v", err)
	}
	if trayInst == nil {
		t.Fatal("NewTray()가 nil을 반환했다")
	}
}

// TestNewTrayWithNilIcon nil 아이콘 데이터로 Tray 생성 시 에러를 검증한다.
func TestNewTrayWithNilIcon(t *testing.T) {
	_, err := NewTray(nil)
	if err == nil {
		t.Fatal("nil 아이콘 데이터로 NewTray() 호출 시 에러가 반환되어야 한다")
	}
}

// TestNewTrayWithEmptyIcon 빈 아이콘 데이터로 Tray 생성 시 에러를 검증한다.
func TestNewTrayWithEmptyIcon(t *testing.T) {
	_, err := NewTray([]byte{})
	if err == nil {
		t.Fatal("빈 아이콘 데이터로 NewTray() 호출 시 에러가 반환되어야 한다")
	}
}

// TestTrayStoresIconData Tray가 아이콘 데이터를 올바르게 저장하는지 검증한다.
func TestTrayStoresIconData(t *testing.T) {
	iconData := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00}

	trayInst, err := NewTray(iconData)
	if err != nil {
		t.Fatalf("NewTray() 오류: %v", err)
	}
	if len(trayInst.iconData) != len(iconData) {
		t.Errorf("아이콘 데이터 크기: want %d, got %d", len(iconData), len(trayInst.iconData))
	}
}

// TestTraySetTooltip 툴팁 설정을 검증한다.
func TestTraySetTooltip(t *testing.T) {
	iconData := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00}
	trayInst, err := NewTray(iconData)
	if err != nil {
		t.Fatalf("NewTray() 오류: %v", err)
	}

	trayInst.SetTooltip("테스트 툴팁")

	if trayInst.tooltip != "테스트 툴팁" {
		t.Errorf("툴팁: want %q, got %q", "테스트 툴팁", trayInst.tooltip)
	}
}

// TestTrayQuitCallsCallback Quit 호출 시 종료 콜백이 실행되는지 검증한다.
func TestTrayQuitCallsCallback(t *testing.T) {
	iconData := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00}
	trayInst, err := NewTray(iconData)
	if err != nil {
		t.Fatalf("NewTray() 오류: %v", err)
	}

	called := false
	trayInst.onQuit = func() { called = true }
	trayInst.quitFn = func() {} // systray.Quit 대체

	trayInst.Quit()

	if !called {
		t.Error("Quit() 호출 시 onQuit 콜백이 실행되어야 한다")
	}
}

// --- 동적 윈도우 목록 테스트 (Phase 5) ---

// TestTray_SetWindowListProvider 윈도우 목록 제공자와 활성화 콜백 설정을 검증한다.
func TestTray_SetWindowListProvider(t *testing.T) {
	trayInst := newTestTray(t)

	provider := func() []WindowListItem {
		return []WindowListItem{
			{ID: 1, Title: "readme.md"},
		}
	}
	activator := func(id int) {}

	trayInst.SetWindowList(provider, activator)

	if trayInst.windowListProvider == nil {
		t.Error("windowListProvider가 설정되어야 한다")
	}
	if trayInst.windowActivator == nil {
		t.Error("windowActivator가 설정되어야 한다")
	}
}

// TestTray_BuildMenuItems 는 윈도우 목록에서 메뉴 항목 구조를 빌드하는지 검증한다.
func TestTray_BuildMenuItems(t *testing.T) {
	trayInst := newTestTray(t)

	windows := []WindowListItem{
		{ID: 1, Title: "readme.md"},
		{ID: 2, Title: "design.md"},
		{ID: 3, Title: "notes.md"},
	}

	provider := func() []WindowListItem { return windows }
	activator := func(id int) {}

	trayInst.SetWindowList(provider, activator)

	items := trayInst.buildWindowMenuItems()
	if len(items) != 3 {
		t.Fatalf("메뉴 항목 수: want 3, got %d", len(items))
	}

	for i, w := range windows {
		if items[i].Title != w.Title {
			t.Errorf("항목[%d] 제목: want %q, got %q", i, w.Title, items[i].Title)
		}
		if items[i].ID != w.ID {
			t.Errorf("항목[%d] ID: want %d, got %d", i, w.ID, items[i].ID)
		}
	}
}

// TestTray_WindowActivation 윈도우 항목 클릭 시 올바른 ID로 활성화 콜백이 호출되는지 검증한다.
func TestTray_WindowActivation(t *testing.T) {
	trayInst := newTestTray(t)

	var activatedID int
	activator := func(id int) { activatedID = id }

	windows := []WindowListItem{
		{ID: 5, Title: "test.md"},
		{ID: 10, Title: "hello.md"},
	}
	provider := func() []WindowListItem { return windows }

	trayInst.SetWindowList(provider, activator)

	// 활성화 콜백 직접 호출 테스트
	trayInst.activateWindow(5)
	if activatedID != 5 {
		t.Errorf("활성화된 ID: want 5, got %d", activatedID)
	}

	trayInst.activateWindow(10)
	if activatedID != 10 {
		t.Errorf("활성화된 ID: want 10, got %d", activatedID)
	}
}

// TestTray_EmptyWindowList 윈도우가 없을 때 빈 메뉴 항목을 반환하는지 검증한다.
func TestTray_EmptyWindowList(t *testing.T) {
	trayInst := newTestTray(t)

	provider := func() []WindowListItem { return nil }
	activator := func(id int) {}

	trayInst.SetWindowList(provider, activator)

	items := trayInst.buildWindowMenuItems()
	if len(items) != 0 {
		t.Errorf("빈 목록일 때 메뉴 항목 수: want 0, got %d", len(items))
	}
}

// TestTray_DynamicUpdate 윈도우 목록이 변경될 때 메뉴 항목이 갱신되는지 검증한다.
func TestTray_DynamicUpdate(t *testing.T) {
	trayInst := newTestTray(t)

	windows := []WindowListItem{
		{ID: 1, Title: "first.md"},
	}
	provider := func() []WindowListItem { return windows }
	activator := func(id int) {}

	trayInst.SetWindowList(provider, activator)

	// 초기 상태 확인
	items := trayInst.buildWindowMenuItems()
	if len(items) != 1 {
		t.Fatalf("초기 항목 수: want 1, got %d", len(items))
	}

	// 윈도우 추가 후 다시 빌드
	windows = []WindowListItem{
		{ID: 1, Title: "first.md"},
		{ID: 2, Title: "second.md"},
	}

	// provider가 클로저이므로 windows 변수를 참조 - 하지만 새 슬라이스를 할당했으므로
	// provider를 다시 설정해야 한다
	trayInst.SetWindowList(func() []WindowListItem { return windows }, activator)

	items = trayInst.buildWindowMenuItems()
	if len(items) != 2 {
		t.Fatalf("갱신 후 항목 수: want 2, got %d", len(items))
	}
	if items[1].Title != "second.md" {
		t.Errorf("두 번째 항목 제목: want %q, got %q", "second.md", items[1].Title)
	}
}

// TestTray_ActivateWithoutActivator 활성화 콜백 없이 activateWindow 호출 시 패닉하지 않는지 검증한다.
func TestTray_ActivateWithoutActivator(t *testing.T) {
	trayInst := newTestTray(t)

	// windowActivator가 nil인 상태에서 호출해도 패닉하지 않아야 한다
	trayInst.activateWindow(1) // 패닉 없이 정상 종료되면 통과
}

// TestTray_BuildMenuItemsWithoutProvider 제공자 없이 빌드 시 빈 목록을 반환하는지 검증한다.
func TestTray_BuildMenuItemsWithoutProvider(t *testing.T) {
	trayInst := newTestTray(t)

	items := trayInst.buildWindowMenuItems()
	if len(items) != 0 {
		t.Errorf("제공자 없이 메뉴 항목 수: want 0, got %d", len(items))
	}
}

// TestQuit_이중호출_콜백한번만실행 은 Quit()을 여러 번 호출해도
// onQuit과 quitFn 콜백이 정확히 한 번만 실행되는지 검증한다.
func TestQuit_이중호출_콜백한번만실행(t *testing.T) {
	trayInst := newTestTray(t)

	var onQuitCount atomic.Int32
	var quitFnCount atomic.Int32

	trayInst.onQuit = func() { onQuitCount.Add(1) }
	trayInst.quitFn = func() { quitFnCount.Add(1) }

	// Quit()을 두 번 호출
	trayInst.Quit()
	trayInst.Quit()

	if got := onQuitCount.Load(); got != 1 {
		t.Errorf("onQuit 호출 횟수: want 1, got %d", got)
	}
	if got := quitFnCount.Load(); got != 1 {
		t.Errorf("quitFn 호출 횟수: want 1, got %d", got)
	}
}

// TestQuit_Run의onExit경로_이중호출방지 는 Run의 onExit 콜백과 외부 Quit() 호출이
// 중복 실행되지 않는지 검증한다.
func TestQuit_Run의onExit경로_이중호출방지(t *testing.T) {
	trayInst := newTestTray(t)

	var onQuitCount atomic.Int32
	onQuitCallback := func() { onQuitCount.Add(1) }

	// Run의 onExit가 호출되는 상황을 시뮬레이션:
	// 1) Quit() 호출 → onQuit + quitFn 실행
	// 2) quitFn이 systray.Quit()을 호출하면 systray가 onExit 콜백을 실행
	// 3) onExit에서 다시 Quit() 호출 → sync.Once로 스킵되어야 함
	trayInst.quitFn = func() {
		// systray.Quit() 시뮬레이션: onExit 콜백 실행
		// Run()의 onExit는 t.Quit()을 호출한다
		trayInst.Quit()
	}
	trayInst.onQuit = onQuitCallback

	trayInst.Quit()

	if got := onQuitCount.Load(); got != 1 {
		t.Errorf("Run의 onExit 경유 시 onQuit 호출 횟수: want 1, got %d", got)
	}
}

// mockMenuItem 는 테스트용 메뉴 항목 모의 객체이다.
type mockMenuItem struct {
	title   string
	clickFn func()
}

func (m *mockMenuItem) Click(fn func()) {
	m.clickFn = fn
}

// TestTray_RefreshMenu 는 메뉴 재구성이 올바르게 동작하는지 검증한다.
func TestTray_RefreshMenu(t *testing.T) {
	trayInst := newTestTray(t)
	trayInst.running = true

	var menuItems []*mockMenuItem
	resetCalled := false

	trayInst.resetMenuFn = func() {
		resetCalled = true
		menuItems = nil
	}
	trayInst.addMenuItemFn = func(title, tooltip string) menuItem {
		m := &mockMenuItem{title: title}
		menuItems = append(menuItems, m)
		return m
	}
	trayInst.addSeparatorFn = func() {}

	windows := []WindowListItem{
		{ID: 1, Title: "readme.md"},
		{ID: 2, Title: "design.md"},
	}
	provider := func() []WindowListItem { return windows }

	var activatedID int
	activator := func(id int) { activatedID = id }

	trayInst.SetWindowList(provider, activator)
	trayInst.RefreshMenu()

	if !resetCalled {
		t.Error("RefreshMenu()가 ResetMenu를 호출해야 한다")
	}

	// 윈도우 항목 2개 + 종료 항목 1개 = 3개
	if len(menuItems) != 3 {
		t.Fatalf("메뉴 항목 수: want 3, got %d", len(menuItems))
	}

	if menuItems[0].title != "readme.md" {
		t.Errorf("첫 번째 항목: want %q, got %q", "readme.md", menuItems[0].title)
	}
	if menuItems[1].title != "design.md" {
		t.Errorf("두 번째 항목: want %q, got %q", "design.md", menuItems[1].title)
	}
	if menuItems[2].title != "종료" {
		t.Errorf("마지막 항목: want %q, got %q", "종료", menuItems[2].title)
	}

	// 윈도우 항목 클릭 시 활성화 콜백 테스트
	menuItems[0].clickFn()
	if activatedID != 1 {
		t.Errorf("첫 번째 항목 클릭 후 활성화 ID: want 1, got %d", activatedID)
	}

	menuItems[1].clickFn()
	if activatedID != 2 {
		t.Errorf("두 번째 항목 클릭 후 활성화 ID: want 2, got %d", activatedID)
	}
}

// TestTray_RefreshMenuNotRunning 은 systray가 실행 중이 아닐 때 RefreshMenu가
// 아무 동작도 하지 않는지 검증한다.
func TestTray_RefreshMenuNotRunning(t *testing.T) {
	trayInst := newTestTray(t)
	// running은 기본값 false

	resetCalled := false
	trayInst.resetMenuFn = func() { resetCalled = true }

	provider := func() []WindowListItem {
		return []WindowListItem{{ID: 1, Title: "test.md"}}
	}
	trayInst.SetWindowList(provider, func(id int) {})

	trayInst.RefreshMenu()

	if resetCalled {
		t.Error("실행 중이 아닐 때 ResetMenu가 호출되면 안 된다")
	}
}

// TestTray_RefreshMenuEmptyList 는 빈 윈도우 목록으로 메뉴 재구성 시
// 종료 항목만 존재하는지 검증한다.
func TestTray_RefreshMenuEmptyList(t *testing.T) {
	trayInst := newTestTray(t)
	trayInst.running = true

	var menuItems []*mockMenuItem
	trayInst.resetMenuFn = func() { menuItems = nil }
	trayInst.addMenuItemFn = func(title, tooltip string) menuItem {
		m := &mockMenuItem{title: title}
		menuItems = append(menuItems, m)
		return m
	}
	trayInst.addSeparatorFn = func() {}

	provider := func() []WindowListItem { return nil }
	trayInst.SetWindowList(provider, func(id int) {})

	trayInst.RefreshMenu()

	// 종료 항목만 1개
	if len(menuItems) != 1 {
		t.Fatalf("빈 목록 메뉴 항목 수: want 1, got %d", len(menuItems))
	}
	if menuItems[0].title != "종료" {
		t.Errorf("유일한 항목: want %q, got %q", "종료", menuItems[0].title)
	}
}
