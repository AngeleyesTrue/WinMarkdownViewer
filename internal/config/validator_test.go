package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateTheme 테마 필드 유효성 검사를 검증한다.
func TestValidateTheme(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"유효한 light 테마", "light", "light"},
		{"유효한 dark 테마", "dark", "dark"},
		{"유효한 system 테마", "system", "system"},
		{"빈 문자열은 기본값으로", "", "system"},
		{"잘못된 테마는 기본값으로", "blue", "system"},
		{"대소문자 구분", "Light", "system"},
		{"공백 포함", " dark ", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Theme: tt.input}
			result := Validate(cfg)
			if result.Theme != tt.expected {
				t.Errorf("Validate() Theme = %q, 기대값 %q", result.Theme, tt.expected)
			}
		})
	}
}

// TestValidateFontSize 폰트 크기 유효성 검사를 검증한다.
func TestValidateFontSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"최소값 경계", 14, 14},
		{"최대값 경계", 24, 24},
		{"기본값", 16, 16},
		{"최소값 미만", 13, 16},
		{"최대값 초과", 25, 16},
		{"0은 기본값으로", 0, 16},
		{"음수는 기본값으로", -1, 16},
		{"매우 큰 값", 9999, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{FontSize: tt.input}
			result := Validate(cfg)
			if result.FontSize != tt.expected {
				t.Errorf("Validate() FontSize = %d, 기대값 %d", result.FontSize, tt.expected)
			}
		})
	}
}

// TestValidateWindowWidth 윈도우 너비 유효성 검사를 검증한다.
func TestValidateWindowWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"최소값 경계", 320, 320},
		{"최대값 경계", 7680, 7680},
		{"기본값", 1024, 1024},
		{"최소값 미만", 319, 1024},
		{"최대값 초과", 7681, 1024},
		{"0은 기본값으로", 0, 1024},
		{"음수는 기본값으로", -100, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{WindowWidth: tt.input}
			result := Validate(cfg)
			if result.WindowWidth != tt.expected {
				t.Errorf("Validate() WindowWidth = %d, 기대값 %d", result.WindowWidth, tt.expected)
			}
		})
	}
}

// TestValidateWindowHeight 윈도우 높이 유효성 검사를 검증한다.
func TestValidateWindowHeight(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"최소값 경계", 240, 240},
		{"최대값 경계", 4320, 4320},
		{"기본값", 768, 768},
		{"최소값 미만", 239, 768},
		{"최대값 초과", 4321, 768},
		{"0은 기본값으로", 0, 768},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{WindowHeight: tt.input}
			result := Validate(cfg)
			if result.WindowHeight != tt.expected {
				t.Errorf("Validate() WindowHeight = %d, 기대값 %d", result.WindowHeight, tt.expected)
			}
		})
	}
}

// TestValidateWindowPosition 윈도우 위치(X, Y) 유효성 검사를 검증한다.
func TestValidateWindowPosition(t *testing.T) {
	tests := []struct {
		name      string
		inputX    int
		inputY    int
		expectedX int
		expectedY int
	}{
		{"기본값 -1", -1, -1, -1, -1},
		{"양수 좌표", 100, 200, 100, 200},
		{"0 좌표", 0, 0, 0, 0},
		{"-1보다 작은 값은 -1로 리셋", -2, -5, -1, -1},
		{"X만 잘못된 값", -3, 100, -1, 100},
		{"Y만 잘못된 값", 50, -10, 50, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{WindowX: tt.inputX, WindowY: tt.inputY}
			result := Validate(cfg)
			if result.WindowX != tt.expectedX {
				t.Errorf("Validate() WindowX = %d, 기대값 %d", result.WindowX, tt.expectedX)
			}
			if result.WindowY != tt.expectedY {
				t.Errorf("Validate() WindowY = %d, 기대값 %d", result.WindowY, tt.expectedY)
			}
		})
	}
}

// TestValidateCustomCSS customCSS 경로 유효성 검사를 검증한다.
func TestValidateCustomCSS(t *testing.T) {
	// 실제 존재하는 CSS 파일 생성
	tmpDir := t.TempDir()
	validCSS := filepath.Join(tmpDir, "custom.css")
	if err := os.WriteFile(validCSS, []byte("body{}"), 0644); err != nil {
		t.Fatalf("테스트 CSS 파일 생성 실패: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"빈 문자열은 유지", "", ""},
		{"존재하는 절대 경로", validCSS, validCSS},
		{"존재하지 않는 파일은 빈 문자열로", filepath.Join(tmpDir, "nonexistent.css"), ""},
		{"상대 경로는 빈 문자열로", "relative/path.css", ""},
		{"디렉토리 경로는 빈 문자열로", tmpDir, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{CustomCSS: tt.input}
			result := Validate(cfg)
			if result.CustomCSS != tt.expected {
				t.Errorf("Validate() CustomCSS = %q, 기대값 %q", result.CustomCSS, tt.expected)
			}
		})
	}
}

// TestValidateLastOpenedFile lastOpenedFile는 검증하지 않음을 확인한다.
func TestValidateLastOpenedFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"빈 문자열", ""},
		{"유효한 경로", "/some/path/file.md"},
		{"존재하지 않는 경로", "/nonexistent/file.md"},
		{"임의 문자열", "anything goes here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LastOpenedFile: tt.input}
			result := Validate(cfg)
			if result.LastOpenedFile != tt.input {
				t.Errorf("Validate() LastOpenedFile = %q, 기대값 %q (변경 없어야 함)", result.LastOpenedFile, tt.input)
			}
		})
	}
}

// TestValidatePreservesValidConfig 모든 값이 유효한 설정은 변경되지 않음을 검증한다.
func TestValidatePreservesValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cssFile := filepath.Join(tmpDir, "style.css")
	if err := os.WriteFile(cssFile, []byte("body{}"), 0644); err != nil {
		t.Fatalf("CSS 파일 생성 실패: %v", err)
	}

	cfg := &Config{
		Theme:          "dark",
		FontSize:       20,
		WindowWidth:    1920,
		WindowHeight:   1080,
		WindowX:        100,
		WindowY:        50,
		CustomCSS:      cssFile,
		LastOpenedFile: "/some/file.md",
	}

	result := Validate(cfg)

	if result.Theme != "dark" {
		t.Errorf("유효한 Theme이 변경됨: %q", result.Theme)
	}
	if result.FontSize != 20 {
		t.Errorf("유효한 FontSize가 변경됨: %d", result.FontSize)
	}
	if result.WindowWidth != 1920 {
		t.Errorf("유효한 WindowWidth가 변경됨: %d", result.WindowWidth)
	}
	if result.WindowHeight != 1080 {
		t.Errorf("유효한 WindowHeight가 변경됨: %d", result.WindowHeight)
	}
	if result.WindowX != 100 {
		t.Errorf("유효한 WindowX가 변경됨: %d", result.WindowX)
	}
	if result.WindowY != 50 {
		t.Errorf("유효한 WindowY가 변경됨: %d", result.WindowY)
	}
	if result.CustomCSS != cssFile {
		t.Errorf("유효한 CustomCSS가 변경됨: %q", result.CustomCSS)
	}
	if result.LastOpenedFile != "/some/file.md" {
		t.Errorf("유효한 LastOpenedFile이 변경됨: %q", result.LastOpenedFile)
	}
}

// TestValidateReturnsNewConfig Validate가 원본을 변경하지 않는지 검증한다.
func TestValidateReturnsNewConfig(t *testing.T) {
	cfg := &Config{Theme: "invalid", FontSize: 999}
	result := Validate(cfg)

	// 원본은 변경되지 않아야 함
	if cfg.Theme != "invalid" {
		t.Error("원본 Config의 Theme이 변경됨")
	}
	if cfg.FontSize != 999 {
		t.Error("원본 Config의 FontSize가 변경됨")
	}

	// 결과는 보정된 값이어야 함
	if result.Theme != "system" {
		t.Errorf("결과 Theme = %q, 기대값 %q", result.Theme, "system")
	}
	if result.FontSize != 16 {
		t.Errorf("결과 FontSize = %d, 기대값 %d", result.FontSize, 16)
	}
}
