package viewer

import "errors"

// ErrWebView2NotInstalled 은 WebView2 Runtime이 설치되지 않았을 때 반환되는 에러이다.
var ErrWebView2NotInstalled = errors.New("Microsoft Edge WebView2 Runtime이 설치되지 않았습니다.\n" +
	"다음 링크에서 설치할 수 있습니다:\n" +
	"https://developer.microsoft.com/en-us/microsoft-edge/webview2/")
