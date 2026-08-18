package viewer

import "unsafe"

// mockWebView에 확장된 WebView 인터페이스 메서드를 추가한다.
// 기존 viewer_test.go의 mockWebView 구조체를 보완하여 인터페이스를 충족시킨다.

func (m *mockWebView) Window() unsafe.Pointer    { return nil }
func (m *mockWebView) Terminate()                { m.destroyed = true }
func (m *mockWebView) Dispatch(f func())         { f() }
