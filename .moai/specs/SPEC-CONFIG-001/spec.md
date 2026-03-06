---
id: SPEC-CONFIG-001
title: "User Configuration - JSON-based Settings System"
version: 1.1.0
status: completed
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P2
lifecycle: spec-first
tags: [config, json, settings, appdata, persistence]
---

# SPEC-CONFIG-001: 사용자 설정 - JSON 기반 설정 시스템

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- Go 1.26.0 (CGO 불필요)

### 1.2 기술 스택
- encoding/json (stdlib) - JSON 직렬화/역직렬화
- os (stdlib) - 파일 시스템 및 %APPDATA% 경로 접근
- filepath (stdlib) - 경로 조작
- sync (stdlib) - 동시성 안전한 설정 접근

### 1.3 디렉토리 구조
```
WinMarkdownViewer/
  internal/
    config/
      config.go              # 설정 구조체 및 기본값 정의
      config_test.go         # 설정 단위 테스트
      loader.go              # 설정 파일 읽기/쓰기
      loader_test.go         # 로더 테스트
      validator.go           # 설정값 검증 로직
      validator_test.go      # 검증 테스트
```

### 1.4 설정 파일 위치
- 경로: `%APPDATA%\WinMarkdownViewer\config.json`
- Windows 예시: `C:\Users\{username}\AppData\Roaming\WinMarkdownViewer\config.json`

### 1.5 선행 조건
- SPEC-UI-001 완료 (WebView2 기반 뷰어가 동작해야 설정이 반영됨)

---

## 2. Assumptions (가정)

### 2.1 기술적 가정
- A1: %APPDATA% 환경변수가 Windows에서 항상 유효한 경로를 반환한다
- A2: 설정 파일의 크기는 항상 수 KB 이하이므로 전체 파일을 메모리에 로드해도 문제없다
- A3: 동시에 두 개 이상의 WinMarkdownViewer 프로세스가 같은 설정 파일에 접근할 수 있다
- A3-1: SPEC-WIN-001 미구현 시 다중 프로세스가 동시에 설정 파일을 접근할 수 있으며, last-write-wins 정책을 적용한다
- A4: encoding/json의 기본 직렬화가 설정 파일 용도로 충분하다

### 2.2 비즈니스 가정
- A5: 사용자는 config.json 파일을 직접 편집할 수 있다 (텍스트 에디터)
- A6: 설정 항목의 수는 향후 확장되지만, 현재 버전에서는 7개 항목으로 시작한다
- A7: 설정 파일에 새 필드가 추가되면 기본값으로 자동 채워진다. 스키마 버전 관리(major 변경)는 범위 외

### 2.3 범위 외 (Out of Scope)
- 설정 UI (WebView2 내 설정 페이지 또는 키보드 단축키)
- 설정 파일 마이그레이션 (스키마 버전 관리)
- 클라우드 동기화
- 레지스트리 기반 설정 저장
- 설정 암호화

---

## 3. Requirements (요구사항)

### 3.1 설정 파일 관리

**REQ-CFG-001** (Event-Driven)
**WHEN** WinMarkdownViewer가 처음 실행되어 설정 파일이 존재하지 않을 때 **THEN** 시스템은 `%APPDATA%\WinMarkdownViewer\config.json` 파일을 기본값으로 자동 생성해야 한다.

**REQ-CFG-002** (Event-Driven)
**WHEN** WinMarkdownViewer가 실행될 때 **THEN** 시스템은 `%APPDATA%\WinMarkdownViewer\config.json`에서 설정을 읽어 적용해야 한다.

**REQ-CFG-003** (Ubiquitous)
시스템은 **항상** 설정 파일이 유효한 JSON 형식인지 검증해야 한다.

### 3.2 설정 항목

**REQ-CFG-010** (Ubiquitous)
시스템은 **항상** 다음 설정 항목을 지원해야 한다:

| 항목 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| theme | string | "system" | "light", "dark", "system" 중 하나 |
| fontSize | int | 16 | 14-24 범위의 정수 |
| windowWidth | int | 1024 | 창 너비 (픽셀) |
| windowHeight | int | 768 | 창 높이 (픽셀) |
| windowX | int | -1 | 창 X 좌표 (-1: 시스템 기본) |
| windowY | int | -1 | 창 Y 좌표 (-1: 시스템 기본) |
| customCSS | string | "" | 사용자 정의 CSS 파일 경로 |
| lastOpenedFile | string | "" | 마지막으로 열었던 파일 경로 |

### 3.3 설정 검증

**REQ-CFG-020** (State-Driven)
**IF** 설정 파일의 값이 허용 범위를 벗어나면 **THEN** 시스템은 해당 항목만 기본값으로 복원해야 한다.

**REQ-CFG-021** (State-Driven)
**IF** 설정 파일이 손상되어 JSON 파싱이 실패하면 **THEN** 시스템은 전체 설정을 기본값으로 초기화해야 한다.

**REQ-CFG-022** (Unwanted)
시스템은 잘못된 설정값으로 인해 크래시**하지 않아야 한다**.

### 3.4 설정 저장

**REQ-CFG-030** (Event-Driven)
**WHEN** 사용자가 WinMarkdownViewer 창을 닫을 때 **THEN** 시스템은 현재 창 크기(windowWidth, windowHeight)와 위치(windowX, windowY)를 설정 파일에 자동 저장해야 한다.

**REQ-CFG-031** (Event-Driven)
**WHEN** 사용자가 .md 파일을 열 때 **THEN** 시스템은 해당 파일 경로를 lastOpenedFile에 저장해야 한다.

**REQ-CFG-032** (Ubiquitous)
시스템은 **항상** 설정 저장 시 들여쓰기된 JSON 형식(indent: 2 spaces)으로 파일을 작성해야 한다.

### 3.5 설정 API

**REQ-CFG-040** (Ubiquitous)
시스템은 **항상** 다음 API를 제공해야 한다:
- `Load() (*Config, error)`: 설정 파일에서 읽기
- `Save(cfg *Config) error`: 설정 파일에 쓰기
- `Default() *Config`: 기본 설정 반환
- `Validate(cfg *Config) *Config`: 검증 후 보정된 설정 반환

**REQ-CFG-041** (Ubiquitous)
시스템은 **항상** 설정 읽기/쓰기 시 동시성 안전(thread-safe)을 보장해야 한다.

---

## 4. Specifications (명세)

### 4.1 Config 구조체

```go
// 구조체 설계 방향 (구현 코드가 아닌 인터페이스 명세)
type Config struct {
    Theme          string `json:"theme"`
    FontSize       int    `json:"fontSize"`
    WindowWidth    int    `json:"windowWidth"`
    WindowHeight   int    `json:"windowHeight"`
    WindowX        int    `json:"windowX"`
    WindowY        int    `json:"windowY"`
    CustomCSS      string `json:"customCSS"`
    LastOpenedFile string `json:"lastOpenedFile"`
}
```

### 4.2 설정 파일 경로 결정 로직

```
1. os.UserConfigDir() 호출 -> %APPDATA% 경로 반환
2. 서브디렉토리 "WinMarkdownViewer" 추가
3. 디렉토리 미존재 시 os.MkdirAll로 생성
4. "config.json" 파일 경로 반환
```

### 4.3 설정 로드 흐름

```
Load() 호출
  -> 설정 파일 경로 결정
  -> 파일 존재 확인
     -> 미존재: Default() 반환 + 파일 자동 생성
     -> 존재: 파일 읽기
        -> JSON 파싱 실패: Default() 반환 + 기존 파일 백업(.bak)
        -> JSON 파싱 성공: Validate() 적용 -> 보정된 Config 반환
```

### 4.4 설정 검증 규칙

| 항목 | 검증 규칙 | 위반 시 기본값 |
|------|----------|---------------|
| theme | "light", "dark", "system" 중 하나 | "system" |
| fontSize | 14 <= n <= 24 | 16 |
| windowWidth | 320 <= n <= 7680 | 1024 |
| windowHeight | 240 <= n <= 4320 | 768 |
| windowX | -1(자동) 또는 0 이상의 정수. 화면 범위 검증은 viewer.go에서 수행 (config에서는 범위 검사 불필요) | -1 |
| windowY | -1(자동) 또는 0 이상의 정수. 화면 범위 검증은 viewer.go에서 수행 (config에서는 범위 검사 불필요) | -1 |
| customCSS | 절대 경로만 허용. 로드 시점에 파일 존재 검증. 렌더링 시점에 파일 미존재 시 기본 CSS로 fallback | "" |
| lastOpenedFile | 검증 없음 (경로만 저장) | "" |

### 4.5 동시성 전략

- `sync.RWMutex`를 사용하여 Config 접근 보호
- 읽기 작업: RLock/RUnlock
- 쓰기 작업: Lock/Unlock
- 파일 I/O: os.O_CREATE|os.O_WRONLY|os.O_TRUNC 플래그로 원자적 쓰기

---

## 5. Traceability (추적성)

| 요구사항 ID | plan.md 참조 | acceptance.md 참조 | 관련 파일 |
|-------------|-------------|-------------------|-----------|
| REQ-CFG-001 | Task 1.1, 1.2 | AC-CFG-001 | internal/config/loader.go |
| REQ-CFG-002 | Task 1.2 | AC-CFG-002 | internal/config/loader.go |
| REQ-CFG-003 | Task 1.3 | AC-CFG-003 | internal/config/validator.go |
| REQ-CFG-010 | Task 1.1 | AC-CFG-010 | internal/config/config.go |
| REQ-CFG-020 | Task 1.3 | AC-CFG-020 | internal/config/validator.go |
| REQ-CFG-021 | Task 1.3 | AC-CFG-021 | internal/config/validator.go |
| REQ-CFG-022 | Task 1.3 | AC-CFG-022 | internal/config/validator.go |
| REQ-CFG-030 | Task 2.1 | AC-CFG-030 | internal/config/loader.go |
| REQ-CFG-031 | Task 2.1 | AC-CFG-031 | internal/config/loader.go |
| REQ-CFG-032 | Task 1.2 | AC-CFG-032 | internal/config/loader.go |
| REQ-CFG-040 | Task 1.1, 1.2 | AC-CFG-040 | internal/config/config.go, loader.go |
| REQ-CFG-041 | Task 1.4 | AC-CFG-041 | internal/config/loader.go |

---

## Implementation Notes (구현 노트)

### 구현 일자
- 2026-03-06

### 구현 커밋
- `3202f3d` feat(config): SPEC-CONFIG-001 JSON 기반 사용자 설정 시스템 구현

### 계획 대비 변경 사항

**구조적 변경:**
- 계획대로 `config.go`, `loader.go`, `validator.go` 3파일 구조로 구현 완료
- 각각의 테스트 파일(`config_test.go`, `loader_test.go`, `validator_test.go`) 포함

**범위 변경:**
- Task 2.1 (창 닫기 시 자동 저장): SPEC-WATCH-001의 app.go 레이어에서 통합 처리
- Task 2.3 (시작 시 설정 적용): viewer.go 수정으로 반영
- Secondary/Tertiary Goal 전체 구현 완료

### 테스트 커버리지
- config 패키지: 88.9%
- 테이블 기반 테스트로 경계값 검증 완료
- 전체 테스트 PASS, go vet 클린

### 알려진 제한사항
- 다중 프로세스 동시 접근 시 last-write-wins 정책 적용 (파일 잠금 미구현)
- 설정 스키마 버전 관리(마이그레이션)는 범위 외
