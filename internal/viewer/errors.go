package viewer

import "errors"

// ErrWebView2NotInstalled 은 WebView2 Runtime이 설치되지 않았을 때 반환되는 에러이다.
// Windows 11에는 기본 포함되어 있으며, Windows 10에서는 별도 설치가 필요하다.
var ErrWebView2NotInstalled = errors.New(
	"[WMV-E001] Microsoft Edge WebView2 Runtime이 설치되지 않았습니다.\n" +
		"설치 방법: https://developer.microsoft.com/en-us/microsoft-edge/webview2/\n" +
		"Windows 11에는 기본 포함되어 있습니다.")
