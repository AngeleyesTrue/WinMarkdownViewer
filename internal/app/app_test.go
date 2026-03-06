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
// Windows에서는 Unix 스타일 파일 권한(chmod 0000)이 지원되지 않으므로 건너뛴다.
// Unix 환경에서 실행 시 권한 오류를 올바르게 감지하는지 검증한다.
func TestValidateFilePermissionError(t *testing.T) {
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Windows에서는 Unix 파일 권한 테스트를 지원하지 않음")
	}
	tmpFile := filepath.Join(t.TempDir(), "noperm.md")
	if err := os.WriteFile(tmpFile, []byte("# Test"), 0000); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}
	err := ValidateFile(tmpFile)
	if err == nil {
		t.Fatal("권한 오류가 반환되어야 한다")
	}
	if !strings.Contains(err.Error(), "권한") {
		t.Errorf("에러 메시지에 '권한'이 포함되어야 함: %v", err)
	}
}

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

// TestRenderMarkdownSuccess 마크다운 렌더링 함수가 HTML 본문만 반환하는지 검증한다.
func TestRenderMarkdownSuccess(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Hello\n\nWorld"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderMarkdown(tmpFile)
	if err != nil {
		t.Fatalf("RenderMarkdown() 오류: %v", err)
	}

	// 전체 HTML 문서가 아닌 본문만 반환해야 한다
	if strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("RenderMarkdown은 전체 HTML 문서가 아닌 본문만 반환해야 한다")
	}
	if !strings.Contains(html, "Hello") {
		t.Error("렌더링 결과에 'Hello'가 포함되어야 한다")
	}
}

// TestRenderMarkdownEmptyFile 빈 파일의 마크다운 렌더링을 검증한다.
func TestRenderMarkdownEmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.md")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderMarkdown(tmpFile)
	if err != nil {
		t.Fatalf("RenderMarkdown() 빈 파일 오류: %v", err)
	}

	if !strings.Contains(html, "내용이 없습니다") {
		t.Error("빈 파일에 대해 '내용이 없습니다' 메시지가 포함되어야 한다")
	}
}

// TestRenderMarkdownFileNotFound 존재하지 않는 파일을 검증한다.
func TestRenderMarkdownFileNotFound(t *testing.T) {
	_, err := RenderMarkdown("nonexistent_12345.md")
	if err == nil {
		t.Fatal("존재하지 않는 파일에 대해 에러가 반환되어야 한다")
	}
}

// TestRenderMarkdownWhitespaceFile 공백만 있는 파일의 렌더링을 검증한다.
func TestRenderMarkdownWhitespaceFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ws.md")
	if err := os.WriteFile(tmpFile, []byte("  \n  \t "), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	html, err := RenderMarkdown(tmpFile)
	if err != nil {
		t.Fatalf("RenderMarkdown() 오류: %v", err)
	}
	if !strings.Contains(html, "내용이 없습니다") {
		t.Error("공백만 있는 파일에 대해 '내용이 없습니다' 메시지가 포함되어야 한다")
	}
}

// TestReadFileNotExist 존재하지 않는 파일 읽기가 적절한 에러를 반환하는지 검증한다.
func TestReadFileNotExist(t *testing.T) {
	_, err := ReadFile(filepath.Join(t.TempDir(), "no_such_file.md"))
	if err == nil {
		t.Fatal("존재하지 않는 파일 읽기 시 에러가 반환되어야 한다")
	}
}

// TestValidateFileValid 정상 파일의 검증을 다시 확인한다.
func TestValidateFileValidMarkdown(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "valid.md")
	if err := os.WriteFile(tmpFile, []byte("# 정상 마크다운 파일"), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	if err := ValidateFile(tmpFile); err != nil {
		t.Fatalf("정상 파일 검증 실패: %v", err)
	}
}
