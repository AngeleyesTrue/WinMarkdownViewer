// Package config는 WinMarkdownViewer의 사용자 설정 관리를 담당한다.
// JSON 기반 설정 파일을 %APPDATA%\WinMarkdownViewer\config.json에 저장하며,
// 스레드 안전한 읽기/쓰기를 지원한다.
package config

// Config는 사용자 설정을 나타내는 구조체이다.
type Config struct {
	Theme          string `json:"theme"`         // 테마: "light", "dark", "system"
	FontSize       int    `json:"fontSize"`       // 폰트 크기: 14-24
	WindowWidth    int    `json:"windowWidth"`    // 윈도우 너비: 320-7680
	WindowHeight   int    `json:"windowHeight"`   // 윈도우 높이: 240-4320
	WindowX        int    `json:"windowX"`        // 윈도우 X 위치: -1 또는 >=0
	WindowY        int    `json:"windowY"`        // 윈도우 Y 위치: -1 또는 >=0
	CustomCSS      string `json:"customCSS"`      // 커스텀 CSS 파일 절대 경로
	LastOpenedFile string `json:"lastOpenedFile"` // 마지막으로 연 파일 경로
}

// Default는 모든 필드가 기본값으로 설정된 새 Config 인스턴스를 반환한다.
func Default() *Config {
	return &Config{
		Theme:          "system",
		FontSize:       16,
		WindowWidth:    1024,
		WindowHeight:   768,
		WindowX:        -1,
		WindowY:        -1,
		CustomCSS:      "",
		LastOpenedFile: "",
	}
}
