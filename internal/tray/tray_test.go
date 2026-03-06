package tray

import (
	"testing"
)

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
