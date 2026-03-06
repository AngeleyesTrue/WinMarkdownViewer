---
spec_id: SPEC-WATCH-001
type: acceptance-criteria
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-WATCH-001 수용 기준

## 1. 핵심 시나리오

### ACC-001: 내장 HTTP 서버를 통한 Markdown 렌더링

```gherkin
Given SPEC-UI-001이 구현 완료되어 기본 뷰어가 동작한다
  And 유효한 UTF-8 Markdown 파일 "test.md"가 존재한다
When 사용자가 "winmdview.exe test.md" 명령을 실행한다
Then 내장 HTTP 서버가 localhost의 랜덤 포트에서 시작된다
  And WebView2가 "http://localhost:{port}" 로 Navigate된다
  And Markdown이 GitHub 스타일로 렌더링되어 표시된다
  And 서버가 0.0.0.0이 아닌 127.0.0.1에만 바인딩되어 있다
```

### ACC-002: 파일 변경 시 자동 업데이트

> **참고**: 이 시나리오는 ACC-007(WebSocket 초기 연결 시 현재 내용 전송)과 연관됩니다. ACC-002는 파일 변경 후 업데이트 흐름을, ACC-007은 최초 연결 시 콘텐츠 전송을 검증합니다.

```gherkin
Given "test.md" 파일이 뷰어에서 열려 있다
  And WebSocket 연결이 수립되어 있다
When 외부 편집기(VS Code 등)에서 "test.md"의 내용을 수정하고 저장한다
Then 1초 이내에 뷰어의 내용이 변경된 Markdown으로 업데이트된다
  And goldmark으로 재렌더링된 HTML이 WebSocket을 통해 전송된다
  And DOM이 부분 업데이트된다 (전체 페이지 새로고침이 아님)
```

### ACC-003: Debounce 동작

```gherkin
Given "test.md" 파일이 뷰어에서 열려 있다
When 50ms 간격으로 3번 연속 파일을 저장한다
Then 렌더링은 마지막 저장 후 100ms 경과 시점에 1회만 수행된다
  And WebSocket 메시지도 1회만 전송된다
  And 최종 파일 내용이 반영된다
```

### ACC-004: 스크롤 위치 보존

```gherkin
Given "long-document.md" 파일이 뷰어에서 열려 있다
  And 문서가 스크롤 가능한 길이이다
  And 사용자가 페이지 중간까지 스크롤한 상태이다
When 파일 내용이 수정되어 자동 업데이트가 발생한다
Then 업데이트 후 스크롤 위치가 업데이트 전과 동일하게 유지된다
  And 사용자가 직접 스크롤하지 않아도 같은 위치를 보고 있다
```

---

## 2. WebSocket 시나리오

### ACC-005: WebSocket 자동 재연결

```gherkin
Given WebSocket 연결이 수립되어 있다
When 네트워크 일시 장애 등으로 WebSocket 연결이 끊어진다
Then 클라이언트 JavaScript가 1초 후 재연결을 시도한다
  And 재연결 실패 시 2초, 4초, 8초로 간격을 늘리며 재시도한다
  And 최대 재연결 간격은 30초이다
  And 재연결 성공 시 현재 파일 내용이 즉시 수신된다
```

**테스트 방법**: 서버 측에서 WebSocket 연결을 강제 종료하여 재연결을 유발한다.

### ACC-006: 윈도우 종료 시 전체 리소스 정리

```gherkin
Given 뷰어가 실행 중이다
  And HTTP 서버, WebSocket 연결, 파일 감시가 모두 활성 상태이다
When 사용자가 WebView2 윈도우를 닫는다
Then 파일 감시(fsnotify)가 중지된다
  And WebSocket 연결이 정상적으로 종료된다
  And HTTP 서버가 graceful shutdown된다
  And 프로세스가 exit code 0으로 종료된다
  And goroutine 누수가 발생하지 않는다
```

### ACC-007: WebSocket 초기 연결 시 현재 내용 전송

> **참고**: 이 시나리오는 ACC-002(파일 변경 시 자동 업데이트)와 연관됩니다. ACC-007은 최초 연결 시 콘텐츠 전송을, ACC-002는 이후 파일 변경에 의한 업데이트를 검증합니다.

```gherkin
Given 뷰어가 실행 중이고 "test.md"가 렌더링되어 있다
When WebSocket 클라이언트가 처음 연결된다
Then 서버가 현재 렌더링된 HTML을 즉시 전송한다
  And 클라이언트가 DOM을 해당 내용으로 업데이트한다
```

---

## 3. 에러 처리 시나리오

### ACC-008: 감시 중 파일 삭제

```gherkin
Given "test.md" 파일이 뷰어에서 열려 있고 감시 중이다
When 파일 시스템에서 "test.md"가 삭제된다
Then 뷰어는 마지막으로 렌더링된 내용을 계속 표시한다
  And 프로그램이 crash하지 않는다
When 같은 경로에 "test.md"가 다시 생성된다
Then 파일 감시가 자동으로 재개된다
  And 새 파일 내용이 렌더링되어 표시된다
```

### ACC-009: 파일 읽기 에러 시 마지막 성공 렌더링 유지

```gherkin
Given "test.md" 파일이 뷰어에서 열려 있다
When 파일의 읽기 권한이 변경되어 읽기가 실패한다
Then 마지막으로 성공한 렌더링 결과를 유지한다
  And WebSocket을 통해 에러 메시지가 전송된다
  And 프로그램이 crash하지 않는다
When 파일의 읽기 권한이 복구된다
Then 다음 파일 변경 감지 시 정상적으로 재렌더링된다
```

### ACC-010: 포트 바인딩 실패 시 재시도

```gherkin
Given HTTP 서버가 포트를 할당받으려 한다
When 선택한 랜덤 포트가 이미 사용 중이다
Then 다른 랜덤 포트로 재시도한다
  And 최대 3회까지 재시도한다
When 3회 모두 실패한다
Then 에러 메시지를 출력하고 프로세스가 종료된다
  And exit code 1이 반환된다
```

---

## 4. 엣지 케이스

### ACC-011: 매우 빈번한 파일 저장

```gherkin
Given "test.md" 파일이 뷰어에서 열려 있다
When 10ms 간격으로 10번 연속 파일을 저장한다
Then debounce에 의해 렌더링은 최대 2회 이내로 수행된다
  And 프로그램이 안정적으로 동작한다
  And goroutine 누수가 발생하지 않는다
```

### ACC-012: 대용량 파일 변경 시 렌더링

```gherkin
Given 5MB 크기의 "large.md" 파일이 뷰어에서 열려 있다
When 파일 내용이 수정되고 저장된다
Then 재렌더링이 완료된다
  And WebSocket을 통해 새 HTML이 전송된다
  And 프로그램이 메모리 부족으로 crash하지 않는다
```

### ACC-013: 빈 파일로 변경

```gherkin
Given "test.md" 파일에 내용이 있고 뷰어에 표시 중이다
When 파일 내용을 모두 삭제하고 빈 파일로 저장한다
Then 뷰어가 빈 HTML 페이지를 표시한다 (또는 "내용이 없습니다" 안내)
  And 프로그램이 crash하지 않는다
```

---

## 5. 빌드 및 실행 검증

### ACC-014: 빌드 성공

```gherkin
Given Go 1.26.0이 설치되어 있다
  And fsnotify, gorilla/websocket 의존성이 추가되어 있다
When "go build ./cmd/winmdview/" 명령을 실행한다
Then 빌드가 에러 없이 성공한다
  And CGO가 사용되지 않는다 (CGO_ENABLED=0)
```

### ACC-015: 테스트 통과

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test ./..." 명령을 실행한다
Then 모든 테스트가 통과한다
  And Watcher debounce 테스트가 포함된다
  And HTTP 서버 시작/종료 테스트가 포함된다
  And WebSocket 메시지 송수신 테스트가 포함된다
  And 테스트 커버리지가 85% 이상이다
```

### ACC-016: 경쟁 조건 검사

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test -race ./..." 명령을 실행한다
Then 데이터 경쟁(race condition)이 감지되지 않는다
  And WebSocket 브로드캐스트와 클라이언트 등록/해제 간 경쟁이 없다
  And 파일 감시와 렌더링 간 경쟁이 없다
```

---

## 6. Quality Gate (품질 게이트)

| 항목 | 기준 | 검증 방법 |
|------|------|-----------|
| 빌드 성공 | `go build` 에러 없음 | `go build ./cmd/winmdview/` |
| 테스트 통과 | 모든 테스트 PASS | `go test ./...` |
| 테스트 커버리지 | 85% 이상 (`internal/watcher/`, `internal/server/` 기준) | `go test -coverprofile=coverage.out ./internal/watcher/... ./internal/server/...` |
| 경쟁 조건 | 감지 없음 | `go test -race ./...` |
| 정적 분석 | 경고 없음 | `go vet ./...` |
| CGO 미사용 | CGO_ENABLED=0 빌드 성공 | `CGO_ENABLED=0 go build ./cmd/winmdview/` |
| 코드 품질 | lint 통과 | `golangci-lint run` (설치 시) |
| 응답 시간 | 파일 변경 후 1초 이내 업데이트 | 수동 검증 |

---

## 7. Definition of Done (완료 정의)

- [ ] 모든 요구사항(REQ-*)에 대응하는 구현 코드 존재
- [ ] 모든 수용 기준(ACC-*)의 Given-When-Then 시나리오 검증 완료
- [ ] `go build ./cmd/winmdview/` 성공 (CGO_ENABLED=0)
- [ ] `go test ./...` 모든 테스트 통과
- [ ] `go test -race ./...` 경쟁 조건 미감지
- [ ] `go vet ./...` 경고 없음
- [ ] 테스트 커버리지 85% 이상 (`internal/watcher/`, `internal/server/` 기준)
- [ ] 실제 Markdown 파일로 실시간 미리보기 동작 수동 검증
- [ ] VS Code에서 파일 저장 시 1초 이내 자동 업데이트 확인
- [ ] 스크롤 위치 보존 동작 확인
- [ ] 윈도우 종료 시 모든 리소스 정리 확인
