package build_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// projectRoot 는 프로젝트 루트 디렉터리 경로를 반환한다.
func projectRoot(t *testing.T) string {
	t.Helper()
	// 테스트 파일이 프로젝트 루트에 위치하므로 runtime.Caller 로 경로를 얻는다.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 실패")
	}
	return filepath.Dir(filename)
}

// TestBuildScriptExists 는 build.ps1 파일이 프로젝트 루트에 존재하는지 검증한다.
func TestBuildScriptExists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "build.ps1")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("build.ps1 파일이 프로젝트 루트에 존재하지 않음")
	}
}

// TestBuildScriptContent 는 build.ps1 에 필수 빌드 플래그와 타겟이 포함되어 있는지 검증한다.
func TestBuildScriptContent(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "build.ps1")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("build.ps1 읽기 실패: %v", err)
	}
	content := string(data)

	t.Run("릴리스_빌드에_windowsgui_플래그_포함", func(t *testing.T) {
		if !strings.Contains(content, "-H windowsgui") {
			t.Error("build.ps1 에 '-H windowsgui' 플래그가 포함되어 있지 않음")
		}
	})

	t.Run("릴리스_빌드에_strip_플래그_포함", func(t *testing.T) {
		if !strings.Contains(content, "-s -w") {
			t.Error("build.ps1 에 '-s -w' strip 플래그가 포함되어 있지 않음")
		}
	})

	t.Run("빌드_출력_경로_winmdview_exe", func(t *testing.T) {
		if !strings.Contains(content, "winmdview.exe") {
			t.Error("build.ps1 에 'winmdview.exe' 출력 경로가 포함되어 있지 않음")
		}
	})

	t.Run("dev_빌드_출력_경로_winmdview_dev_exe", func(t *testing.T) {
		if !strings.Contains(content, "winmdview-dev.exe") {
			t.Error("build.ps1 에 'winmdview-dev.exe' dev 빌드 출력 경로가 포함되어 있지 않음")
		}
	})

	t.Run("cmd_winmdview_패키지_참조", func(t *testing.T) {
		if !strings.Contains(content, "./cmd/winmdview") {
			t.Error("build.ps1 에 './cmd/winmdview' 패키지 참조가 포함되어 있지 않음")
		}
	})

	// 타겟 검증: build, dev, test, clean, help
	targets := []string{"build", "dev", "test", "clean", "help"}
	for _, target := range targets {
		target := target
		t.Run("타겟_"+target+"_존재", func(t *testing.T) {
			// PowerShell switch 문에서 타겟 이름이 따옴표로 감싸져 있을 수 있다.
			hasTarget := strings.Contains(content, "\""+target+"\"") ||
				strings.Contains(content, "'"+target+"'")
			if !hasTarget {
				t.Errorf("build.ps1 에 '%s' 타겟이 포함되어 있지 않음", target)
			}
		})
	}
}

// TestBuildScriptDevBuildNoWindowsGui 는 dev 빌드가 windowsgui 플래그 없이 빌드되는지 검증한다.
// dev 빌드 섹션에서 ldflags 가 없거나 windowsgui 를 포함하지 않아야 한다.
func TestBuildScriptDevBuildNoWindowsGui(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "build.ps1")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("build.ps1 읽기 실패: %v", err)
	}
	content := string(data)

	// dev 빌드 명령어를 찾아서 windowsgui 가 없는지 확인한다.
	// dev 섹션은 "dev" 타겟과 다음 타겟 사이의 영역이다.
	lines := strings.Split(content, "\n")
	inDevSection := false
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		// dev 섹션 진입 감지
		if strings.Contains(lower, "\"dev\"") || strings.Contains(lower, "'dev'") {
			inDevSection = true
			continue
		}
		// 다른 타겟 섹션 진입 시 dev 섹션 종료
		if inDevSection && (strings.Contains(lower, "\"build\"") || strings.Contains(lower, "\"test\"") ||
			strings.Contains(lower, "\"clean\"") || strings.Contains(lower, "\"help\"") ||
			strings.Contains(lower, "'build'") || strings.Contains(lower, "'test'") ||
			strings.Contains(lower, "'clean'") || strings.Contains(lower, "'help'")) {
			break
		}
		// dev 섹션 내에서 go build 명령어에 windowsgui 가 없어야 한다.
		if inDevSection && strings.Contains(lower, "go build") {
			if strings.Contains(line, "-H windowsgui") {
				t.Error("dev 빌드에 '-H windowsgui' 플래그가 포함되어 있음 (콘솔 창이 표시되어야 함)")
			}
		}
	}
}

// TestBuildScriptTestTarget 은 test 타겟에 coverprofile 옵션이 포함되어 있는지 검증한다.
func TestBuildScriptTestTarget(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "build.ps1")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("build.ps1 읽기 실패: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "-coverprofile") {
		t.Error("build.ps1 test 타겟에 '-coverprofile' 옵션이 포함되어 있지 않음")
	}
}
