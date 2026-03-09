package main

import (
	"testing"
)

// TestParseFlagsRegister --register 플래그 파싱을 검증한다.
func TestParseFlagsRegister(t *testing.T) {
	flags := parseFlags([]string{"--register"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	if flags.unregister {
		t.Error("--unregister 플래그가 false여야 한다")
	}
	if flags.filePath != "" {
		t.Errorf("filePath가 비어있어야 한다: got %q", flags.filePath)
	}
}

// TestParseFlagsUnregister --unregister 플래그 파싱을 검증한다.
func TestParseFlagsUnregister(t *testing.T) {
	flags := parseFlags([]string{"--unregister"})

	if !flags.unregister {
		t.Error("--unregister 플래그가 true여야 한다")
	}
	if flags.register {
		t.Error("--register 플래그가 false여야 한다")
	}
}

// TestParseFlagsSetDefault --set-default 플래그 파싱을 검증한다.
func TestParseFlagsSetDefault(t *testing.T) {
	flags := parseFlags([]string{"--set-default"})

	if !flags.setDefault {
		t.Error("--set-default 플래그가 true여야 한다")
	}
}

// TestParseFlagsFilePath 파일 경로 인자 파싱을 검증한다.
func TestParseFlagsFilePath(t *testing.T) {
	flags := parseFlags([]string{"test.md"})

	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
	if flags.register {
		t.Error("register 플래그가 false여야 한다")
	}
}

// TestParseFlagsRegisterWithFilePath --register와 파일 경로가 함께 전달된 경우를 검증한다 (ACC-014).
// 등록만 수행하고 파일을 열지 않아야 한다.
func TestParseFlagsRegisterWithFilePath(t *testing.T) {
	flags := parseFlags([]string{"--register", "test.md"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	// filePath도 파싱되지만, register가 true이면 파일 열기는 수행하지 않는다
	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
}

// TestParseFlagsNoArgs 인자 없이 호출된 경우를 검증한다.
func TestParseFlagsNoArgs(t *testing.T) {
	flags := parseFlags([]string{})

	if flags.register || flags.unregister || flags.setDefault {
		t.Error("플래그가 모두 false여야 한다")
	}
	if flags.filePath != "" {
		t.Errorf("filePath가 비어있어야 한다: got %q", flags.filePath)
	}
}

// TestParseFlagsMultipleFlags 여러 플래그 조합을 검증한다.
func TestParseFlagsMultipleFlags(t *testing.T) {
	flags := parseFlags([]string{"--register", "--set-default"})

	if !flags.register {
		t.Error("--register 플래그가 true여야 한다")
	}
	if !flags.setDefault {
		t.Error("--set-default 플래그가 true여야 한다")
	}
}

// TestParseFlagsUnknownFlagIgnored 알 수 없는 플래그는 무시된다.
func TestParseFlagsUnknownFlagIgnored(t *testing.T) {
	flags := parseFlags([]string{"--unknown", "test.md"})

	if flags.filePath != "test.md" {
		t.Errorf("filePath: want %q, got %q", "test.md", flags.filePath)
	}
}

// TestParseFlagsWindowsPath 윈도우 절대 경로를 올바르게 파싱하는지 검증한다.
func TestParseFlagsWindowsPath(t *testing.T) {
	flags := parseFlags([]string{`C:\Users\test\doc.md`})

	if flags.filePath != `C:\Users\test\doc.md` {
		t.Errorf("filePath: want %q, got %q", `C:\Users\test\doc.md`, flags.filePath)
	}
}
