package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseArgsSuccess 정상적인 인자 파싱을 검증한다.
func TestParseArgsSuccess(t *testing.T) {
	args := []string{"winmdview", "test.md"}
	path, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs() 오류 발생: %v", err)
	}
	if path != "test.md" {
		t.Errorf("ParseArgs() = %q, want %q", path, "test.md")
	}
}

// TestParseArgsNoArgs 인자가 없는 경우를 검증한다 (REQ-E-003).
func TestParseArgsNoArgs(t *testing.T) {
	args := []string{"winmdview"}
	_, err := ParseArgs(args)
	if err == nil {
		t.Fatal("인자 없이 호출 시 에러가 반환되어야 한다")
	}
	if !strings.Contains(err.Error(), "사용법") {
		t.Errorf("에러 메시지에 사용법 안내가 포함되어야 함: %v", err)
	}
}

// TestParseArgsEmpty 빈 인자 배열을 검증한다.
func TestParseArgsEmpty(t *testing.T) {
	args := []string{}
	_, err := ParseArgs(args)
	if err == nil {
		t.Fatal("빈 인자에서 에러가 반환되어야 한다")
	}
}

// TestValidateFileNotFound 존재하지 않는 파일을 검증한다 (REQ-S-001).
func TestValidateFileNotFound(t *testing.T) {
	err := ValidateFile("nonexistent_file_12345.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일에 대해 에러가 반환되어야 한다")
	}
	if !strings.Contains(err.Error(), "찾을 수 없습니다") {
		t.Errorf("에러 메시지에 '찾을 수 없습니다'가 포함되어야 함: %v", err)
	}
}

// TestValidateFileIsDirectory 디렉토리를 파일로 열 수 없음을 검증한다.
func TestValidateFileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	err := ValidateFile(tmpDir)
	if err == nil {
		t.Fatal("디렉토리에 대해 에러가 반환되어야 한다")
	}
	if !strings.Contains(err.Error(), "디렉토리") {
		t.Errorf("에러 메시지에 '디렉토리'가 포함되어야 함: %v", err)
	}
}

// TestValidateFilePermissionError 파일 읽기 권한 오류를 검증한다 (REQ-S-002).
// 참고: Windows에서는 파일 권한 테스트가 제한적이므로 존재 여부 검증만 수행한다.

// TestValidateFileSuccess 정상적인 파일 검증을 확인한다.
func TestValidateFileSuccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	err := ValidateFile(tmpFile)
	if err != nil {
		t.Fatalf("정상 파일 검증 실패: %v", err)
	}
}

// TestReadFileSuccess 파일 읽기를 검증한다.
func TestReadFileSuccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	content := "# Hello World"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	data, err := ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile() 오류: %v", err)
	}
	if string(data) != content {
		t.Errorf("ReadFile() = %q, want %q", string(data), content)
	}
}

// TestReadFileNotFound 존재하지 않는 파일 읽기를 검증한다.
func TestReadFileNotFound(t *testing.T) {
	_, err := ReadFile("nonexistent_12345.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일 읽기 시 에러가 반환되어야 한다")
	}
}

// TestRenderPipelineSuccess 전체 렌더링 파이프라인을 검증한다.
func TestRenderPipelineSuccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Hello\n\nWorld"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderPipeline(tmpFile)
	if err != nil {
		t.Fatalf("RenderPipeline() 오류: %v", err)
	}

	checks := []string{
		"<!DOCTYPE html>",
		"Hello",
		"World",
		"markdown-body",
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Errorf("파이프라인 결과에 %q 가 포함되지 않음", want)
		}
	}
}

// TestRenderPipelineEmptyFile 빈 파일의 렌더링 파이프라인을 검증한다 (REQ-O-002).
func TestRenderPipelineEmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.md")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderPipeline(tmpFile)
	if err != nil {
		t.Fatalf("RenderPipeline() 빈 파일 오류: %v", err)
	}

	if !strings.Contains(html, "내용이 없습니다") {
		t.Error("빈 파일에 대해 '내용이 없습니다' 메시지가 포함되어야 한다")
	}
}

// TestRenderPipelineWhitespaceOnlyFile 공백만 있는 파일도 빈 파일로 처리됨을 검증한다.
func TestRenderPipelineWhitespaceOnlyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "whitespace.md")
	if err := os.WriteFile(tmpFile, []byte("   \n\n  \t  "), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderPipeline(tmpFile)
	if err != nil {
		t.Fatalf("RenderPipeline() 공백 파일 오류: %v", err)
	}

	if !strings.Contains(html, "내용이 없습니다") {
		t.Error("공백만 있는 파일에 대해 '내용이 없습니다' 메시지가 포함되어야 한다")
	}
}

// TestRenderPipelineFileNotFound 존재하지 않는 파일의 렌더링 파이프라인을 검증한다.
func TestRenderPipelineFileNotFound(t *testing.T) {
	_, err := RenderPipeline("nonexistent_12345.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일에 대해 에러가 반환되어야 한다")
	}
}
