package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDefaultConfig Default()가 올바른 기본값을 반환하는지 검증한다.
func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Theme != "system" {
		t.Errorf("Theme 기본값 = %q, 기대값 %q", cfg.Theme, "system")
	}
	if cfg.FontSize != 16 {
		t.Errorf("FontSize 기본값 = %d, 기대값 %d", cfg.FontSize, 16)
	}
	if cfg.WindowWidth != 1024 {
		t.Errorf("WindowWidth 기본값 = %d, 기대값 %d", cfg.WindowWidth, 1024)
	}
	if cfg.WindowHeight != 768 {
		t.Errorf("WindowHeight 기본값 = %d, 기대값 %d", cfg.WindowHeight, 768)
	}
	if cfg.WindowX != -1 {
		t.Errorf("WindowX 기본값 = %d, 기대값 %d", cfg.WindowX, -1)
	}
	if cfg.WindowY != -1 {
		t.Errorf("WindowY 기본값 = %d, 기대값 %d", cfg.WindowY, -1)
	}
	if cfg.CustomCSS != "" {
		t.Errorf("CustomCSS 기본값 = %q, 기대값 %q", cfg.CustomCSS, "")
	}
	if cfg.LastOpenedFile != "" {
		t.Errorf("LastOpenedFile 기본값 = %q, 기대값 %q", cfg.LastOpenedFile, "")
	}
}

// TestDefaultConfigReturnsNewInstance Default()가 매 호출마다 새 인스턴스를 반환하는지 검증한다.
func TestDefaultConfigReturnsNewInstance(t *testing.T) {
	cfg1 := Default()
	cfg2 := Default()

	if cfg1 == cfg2 {
		t.Error("Default()는 매 호출마다 새 인스턴스를 반환해야 한다")
	}

	// 값을 변경해도 다른 인스턴스에 영향 없음을 검증
	cfg1.Theme = "dark"
	cfg3 := Default()
	if cfg3.Theme != "system" {
		t.Errorf("Default() 반환값이 이전 인스턴스에 의해 오염됨: %q", cfg3.Theme)
	}
}

// TestConfigJSONTags JSON 직렬화 태그가 올바른지 검증한다.
func TestConfigJSONTags(t *testing.T) {
	cfg := Default()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{
		`"theme"`,
		`"fontSize"`,
		`"windowWidth"`,
		`"windowHeight"`,
		`"windowX"`,
		`"windowY"`,
		`"customCSS"`,
		`"lastOpenedFile"`,
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON에 필드 %s 가 없음.\nJSON: %s", field, jsonStr)
		}
	}
}
