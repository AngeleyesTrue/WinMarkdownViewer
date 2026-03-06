package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestConfigPath ConfigPath()가 올바른 경로를 반환하는지 검증한다.
func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() 오류: %v", err)
	}

	expected := filepath.Join(tmpDir, "config.json")
	if path != expected {
		t.Errorf("ConfigPath() = %q, 기대값 %q", path, expected)
	}
}

// TestConfigPathEnvOverride 환경 변수로 설정 경로를 오버라이드할 수 있는지 검증한다.
func TestConfigPathEnvOverride(t *testing.T) {
	customDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", customDir)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() 오류: %v", err)
	}

	if !strings.HasPrefix(path, customDir) {
		t.Errorf("ConfigPath()가 환경 변수 경로를 사용하지 않음: %q", path)
	}
}

// TestConfigPathDefaultFallback 환경 변수 미설정 시 기본 경로를 사용하는지 검증한다.
func TestConfigPathDefaultFallback(t *testing.T) {
	t.Setenv("WINMDVIEW_CONFIG_DIR", "")

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() 오류: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("WinMarkdownViewer", "config.json")) {
		t.Errorf("ConfigPath()가 올바른 기본 경로가 아님: %q", path)
	}
}

// TestLoadCreatesDefaultOnFirstRun 파일이 없을 때 기본 설정을 생성하는지 검증한다.
func TestLoadCreatesDefaultOnFirstRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	// 기본값이어야 함
	defaults := Default()
	if cfg.Theme != defaults.Theme {
		t.Errorf("Load() Theme = %q, 기대값 %q", cfg.Theme, defaults.Theme)
	}
	if cfg.FontSize != defaults.FontSize {
		t.Errorf("Load() FontSize = %d, 기대값 %d", cfg.FontSize, defaults.FontSize)
	}

	// 파일이 생성되었는지 확인
	configFile := filepath.Join(tmpDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Load()가 설정 파일을 생성하지 않음")
	}
}

// TestLoadCreatesDirectory 설정 디렉토리가 없을 때 자동 생성하는지 검증한다.
func TestLoadCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nested", "config", "dir")
	t.Setenv("WINMDVIEW_CONFIG_DIR", configDir)

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Load()가 설정 디렉토리를 생성하지 않음")
	}
}

// TestLoadExistingConfig 기존 설정 파일을 올바르게 읽는지 검증한다.
func TestLoadExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	// 설정 파일 직접 작성
	cfg := &Config{
		Theme:        "dark",
		FontSize:     20,
		WindowWidth:  1920,
		WindowHeight: 1080,
		WindowX:      100,
		WindowY:      50,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("설정 파일 작성 실패: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	if loaded.Theme != "dark" {
		t.Errorf("Load() Theme = %q, 기대값 %q", loaded.Theme, "dark")
	}
	if loaded.FontSize != 20 {
		t.Errorf("Load() FontSize = %d, 기대값 %d", loaded.FontSize, 20)
	}
	if loaded.WindowWidth != 1920 {
		t.Errorf("Load() WindowWidth = %d, 기대값 %d", loaded.WindowWidth, 1920)
	}
}

// TestLoadCorruptedJSON 손상된 JSON 파일에 대해 백업 후 기본값을 생성하는지 검증한다.
func TestLoadCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte("{invalid json!!!"), 0644); err != nil {
		t.Fatalf("손상된 설정 파일 작성 실패: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 손상된 JSON에서 오류가 아닌 기본값을 반환해야 함: %v", err)
	}

	// 기본값이어야 함
	if cfg.Theme != "system" {
		t.Errorf("Load() Theme = %q, 기대값 %q", cfg.Theme, "system")
	}

	// .bak 파일이 생성되었는지 확인
	bakFile := configFile + ".bak"
	if _, err := os.Stat(bakFile); os.IsNotExist(err) {
		t.Error("Load()가 손상된 파일의 백업(.bak)을 생성하지 않음")
	}

	// 백업 내용이 원본과 동일한지 확인
	bakData, err := os.ReadFile(bakFile)
	if err != nil {
		t.Fatalf("백업 파일 읽기 실패: %v", err)
	}
	if string(bakData) != "{invalid json!!!" {
		t.Errorf("백업 파일 내용이 원본과 다름: %q", string(bakData))
	}
}

// TestLoadPartialConfig 일부 필드만 있는 설정에서 누락 필드를 기본값으로 채우는지 검증한다.
func TestLoadPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	// theme만 설정된 부분 설정
	partial := `{"theme": "dark"}`
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(partial), 0644); err != nil {
		t.Fatalf("부분 설정 파일 작성 실패: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	// 명시된 필드는 유지
	if cfg.Theme != "dark" {
		t.Errorf("Load() Theme = %q, 기대값 %q", cfg.Theme, "dark")
	}

	// 누락된 필드는 Validate를 통해 기본값으로 보정
	defaults := Default()
	if cfg.FontSize != defaults.FontSize {
		t.Errorf("Load() FontSize = %d, 기대값 %d", cfg.FontSize, defaults.FontSize)
	}
	if cfg.WindowWidth != defaults.WindowWidth {
		t.Errorf("Load() WindowWidth = %d, 기대값 %d", cfg.WindowWidth, defaults.WindowWidth)
	}
}

// TestLoadUnknownFields 알 수 없는 필드는 무시하는지 검증한다.
func TestLoadUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	jsonData := `{"theme": "dark", "fontSize": 18, "unknownField": "value", "anotherUnknown": 42}`
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(jsonData), 0644); err != nil {
		t.Fatalf("설정 파일 작성 실패: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 알 수 없는 필드가 있어도 오류 없이 로드해야 함: %v", err)
	}

	if cfg.Theme != "dark" {
		t.Errorf("Load() Theme = %q, 기대값 %q", cfg.Theme, "dark")
	}
	if cfg.FontSize != 18 {
		t.Errorf("Load() FontSize = %d, 기대값 %d", cfg.FontSize, 18)
	}
}

// TestSaveConfig Save()가 올바른 형식으로 파일을 저장하는지 검증한다.
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	cfg := Default()
	cfg.Theme = "dark"
	cfg.FontSize = 20

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() 오류: %v", err)
	}

	// 파일 내용 확인
	configFile := filepath.Join(tmpDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("저장된 설정 파일 읽기 실패: %v", err)
	}

	// 2-space indent 형식 확인
	if !strings.Contains(string(data), "  \"theme\"") {
		t.Error("Save()가 2-space indent를 사용하지 않음")
	}

	// JSON 파싱 확인
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("저장된 JSON 파싱 실패: %v", err)
	}
	if loaded.Theme != "dark" {
		t.Errorf("저장된 Theme = %q, 기대값 %q", loaded.Theme, "dark")
	}
	if loaded.FontSize != 20 {
		t.Errorf("저장된 FontSize = %d, 기대값 %d", loaded.FontSize, 20)
	}
}

// TestSaveCreatesDirectory Save()가 디렉토리를 자동 생성하는지 검증한다.
func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "new", "dir")
	t.Setenv("WINMDVIEW_CONFIG_DIR", configDir)

	cfg := Default()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() 오류: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Save()가 설정 파일을 생성하지 않음")
	}
}

// TestSaveAndLoadRoundTrip Save 후 Load가 동일한 설정을 반환하는지 검증한다.
func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	original := &Config{
		Theme:          "light",
		FontSize:       18,
		WindowWidth:    1600,
		WindowHeight:   900,
		WindowX:        200,
		WindowY:        100,
		CustomCSS:      "",
		LastOpenedFile: "/test/file.md",
	}

	if err := Save(original); err != nil {
		t.Fatalf("Save() 오류: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	if loaded.Theme != original.Theme {
		t.Errorf("라운드트립 Theme = %q, 기대값 %q", loaded.Theme, original.Theme)
	}
	if loaded.FontSize != original.FontSize {
		t.Errorf("라운드트립 FontSize = %d, 기대값 %d", loaded.FontSize, original.FontSize)
	}
	if loaded.WindowWidth != original.WindowWidth {
		t.Errorf("라운드트립 WindowWidth = %d, 기대값 %d", loaded.WindowWidth, original.WindowWidth)
	}
	if loaded.WindowHeight != original.WindowHeight {
		t.Errorf("라운드트립 WindowHeight = %d, 기대값 %d", loaded.WindowHeight, original.WindowHeight)
	}
	if loaded.WindowX != original.WindowX {
		t.Errorf("라운드트립 WindowX = %d, 기대값 %d", loaded.WindowX, original.WindowX)
	}
	if loaded.WindowY != original.WindowY {
		t.Errorf("라운드트립 WindowY = %d, 기대값 %d", loaded.WindowY, original.WindowY)
	}
	if loaded.LastOpenedFile != original.LastOpenedFile {
		t.Errorf("라운드트립 LastOpenedFile = %q, 기대값 %q", loaded.LastOpenedFile, original.LastOpenedFile)
	}
}

// TestConcurrentLoadSave 동시 Load/Save에서 데이터 경합이 없는지 검증한다.
func TestConcurrentLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	// 초기 설정 파일 생성
	cfg := Default()
	if err := Save(cfg); err != nil {
		t.Fatalf("초기 Save() 오류: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	// 동시 Load 10개
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Load(); err != nil {
				errCh <- err
			}
		}()
	}

	// 동시 Save 10개
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := Default()
			c.FontSize = 14 + idx%11
			if err := Save(c); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("동시 작업 중 오류 발생: %v", err)
	}
}

// TestLoadValidatesConfig Load가 유효성 검사를 수행하는지 검증한다.
func TestLoadValidatesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	// 잘못된 값이 포함된 설정 파일
	invalidCfg := `{"theme": "invalid", "fontSize": 999, "windowWidth": -1}`
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(invalidCfg), 0644); err != nil {
		t.Fatalf("설정 파일 작성 실패: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 오류: %v", err)
	}

	// Validate를 통해 보정되어야 함
	if cfg.Theme != "system" {
		t.Errorf("Load() Theme = %q, 기대값 %q (유효성 검사 후)", cfg.Theme, "system")
	}
	if cfg.FontSize != 16 {
		t.Errorf("Load() FontSize = %d, 기대값 %d (유효성 검사 후)", cfg.FontSize, 16)
	}
	if cfg.WindowWidth != 1024 {
		t.Errorf("Load() WindowWidth = %d, 기대값 %d (유효성 검사 후)", cfg.WindowWidth, 1024)
	}
}

// TestSaveJSONFormat Save가 2-space indent로 저장하는지 검증한다.
func TestSaveJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WINMDVIEW_CONFIG_DIR", tmpDir)

	cfg := Default()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() 오류: %v", err)
	}

	configFile := filepath.Join(tmpDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("설정 파일 읽기 실패: %v", err)
	}

	content := string(data)

	// 2-space indent 확인 (tab이 아닌 공백 2개)
	if strings.Contains(content, "\t") {
		t.Error("Save() JSON에 탭이 포함됨 (2-space indent여야 함)")
	}
	if !strings.Contains(content, "  ") {
		t.Error("Save() JSON에 2-space indent가 없음")
	}

	// 유효한 JSON인지 확인
	var check map[string]interface{}
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("저장된 파일이 유효한 JSON이 아님: %v", err)
	}
}
