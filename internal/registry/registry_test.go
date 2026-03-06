package registry_test

import (
	"strings"
	"testing"

	regpkg "github.com/AngeleyesTrue/WinMarkdownViewer/internal/registry"
	"golang.org/x/sys/windows/registry"
)

// 테스트 전용 레지스트리 경로 (프로덕션 키와 충돌 방지)
const testShellKeyPath = `Software\Classes\.md-test\shell\WinMarkdownViewer_Test`

func init() {
	// 테스트 실행 시 테스트 전용 경로로 오버라이드
	regpkg.SetShellKeyPath(testShellKeyPath)
}

// cleanupTestKeys 는 테스트 레지스트리 키를 정리한다.
func cleanupTestKeys(t *testing.T) {
	t.Helper()
	// command 하위 키 먼저 삭제
	_ = registry.DeleteKey(registry.CURRENT_USER, testShellKeyPath+`\command`)
	// shell 키 삭제
	_ = registry.DeleteKey(registry.CURRENT_USER, testShellKeyPath)
	// 상위 shell 키 삭제
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\.md-test\shell`)
	// .md-test 키 삭제
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\.md-test`)
}

// TestRegister_정상등록 은 Register가 올바른 레지스트리 키와 값을 생성하는지 검증한다.
func TestRegister_정상등록(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	exePath := `C:\Program Files\WinMarkdownViewer\winmdview.exe`
	err := regpkg.Register(exePath)
	if err != nil {
		t.Fatalf("Register() 오류: %v", err)
	}

	// shell 키의 기본값 확인: "마크다운 뷰어로 열기"
	k, err := registry.OpenKey(registry.CURRENT_USER, testShellKeyPath, registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("shell 키 열기 실패: %v", err)
	}
	defer k.Close()

	defaultVal, _, err := k.GetStringValue("")
	if err != nil {
		t.Fatalf("기본값 읽기 실패: %v", err)
	}
	if defaultVal != "마크다운 뷰어로 열기" {
		t.Errorf("기본값 = %q, want %q", defaultVal, "마크다운 뷰어로 열기")
	}

	// command 키의 기본값 확인
	ck, err := registry.OpenKey(registry.CURRENT_USER, testShellKeyPath+`\command`, registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("command 키 열기 실패: %v", err)
	}
	defer ck.Close()

	cmdVal, _, err := ck.GetStringValue("")
	if err != nil {
		t.Fatalf("command 기본값 읽기 실패: %v", err)
	}
	expectedCmd := `"` + exePath + `" "%1"`
	if cmdVal != expectedCmd {
		t.Errorf("command 값 = %q, want %q", cmdVal, expectedCmd)
	}
}

// TestUnregister_정상삭제 는 Unregister가 모든 레지스트리 키를 삭제하는지 검증한다.
func TestUnregister_정상삭제(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	exePath := `C:\Test\winmdview.exe`
	if err := regpkg.Register(exePath); err != nil {
		t.Fatalf("사전 등록 실패: %v", err)
	}

	err := regpkg.Unregister()
	if err != nil {
		t.Fatalf("Unregister() 오류: %v", err)
	}

	// command 키가 삭제되었는지 확인
	_, err = registry.OpenKey(registry.CURRENT_USER, testShellKeyPath+`\command`, registry.QUERY_VALUE)
	if err == nil {
		t.Error("command 키가 삭제되지 않았다")
	}

	// shell 키가 삭제되었는지 확인
	_, err = registry.OpenKey(registry.CURRENT_USER, testShellKeyPath, registry.QUERY_VALUE)
	if err == nil {
		t.Error("shell 키가 삭제되지 않았다")
	}
}

// TestIsRegistered_등록상태확인 은 IsRegistered가 올바른 상태를 반환하는지 검증한다.
func TestIsRegistered_등록상태확인(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	// 등록 전: false
	registered, err := regpkg.IsRegistered()
	if err != nil {
		t.Fatalf("IsRegistered() 오류: %v", err)
	}
	if registered {
		t.Error("등록 전에 IsRegistered() = true, want false")
	}

	// 등록 후: true
	if err := regpkg.Register(`C:\Test\winmdview.exe`); err != nil {
		t.Fatalf("Register() 오류: %v", err)
	}

	registered, err = regpkg.IsRegistered()
	if err != nil {
		t.Fatalf("IsRegistered() 오류: %v", err)
	}
	if !registered {
		t.Error("등록 후에 IsRegistered() = false, want true")
	}

	// 해제 후: false
	if err := regpkg.Unregister(); err != nil {
		t.Fatalf("Unregister() 오류: %v", err)
	}

	registered, err = regpkg.IsRegistered()
	if err != nil {
		t.Fatalf("IsRegistered() 오류: %v", err)
	}
	if registered {
		t.Error("해제 후에 IsRegistered() = true, want false")
	}
}

// TestRegister_재등록시경로업데이트 는 이미 등록된 상태에서 새 경로로 재등록하면
// exe 경로가 업데이트되는지 검증한다 (ACC-009).
func TestRegister_재등록시경로업데이트(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	oldPath := `C:\Old\winmdview.exe`
	newPath := `C:\New\winmdview.exe`

	if err := regpkg.Register(oldPath); err != nil {
		t.Fatalf("첫 번째 Register() 오류: %v", err)
	}

	if err := regpkg.Register(newPath); err != nil {
		t.Fatalf("두 번째 Register() 오류: %v", err)
	}

	// command 값이 새 경로로 업데이트되었는지 확인
	ck, err := registry.OpenKey(registry.CURRENT_USER, testShellKeyPath+`\command`, registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("command 키 열기 실패: %v", err)
	}
	defer ck.Close()

	cmdVal, _, err := ck.GetStringValue("")
	if err != nil {
		t.Fatalf("command 값 읽기 실패: %v", err)
	}
	expectedCmd := `"` + newPath + `" "%1"`
	if cmdVal != expectedCmd {
		t.Errorf("재등록 후 command 값 = %q, want %q", cmdVal, expectedCmd)
	}
}

// TestUnregister_미등록시메시지 는 등록되지 않은 상태에서 Unregister 호출 시
// 적절한 메시지가 포함된 에러를 반환하는지 검증한다 (ACC-010).
func TestUnregister_미등록시메시지(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	err := regpkg.Unregister()
	if err == nil {
		t.Fatal("미등록 상태에서 Unregister() 호출 시 에러가 반환되어야 한다")
	}
	if !strings.Contains(err.Error(), "등록되어 있지 않") {
		t.Errorf("에러 메시지에 '등록되어 있지 않'이 포함되어야 함: %v", err)
	}
}

// TestRegister_HKLM미접근 은 Register가 HKLM을 절대 건드리지 않는지 검증한다 (REQ-N-001).
func TestRegister_HKLM미접근(t *testing.T) {
	t.Cleanup(func() { cleanupTestKeys(t) })
	cleanupTestKeys(t)

	exePath := `C:\Test\winmdview.exe`
	if err := regpkg.Register(exePath); err != nil {
		t.Fatalf("Register() 오류: %v", err)
	}

	// HKLM에 테스트 키가 존재하지 않아야 한다
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, testShellKeyPath, registry.QUERY_VALUE)
	if err == nil {
		t.Error("HKLM에 레지스트리 키가 생성되었다 - REQ-N-001 위반")
	}
}
