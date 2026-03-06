---
spec-id: SPEC-CONFIG-001
title: "User Configuration - Acceptance Criteria"
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-CONFIG-001: 사용자 설정 - 수용 기준

## 1. 설정 파일 자동 생성 시나리오

### AC-CFG-001: 첫 실행 시 설정 파일 자동 생성

```gherkin
Given WinMarkdownViewer가 처음 실행되어 "%APPDATA%\WinMarkdownViewer\" 디렉토리가 존재하지 않을 때
When WinMarkdownViewer가 시작되면
Then "%APPDATA%\WinMarkdownViewer\" 디렉토리가 생성되어야 한다
And "%APPDATA%\WinMarkdownViewer\config.json" 파일이 기본값으로 생성되어야 한다
And config.json의 내용이 유효한 JSON 형식이어야 한다
```

### AC-CFG-002: 시작 시 설정 로드

```gherkin
Given "%APPDATA%\WinMarkdownViewer\config.json"에 fontSize가 20으로 설정되어 있을 때
When WinMarkdownViewer가 시작되면
Then body 요소의 CSS font-size 속성이 20px로 설정된다
And HTML 템플릿에서 body { font-size: {{.FontSize}}px } 매핑으로 적용된다
```

### AC-CFG-003: JSON 형식 검증

```gherkin
Given "%APPDATA%\WinMarkdownViewer\config.json"이 유효한 JSON 파일일 때
When WinMarkdownViewer가 설정을 로드하면
Then 모든 설정 항목이 정상적으로 파싱되어야 한다
And 프로그램이 오류 없이 시작되어야 한다
```

---

## 2. 설정 항목 시나리오

### AC-CFG-010: 기본 설정 항목 지원

```gherkin
Given config.json이 존재하지 않아 기본값으로 생성될 때
When 생성된 config.json의 내용을 확인하면
Then theme 값이 "system"이어야 한다
And fontSize 값이 16이어야 한다
And windowWidth 값이 1024여야 한다
And windowHeight 값이 768이어야 한다
And windowX 값이 -1이어야 한다
And windowY 값이 -1이어야 한다
And customCSS 값이 빈 문자열이어야 한다
And lastOpenedFile 값이 빈 문자열이어야 한다
```

---

## 3. 설정 검증 시나리오

### AC-CFG-020: 범위 벗어난 값 보정

```gherkin
Given config.json에 fontSize가 30으로 설정되어 있을 때 (허용 범위: 14-24)
When WinMarkdownViewer가 설정을 로드하면
Then fontSize가 기본값 16으로 보정되어야 한다
And 다른 설정 항목은 변경되지 않아야 한다
And 프로그램이 정상적으로 시작되어야 한다

Given config.json에 theme이 "blue"로 설정되어 있을 때 (허용값: light, dark, system)
When WinMarkdownViewer가 설정을 로드하면
Then theme이 기본값 "system"으로 보정되어야 한다
```

### AC-CFG-021: 손상된 JSON 파일 복구

```gherkin
Given config.json의 내용이 "{ invalid json }"일 때
When WinMarkdownViewer가 설정을 로드하면
Then 기존 config.json이 config.json.bak으로 백업되어야 한다
And 새로운 config.json이 기본값으로 생성되어야 한다
And 프로그램이 정상적으로 시작되어야 한다
```

### AC-CFG-022: 잘못된 설정으로 인한 크래시 방지

```gherkin
Given config.json에 windowWidth가 -500으로 설정되어 있을 때
When WinMarkdownViewer가 설정을 로드하면
Then 프로그램이 크래시하지 않아야 한다
And windowWidth가 기본값 1024로 보정되어야 한다
```

---

## 4. 설정 저장 시나리오

### AC-CFG-030: 창 닫기 시 크기/위치 자동 저장

```gherkin
Given WinMarkdownViewer 창의 크기가 1280x800이고 위치가 (100, 50)일 때
When 사용자가 창을 닫으면
Then config.json의 windowWidth가 1280으로 저장되어야 한다
And config.json의 windowHeight가 800으로 저장되어야 한다
And config.json의 windowX가 100으로 저장되어야 한다
And config.json의 windowY가 50으로 저장되어야 한다

Given WinMarkdownViewer를 다시 실행할 때
When 프로그램이 시작되면
Then 창 크기가 1280x800으로 복원되어야 한다
And 창 위치가 (100, 50)으로 복원되어야 한다
```

### AC-CFG-031: 파일 열기 시 경로 저장

```gherkin
Given 사용자가 "C:\docs\README.md" 파일을 WinMarkdownViewer로 열 때
When 파일이 성공적으로 로드되면
Then config.json의 lastOpenedFile이 "C:\docs\README.md"로 저장되어야 한다
```

### AC-CFG-032: 들여쓰기된 JSON 저장

```gherkin
Given 설정 변경이 발생하여 config.json을 저장할 때
When 저장된 config.json을 텍스트 에디터에서 열면
Then JSON이 2-space 들여쓰기로 포맷팅되어 있어야 한다
And 사람이 읽기 쉬운 형식이어야 한다
```

---

## 5. API 시나리오

### AC-CFG-040: 설정 API 제공

```gherkin
Given internal/config 패키지가 import된 상태에서
When config.Load()를 호출하면
Then *Config 포인터와 nil error를 반환해야 한다 (정상 케이스)
And 반환된 Config의 모든 필드가 유효한 값을 가져야 한다

Given 유효한 Config 인스턴스가 있을 때
When config.Save(cfg)를 호출하면
Then config.json 파일이 해당 값으로 갱신되어야 한다
And nil error를 반환해야 한다

Given 설정 파일이 없는 상태에서
When config.Default()를 호출하면
Then 모든 필드가 기본값으로 채워진 Config를 반환해야 한다
```

### AC-CFG-041: 동시성 안전 보장

```gherkin
Given 두 개의 goroutine이 동시에 설정을 읽고 쓸 때
When go test -race ./internal/config/... 명령을 실행하면
Then 데이터 레이스가 감지되지 않아야 한다
And 모든 테스트가 통과해야 한다
```

---

## 6. Edge Case 시나리오

### EC-001: %APPDATA% 접근 불가

```gherkin
Given ConfigPath 오버라이드로 임시 디렉토리를 지정한 후
And 해당 디렉토리에 읽기 전용 권한을 설정하였을 때
When WinMarkdownViewer가 시작되면
Then 프로그램이 크래시하지 않아야 한다
And 메모리 내 기본 설정으로 정상 동작해야 한다
And 설정 저장 시도 시 에러를 로깅해야 한다
```

### EC-002: 설정 파일에 알 수 없는 필드 존재

```gherkin
Given config.json에 "unknownField": "value"라는 알 수 없는 필드가 있을 때
When WinMarkdownViewer가 설정을 로드하면
Then 알 수 없는 필드는 무시되어야 한다
And 알려진 설정 항목은 정상 로드되어야 한다
And 프로그램이 정상 시작되어야 한다
```

### EC-003: 설정 파일에 일부 필드만 존재

```gherkin
Given config.json에 {"theme": "dark"} 만 있을 때 (나머지 필드 누락)
When WinMarkdownViewer가 설정을 로드하면
Then theme은 "dark"가 적용되어야 한다
And 누락된 필드(fontSize, windowWidth 등)는 기본값이 적용되어야 한다
```

### EC-004: customCSS 경로의 파일이 존재하지 않음

```gherkin
Given config.json에 customCSS가 "C:\nonexistent\style.css"로 설정되어 있을 때
When WinMarkdownViewer가 설정을 로드하면
Then customCSS 값은 빈 문자열로 보정되어야 한다
And 기본 CSS가 적용되어야 한다
```

### EC-005: 동시 프로세스 간 설정 충돌

```gherkin
Given 두 개의 WinMarkdownViewer 인스턴스가 실행 중일 때
When 두 인스턴스가 동시에 다른 창 크기로 설정을 저장하면
Then 마지막 쓰기가 유지되어야 한다 (last-write-wins)
And 설정 파일이 손상되지 않아야 한다
```

---

## 7. Quality Gates

### 7.1 테스트 품질
- [ ] `go test -race ./internal/config/...` 모든 테스트 통과
- [ ] 데이터 레이스 없음 확인
- [ ] 테스트 커버리지 85% 이상
- [ ] 테이블 기반 테스트로 경계값 검증

### 7.2 코드 품질
- [ ] `go vet ./internal/config/...` 경고 없음
- [ ] 모든 공개 함수에 GoDoc 주석 존재
- [ ] 에러 처리가 모든 I/O 경로에 적용됨

### 7.3 통합 품질
- [ ] 뷰어와 통합 시 설정이 정상 적용됨
- [ ] 창 닫기/열기 사이클에서 설정 유지 확인
- [ ] 손상된 파일에서 정상 복구 확인

### 7.4 Definition of Done
- [ ] 모든 AC(수용 기준) 시나리오 통과
- [ ] 모든 Edge Case 시나리오 통과
- [ ] `go test -race` 통과
- [ ] 테스트 커버리지 85% 이상
- [ ] 코드 리뷰 완료
- [ ] GoDoc 주석 완성
