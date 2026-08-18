package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// mu는 설정 파일 읽기/쓰기의 동시성을 보호하는 뮤텍스이다.
var mu sync.RWMutex

// ConfigPath는 설정 파일의 전체 경로를 반환한다.
// WINMDVIEW_CONFIG_DIR 환경 변수가 설정되어 있으면 해당 경로를 사용하고,
// 그렇지 않으면 os.UserConfigDir()을 기본 경로로 사용한다.
func ConfigPath() (string, error) {
	dir := os.Getenv("WINMDVIEW_CONFIG_DIR")
	if dir == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(userConfigDir, "WinMarkdownViewer")
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load는 설정 파일을 읽어 Config를 반환한다.
// 파일이 없으면 기본 설정을 생성하고, 손상된 JSON은 백업 후 기본값으로 복원한다.
// 로드된 설정은 Validate를 통해 유효성 검사를 거친다.
func Load() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// 디렉토리 생성 (존재하지 않는 경우)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 파일이 존재하지 않으면 기본 설정 생성 후 반환
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg := Default()
		if saveErr := saveToFile(configPath, cfg); saveErr != nil {
			return nil, saveErr
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	// JSON 파싱 시도
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// 손상된 JSON: 백업 후 기본값 생성
		bakPath := configPath + ".bak"
		_ = os.WriteFile(bakPath, data, 0644)

		defaultCfg := Default()
		if saveErr := saveToFile(configPath, defaultCfg); saveErr != nil {
			return nil, saveErr
		}
		return defaultCfg, nil
	}

	// 유효성 검사 및 보정
	validated := Validate(&cfg)

	return validated, nil
}

// Save는 Config를 설정 파일에 저장한다.
// 2-space indent JSON 형식으로 저장하며, 디렉토리가 없으면 자동 생성한다.
func Save(cfg *Config) error {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// 디렉토리 생성
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return saveToFile(configPath, cfg)
}

// saveToFile은 Config를 지정된 경로에 2-space indent JSON으로 저장한다.
// mu 잠금이 이미 획득된 상태에서 호출되어야 한다.
func saveToFile(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// 줄 바꿈 추가
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}
