package config

import (
	"os"
	"path/filepath"
)

// validThemes는 허용되는 테마 값 목록이다.
var validThemes = map[string]bool{
	"light":  true,
	"dark":   true,
	"system": true,
}

// Validate는 Config의 각 필드를 검증하고, 잘못된 값은 기본값으로 보정한 새 Config를 반환한다.
// 원본 Config는 변경하지 않는다.
func Validate(cfg *Config) *Config {
	defaults := Default()
	result := &Config{}

	// 테마 검증
	if validThemes[cfg.Theme] {
		result.Theme = cfg.Theme
	} else {
		result.Theme = defaults.Theme
	}

	// 폰트 크기 검증: 14 <= n <= 24
	if cfg.FontSize >= 14 && cfg.FontSize <= 24 {
		result.FontSize = cfg.FontSize
	} else {
		result.FontSize = defaults.FontSize
	}

	// 윈도우 너비 검증: 320 <= n <= 7680
	if cfg.WindowWidth >= 320 && cfg.WindowWidth <= 7680 {
		result.WindowWidth = cfg.WindowWidth
	} else {
		result.WindowWidth = defaults.WindowWidth
	}

	// 윈도우 높이 검증: 240 <= n <= 4320
	if cfg.WindowHeight >= 240 && cfg.WindowHeight <= 4320 {
		result.WindowHeight = cfg.WindowHeight
	} else {
		result.WindowHeight = defaults.WindowHeight
	}

	// 윈도우 X 위치 검증: -1 또는 >= 0 (< -1은 -1로 리셋)
	if cfg.WindowX >= 0 || cfg.WindowX == -1 {
		result.WindowX = cfg.WindowX
	} else {
		result.WindowX = defaults.WindowX
	}

	// 윈도우 Y 위치 검증: -1 또는 >= 0 (< -1은 -1로 리셋)
	if cfg.WindowY >= 0 || cfg.WindowY == -1 {
		result.WindowY = cfg.WindowY
	} else {
		result.WindowY = defaults.WindowY
	}

	// customCSS 검증: 절대 경로이고 파일이 존재해야 함
	result.CustomCSS = validateCustomCSS(cfg.CustomCSS)

	// lastOpenedFile은 검증하지 않음
	result.LastOpenedFile = cfg.LastOpenedFile

	return result
}

// validateCustomCSS는 customCSS 경로를 검증한다.
// 빈 문자열은 그대로 반환하고, 절대 경로가 아니거나 파일이 존재하지 않으면 빈 문자열을 반환한다.
func validateCustomCSS(path string) string {
	if path == "" {
		return ""
	}

	// 절대 경로 확인
	if !filepath.IsAbs(path) {
		return ""
	}

	// 파일 존재 여부 및 일반 파일인지 확인
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		return ""
	}

	return path
}
