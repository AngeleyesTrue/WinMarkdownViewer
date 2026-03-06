---
spec-id: SPEC-CONFIG-001
title: "User Configuration - Implementation Plan"
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
development-mode: tdd
---

# SPEC-CONFIG-001: 사용자 설정 - 구현 계획

## 1. TDD 접근 방식

### 1.1 RED-GREEN-REFACTOR 사이클

Go의 테스트 도구(`go test -race`)를 활용하여 다음 순서로 구현한다:

1. **RED**: 실패하는 테스트 작성 (config_test.go, loader_test.go, validator_test.go)
2. **GREEN**: 테스트를 통과하는 최소한의 코드 작성
3. **REFACTOR**: 중복 제거, 인터페이스 추출, 에러 처리 개선

### 1.2 테스트 전략

- **단위 테스트**: Config 구조체, 기본값, 검증 로직, 로드/세이브
- **통합 테스트**: 실제 파일 시스템 I/O (t.TempDir() 활용)
- **동시성 테스트**: `go test -race`로 데이터 레이스 감지
- **테이블 기반 테스트**: 다양한 입력/검증 시나리오를 table-driven test로 구성

---

## 2. 마일스톤 (우선순위 기반)

### Primary Goal: 설정 구조체 및 기본값

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 1.1 | Config 구조체 및 Default() | internal/config/config.go | 설정 구조체 정의, 기본값 함수 |
| Task 1.2 | Load/Save 구현 | internal/config/loader.go | JSON 파일 읽기/쓰기, 디렉토리 자동 생성 |
| Task 1.3 | Validate() 구현 | internal/config/validator.go | 범위 검증, 열거형 검증, 보정 로직 |
| Task 1.4 | 동시성 안전 보장 | internal/config/loader.go | sync.RWMutex 적용 |
| Task 1.5 | ConfigPath 환경변수 오버라이드 | internal/config/loader.go | 테스트 격리를 위해 설정 파일 경로를 환경변수로 오버라이드 가능하게 함 |

### Secondary Goal: 뷰어 통합

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 2.1 | 창 닫기 시 자동 저장 | internal/viewer/viewer.go | 창 크기/위치를 config에 저장. 주의: SPEC-UI-001의 viewer.go가 창 닫기 이벤트 콜백을 제공해야 함. 미제공 시 viewer.go 인터페이스 확장 Task 추가 필요 |
| Task 2.2 | 파일 열기 시 경로 저장 | cmd/winmdview/main.go | lastOpenedFile 업데이트 |
| Task 2.3 | 시작 시 설정 적용 | internal/viewer/viewer.go | 창 크기/위치/테마 적용 |

### Tertiary Goal: 내결함성

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 3.1 | 손상된 JSON 복구 | internal/config/loader.go | 파싱 실패 시 .bak 백업 + 기본값 복원 |
| Task 3.2 | 부분 설정 병합 | internal/config/loader.go | 누락된 필드만 기본값으로 채우기 |

### Optional Goal: 확장 준비

(해당 마일스톤의 항목 없음 - ConfigPath 오버라이드는 Primary Goal로 승격됨)

---

## 3. 파일 영향 분석 (File Impact)

### 3.1 신규 생성 파일

| 파일 | 목적 | 크기 예상 |
|------|------|-----------|
| internal/config/config.go | Config 구조체, Default() 함수 | ~60 줄 |
| internal/config/config_test.go | 구조체 및 기본값 테스트 | ~80 줄 |
| internal/config/loader.go | Load(), Save() 함수, 경로 결정 | ~120 줄 |
| internal/config/loader_test.go | 로드/세이브 통합 테스트 | ~150 줄 |
| internal/config/validator.go | Validate() 함수, 범위 검증 | ~80 줄 |
| internal/config/validator_test.go | 검증 로직 테이블 기반 테스트 | ~120 줄 |

### 3.2 수정 대상 파일

| 파일 | 변경 내용 |
|------|-----------|
| cmd/winmdview/main.go | config.Load() 호출, 설정 적용, lastOpenedFile 저장 |
| internal/viewer/viewer.go | 창 크기/위치 설정 적용, 닫기 시 저장 콜백 |

---

## 4. 의존성

### 4.1 선행 SPEC 의존성
- **SPEC-UI-001**: WebView2 뷰어가 동작해야 설정이 의미를 가짐
- 설정 모듈 자체는 독립적이므로, 뷰어와 병렬 개발 가능

### 4.2 외부 라이브러리 의존성
- 없음 (Go 표준 라이브러리만 사용)

### 4.3 후속 SPEC 영향
- **SPEC-THEME-001**: theme 설정값을 참조하여 CSS 테마 전환
- **SPEC-INSTALL-001**: 제거 시 %APPDATA% 설정 보존 정책 참조

---

## 5. 기술적 접근 방향

### 5.1 설정 파일 경로

`os.UserConfigDir()`를 사용하여 플랫폼 독립적인 설정 디렉토리를 가져온다.
Windows에서는 `%APPDATA%`를 반환하므로 최종 경로는:
`%APPDATA%\WinMarkdownViewer\config.json`

### 5.2 JSON 직렬화

`encoding/json`의 `MarshalIndent`를 사용하여 사람이 읽을 수 있는 형식으로 저장한다:
- 들여쓰기: 2 spaces
- 유니코드 이스케이프 없음 (경로에 한글 포함 가능)

### 5.3 설정 검증 패턴

Validate() 함수는 "보정" 패턴을 사용한다:
- 잘못된 값을 발견하면 에러를 반환하는 대신 기본값으로 교체
- 보정된 Config를 반환하여 호출자가 즉시 사용 가능
- 원본 파일은 보정된 값으로 자동 갱신하지 않음 (Save 호출은 별도)

### 5.4 동시성 설계

- 전역 Config 인스턴스는 sync.RWMutex로 보호
- Load(): RLock으로 파일 읽기 + JSON 파싱, 결과를 새 Config로 반환
- Save(): Lock으로 파일 쓰기
- Get/Set 개별 필드 접근은 내부적으로 뮤텍스 획득

---

## 6. 리스크 및 대응 방안

| 리스크 | 발생 확률 | 영향도 | 대응 방안 |
|--------|----------|--------|-----------|
| %APPDATA% 접근 불가 (권한 문제) | 낮음 | 높음 | 에러 로깅 + 메모리 내 기본 설정으로 동작 |
| 동시 프로세스 간 설정 파일 충돌 | 중 | 중 | 마지막 쓰기 우선(last-write-wins) 정책 |
| JSON 파일 수동 편집 시 구문 오류 | 중 | 낮음 | 파싱 실패 시 .bak 백업 + 기본값 복원 |
| 모니터 해상도 변경 시 창 위치 초과 | 낮음 | 낮음 | windowX/windowY의 -1 기본값으로 복원 |
| 설정 항목 추가 시 하위 호환성 | 중 | 중 | 누락 필드는 기본값으로 채움 (부분 병합) |
