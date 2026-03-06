---
spec_id: SPEC-UI-001
type: acceptance-criteria
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-UI-001 수용 기준

## 1. 핵심 시나리오

### ACC-001: 기본 Markdown 파일 열기 및 렌더링

```gherkin
Given 유효한 UTF-8 Markdown 파일 "test.md"가 존재한다
  And 파일에 제목, 본문, 코드 블록이 포함되어 있다
When 사용자가 "winmdview.exe test.md" 명령을 실행한다
Then WebView2 윈도우가 1024x768 크기로 열린다
  And 윈도우 타이틀에 "test.md - WinMarkdownViewer"가 표시된다
  And Markdown이 GitHub 스타일 CSS가 적용된 HTML로 렌더링된다
  And 코드 블록에 구문 강조가 적용된다
```

### ACC-002: GFM 확장 문법 지원

```gherkin
Given Markdown 파일에 다음 GFM 요소가 포함되어 있다:
  | 요소 | 예시 |
  | 테이블 | `| A | B |` 형식 |
  | 취소선 | `~~삭제된 텍스트~~` |
  | 자동 링크 | `https://example.com` |
  | 태스크 리스트 | `- [x] 완료` |
When 해당 파일을 렌더링한다
Then 테이블이 HTML <table>로 변환된다
  And 취소선이 <del> 태그로 변환된다
  And URL이 클릭 가능한 <a> 태그로 변환된다
  And 태스크 리스트가 체크박스로 변환된다
```

### ACC-003: 윈도우 종료 시 정상 종료

```gherkin
Given WebView2 윈도우가 열려 있다
When 사용자가 윈도우의 닫기 버튼(X)을 클릭한다
Then WebView2 리소스가 해제된다
  And 프로세스가 exit code 0으로 종료된다
  And 메모리 누수 없이 정리된다
```

---

## 2. 에러 처리 시나리오

### ACC-004: 인자 없이 실행

```gherkin
Given 사용자가 명령줄 인자 없이 프로그램을 실행한다
When "winmdview.exe"만 실행한다
Then 표준 에러(stderr)에 사용법 안내가 출력된다:
  """
  사용법: winmdview <파일경로.md>
  """
  And 프로세스가 exit code 1로 종료된다
  And WebView2 윈도우는 생성되지 않는다
```

### ACC-005: 외부 리소스 미사용 검증

```gherkin
Given 프로그램이 빌드되어 있다
When 네트워크 연결 없는 환경에서 Markdown 파일을 열어본다
Then 렌더링이 정상적으로 완료된다
  And CSS 스타일이 정상 적용된다
  And 구문 강조가 정상 동작한다
  And 외부 HTTP 요청이 발생하지 않는다
```

### ACC-006: 파일 미존재 시 에러 메시지

```gherkin
Given "nonexistent.md" 파일이 존재하지 않는다
When 사용자가 "winmdview.exe nonexistent.md"를 실행한다
Then 표준 에러에 "파일을 찾을 수 없습니다: nonexistent.md" 메시지가 출력된다
  And 프로세스가 exit code 1로 종료된다
```

```gherkin
Given "readonly.md" 파일에 읽기 권한이 없다
When 사용자가 "winmdview.exe readonly.md"를 실행한다
Then 표준 에러에 파일 읽기 권한 관련 에러 메시지가 출력된다
  And 프로세스가 exit code 1로 종료된다
```

### ACC-007: WebView2 Runtime 미설치

```gherkin
Given WebView2 Runtime이 설치되지 않은 Windows 환경이다
When 사용자가 Markdown 파일을 열려고 한다
Then 표준 에러에 WebView2 Runtime 설치 안내 메시지가 출력된다
  And 설치 다운로드 URL이 안내에 포함된다
  And 프로세스가 exit code 1로 종료된다
```

---

## 3. 엣지 케이스

### ACC-008: 빈 Markdown 파일

```gherkin
Given "empty.md" 파일이 존재하고 내용이 비어 있다 (0 바이트)
When 사용자가 "winmdview.exe empty.md"를 실행한다
Then WebView2 윈도우가 열린다
  And 빈 HTML 페이지가 표시된다 (또는 "내용이 없습니다" 안내)
  And 프로그램이 crash하지 않는다
```

### ACC-009: 대용량 Markdown 파일

```gherkin
Given "large.md" 파일이 5MB 크기의 Markdown 내용을 포함한다
When 사용자가 "winmdview.exe large.md"를 실행한다
Then 파일이 성공적으로 렌더링된다
  And WebView2 윈도우에 HTML이 표시된다
  And 프로그램이 메모리 부족으로 crash하지 않는다
```

### ACC-010: 특수 문자가 포함된 파일 경로

```gherkin
Given 파일 경로에 공백이 포함된 "내 문서/readme.md" 파일이 존재한다
When 사용자가 "winmdview.exe '내 문서/readme.md'"를 실행한다
Then 파일이 정상적으로 읽히고 렌더링된다
```

### ACC-011: Markdown이 아닌 확장자 파일

```gherkin
Given "readme.txt" 파일이 Markdown 형식의 내용을 포함하고 있다
When 사용자가 "winmdview.exe readme.txt"를 실행한다
Then 파일 내용이 Markdown으로 파싱되어 렌더링된다
  And 확장자에 관계없이 내용을 Markdown으로 처리한다
```

---

## 4. 빌드 및 실행 검증

### ACC-012: 빌드 성공

```gherkin
Given Go 1.26.0이 설치되어 있다
  And 모든 소스 파일이 작성되어 있다
When "go build ./cmd/winmdview/" 명령을 실행한다
Then 빌드가 에러 없이 성공한다
  And "winmdview.exe" 바이너리가 생성된다
  And CGO가 사용되지 않는다 (CGO_ENABLED=0)
```

### ACC-013: 테스트 통과

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test ./..." 명령을 실행한다
Then 모든 테스트가 통과한다
  And Markdown 렌더링 테스트가 포함된다
  And CLI 인자 파싱 테스트가 포함된다
  And 테스트 커버리지가 85% 이상이다
```

### ACC-014: 경쟁 조건 검사

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test -race ./..." 명령을 실행한다
Then 데이터 경쟁(race condition)이 감지되지 않는다
```

---

## 5. Quality Gate (품질 게이트)

| 항목 | 기준 | 검증 방법 |
|------|------|-----------|
| 빌드 성공 | `go build` 에러 없음 | `go build ./cmd/winmdview/` |
| 테스트 통과 | 모든 테스트 PASS | `go test ./...` |
| 테스트 커버리지 | 85% 이상 (GUI 제외) | `go test -coverprofile=coverage.out ./internal/...` |
| 경쟁 조건 | 감지 없음 | `go test -race ./...` |
| 정적 분석 | 경고 없음 | `go vet ./...` |
| CGO 미사용 | CGO_ENABLED=0 빌드 성공 | `CGO_ENABLED=0 go build ./cmd/winmdview/` |
| 코드 품질 | lint 통과 | `golangci-lint run` (설치 시) |

---

## 6. Definition of Done (완료 정의)

- [ ] 모든 요구사항(REQ-*)에 대응하는 구현 코드 존재
- [ ] 모든 수용 기준(ACC-*)의 Given-When-Then 시나리오 검증 완료
- [ ] `go build ./cmd/winmdview/` 성공 (CGO_ENABLED=0)
- [ ] `go test ./...` 모든 테스트 통과
- [ ] `go test -race ./...` 경쟁 조건 미감지
- [ ] `go vet ./...` 경고 없음
- [ ] 테스트 커버리지 85% 이상 (internal/ 패키지 기준)
- [ ] 실제 Markdown 파일로 수동 실행 검증 완료
- [ ] WebView2 윈도우에서 GitHub 스타일 렌더링 확인
