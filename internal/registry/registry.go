// Package registry 는 Windows 레지스트리를 통한 컨텍스트 메뉴 등록/해제를 관리한다.
// HKCU\Software\Classes\SystemFileAssociations\.md\shell\WinMarkdownViewer 경로에
// "마크다운 뷰어로 열기" 컨텍스트 메뉴를 등록한다.
// SystemFileAssociations를 사용하면 .md 파일의 기본 프로그램(VS Code 등)과 관계없이
// 컨텍스트 메뉴가 항상 표시된다.
// @MX:NOTE: [AUTO] HKLM은 절대 사용하지 않음 (REQ-N-001)
package registry

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// shellKeyPath 는 컨텍스트 메뉴 등록에 사용되는 레지스트리 경로이다.
// SystemFileAssociations를 사용하여 기본 프로그램과 무관하게 메뉴가 표시되도록 한다.
// 테스트에서 SetShellKeyPath로 오버라이드할 수 있다.
var shellKeyPath = `Software\Classes\SystemFileAssociations\.md\shell\WinMarkdownViewer`

// menuLabel 은 컨텍스트 메뉴에 표시되는 텍스트이다.
const menuLabel = "마크다운 뷰어로 열기"

// SetShellKeyPath 는 레지스트리 경로를 변경한다 (테스트 전용).
func SetShellKeyPath(path string) {
	shellKeyPath = path
}

// Register 는 Windows 탐색기 컨텍스트 메뉴에 "마크다운 뷰어로 열기" 항목을 등록한다.
// exePath 는 실행 파일의 전체 경로이다.
// 이미 등록된 경우 exe 경로를 업데이트한다 (ACC-009).
func Register(exePath string) error {
	// shell 키 생성 (이미 존재하면 열기)
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		shellKeyPath,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("레지스트리 shell 키 생성 실패: %w", err)
	}
	defer k.Close()

	// 기본값 설정: 메뉴 표시 텍스트
	if err := k.SetStringValue("", menuLabel); err != nil {
		return fmt.Errorf("메뉴 레이블 설정 실패: %w", err)
	}

	// command 하위 키 생성
	ck, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		shellKeyPath+`\command`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("레지스트리 command 키 생성 실패: %w", err)
	}
	defer ck.Close()

	// command 기본값 설정: "exe경로" "%1"
	cmdValue := fmt.Sprintf(`"%s" "%%1"`, exePath)
	if err := ck.SetStringValue("", cmdValue); err != nil {
		return fmt.Errorf("명령어 값 설정 실패: %w", err)
	}

	return nil
}

// Unregister 는 컨텍스트 메뉴 등록을 해제한다.
// 등록되어 있지 않은 경우 에러를 반환한다 (ACC-010).
func Unregister() error {
	// 먼저 등록 상태 확인
	registered, err := IsRegistered()
	if err != nil {
		return fmt.Errorf("등록 상태 확인 실패: %w", err)
	}
	if !registered {
		return fmt.Errorf("컨텍스트 메뉴가 등록되어 있지 않습니다")
	}

	// command 하위 키 먼저 삭제 (하위 키가 있으면 상위 삭제 불가)
	if err := registry.DeleteKey(registry.CURRENT_USER, shellKeyPath+`\command`); err != nil {
		return fmt.Errorf("command 키 삭제 실패: %w", err)
	}

	// shell 키 삭제
	if err := registry.DeleteKey(registry.CURRENT_USER, shellKeyPath); err != nil {
		return fmt.Errorf("shell 키 삭제 실패: %w", err)
	}

	return nil
}

// IsRegistered 는 컨텍스트 메뉴가 등록되어 있는지 확인한다.
func IsRegistered() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, shellKeyPath, registry.QUERY_VALUE)
	if err != nil {
		// 키가 존재하지 않으면 미등록 상태
		return false, nil
	}
	k.Close()
	return true, nil
}
