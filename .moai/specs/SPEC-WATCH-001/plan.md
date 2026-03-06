---
spec_id: SPEC-WATCH-001
type: implementation-plan
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-WATCH-001 구현 계획

## 1. 구현 전략

### 1.1 개발 방법론

TDD (Test-Driven Development) - RED-GREEN-REFACTOR 사이클에 따라 구현합니다.

### 1.2 전체 접근 방식

바텀업(Bottom-Up) 방식으로 독립 모듈부터 구현 후 통합합니다:

1. 파일 감시 모듈 (internal/watcher/) - fsnotify + debounce
2. HTTP 서버 + WebSocket 모듈 (internal/server/) - 내장 서버
3. HTML 템플릿 수정 - WebSocket 클라이언트 JavaScript
4. 뷰어 전환 - SetHtml -> Navigate 전환
5. 진입점 통합 - 전체 파이프라인 연결

### 1.3 선행 조건

- SPEC-UI-001 구현 완료 (기본 뷰어 MVP)
  - `internal/markdown/renderer.go` - goldmark 렌더링 엔진
  - `internal/viewer/viewer.go` - WebView2 윈도우 관리
  - `web/` - HTML 템플릿 및 CSS
  - `cmd/winmdview/main.go` - CLI 진입점

---

## 2. 마일스톤

### Primary Goal: 파일 감시 모듈

**Task 1: fsnotify 의존성 추가**
- 파일: `go.mod`
- 작업: `go get github.com/fsnotify/fsnotify`
- 의존성: 없음

**Task 2: Watcher 테스트 작성 (RED)**
- 파일: `internal/watcher/watcher_test.go`
- 작업:
  - 파일 생성/수정 시 이벤트 수신 테스트
  - Debounce 동작 테스트 (100ms 이내 연속 변경 시 단일 이벤트)
  - 파일 삭제 후 재생성 시 감시 복구 테스트
  - context 취소 시 정상 종료 테스트
- 의존성: Task 1

**Task 3: Watcher 구현 (GREEN)**
- 파일: `internal/watcher/watcher.go`
- 작업:
  - `NewWatcher(filePath string) (*Watcher, error)` 생성자
  - `Watch(ctx context.Context) <-chan struct{}` 감시 시작 및 이벤트 채널 반환
  - `Close() error` 리소스 정리
  - fsnotify Write 이벤트 감지 + debounce 로직 (time.Timer 기반)
  - 파일 삭제 시 감시 유지, 재생성 시 재등록
- 의존성: Task 2

**Task 4: Watcher 리팩토링 (REFACTOR)**
- 파일: `internal/watcher/watcher.go`
- 작업: debounce 로직 분리, 에러 처리 개선, 로깅 추가
- 의존성: Task 3

### Secondary Goal: HTTP 서버 + WebSocket 모듈

**Task 5: gorilla/websocket 의존성 추가**
- 파일: `go.mod`
- 작업: `go get github.com/gorilla/websocket`
- 의존성: 없음

**Task 6: Server 테스트 작성 (RED)**
- 파일: `internal/server/server_test.go`
- 작업:
  - HTTP 서버 시작/종료 테스트
  - localhost 바인딩 검증 테스트
  - `GET /` 엔드포인트 HTML 응답 테스트
  - WebSocket 핸드셰이크 및 메시지 수신 테스트
  - Broadcast 시 연결된 모든 클라이언트 수신 테스트
  - 연결 초기화 시 현재 HTML 즉시 전송 테스트
- 의존성: Task 5

**Task 7: Server 구현 (GREEN)**
- 파일: `internal/server/server.go`, `internal/server/handler.go`
- 작업:
  - `NewServer(opts ...Option) (*Server, error)` 생성자
  - `Start() (int, error)` 랜덤 포트 서버 시작
  - `Shutdown(ctx context.Context) error` graceful shutdown
  - `SetContent(html string)` 현재 콘텐츠 설정
  - `Broadcast(html string)` 전체 WebSocket 클라이언트에 전송
  - HTTP 핸들러: `GET /` (HTML 페이지), `GET /ws` (WebSocket 업그레이드)
  - WebSocket 클라이언트 관리 (등록, 해제, 브로드캐스트)
  - gorilla/websocket Upgrader 설정 (localhost origin 검증)
- 의존성: Task 6

**Task 8: Server 리팩토링 (REFACTOR)**
- 파일: `internal/server/server.go`, `internal/server/handler.go`
- 작업: 핸들러 분리, 에러 처리 개선, 연결 관리 동시성 안전 확보
- 의존성: Task 7

### Final Goal: 통합 및 UI 전환

**Task 9: HTML 템플릿 수정**
- 파일: `web/templates/viewer.html`
- 작업:
  - WebSocket 클라이언트 JavaScript 추가
  - 메시지 수신 시 DOM 업데이트 로직
  - 스크롤 위치 저장/복원 로직
  - 자동 재연결 로직 (exponential backoff)
  - WebSocket URL 동적 구성 (`ws://${location.host}/ws`)
- 의존성: Task 7

**Task 10: Viewer 모듈 수정**
- 파일: `internal/viewer/viewer.go`
- 작업:
  - `Navigate(url string) error` 메서드 추가 (또는 기존 메서드 수정)
  - `SetHtml()` 대신 `Navigate()` 사용하도록 인터페이스 변경
- 의존성: Task 9

**Task 10.5: SPEC-UI-001 기존 테스트 업데이트**
- 파일: `internal/viewer/viewer_test.go` 등 기존 테스트 파일
- 작업:
  - viewer.go 인터페이스 추상화 (SetHtml/Navigate 모두 지원)
  - 기존 SetHtml 기반 테스트가 인터페이스 변경 후에도 통과하도록 수정
  - Navigate 방식 추가에 따른 테스트 케이스 보완
- 의존성: Task 10

**Task 11: 진입점 통합 테스트 작성 (RED)**
- 파일: `cmd/winmdview/main_test.go`
- 작업:
  - 서버 시작 -> Navigate 흐름 테스트 (통합)
  - Go 레벨에서 watcher -> server.Broadcast() 흐름만 자동 테스트. WebSocket -> DOM 업데이트는 수동 검증으로 분류
  - 종료 시 전체 리소스 정리 테스트
- 의존성: Task 10.5

**Task 12: 진입점 통합 구현 (GREEN)**
- 파일: `cmd/winmdview/main.go`
- 작업:
  - HTTP 서버 생성 및 시작
  - Watcher 생성 및 파일 감시 시작
  - `Navigate(fmt.Sprintf("http://localhost:%d", port))` 호출
  - goroutine: 파일 변경 이벤트 -> 재렌더링 -> Broadcast
  - context.Context 기반 graceful shutdown 체인
  - 윈도우 종료 -> ctx cancel -> watcher/server 정리
- 의존성: Task 11

**Task 13: 빌드 및 통합 검증**
- 작업:
  - `go build ./cmd/winmdview/` 성공 확인
  - `go test -race ./...` 경쟁 조건 검사
  - 실제 .md 파일로 실시간 미리보기 동작 검증
- 의존성: Task 12

---

## 3. 파일 영향 분석

| 파일 | 작업 유형 | 복잡도 | 관련 Task |
|------|-----------|--------|-----------|
| `go.mod` | 수정 | 낮음 | Task 1, 5 |
| `internal/watcher/watcher.go` | 신규 생성 | 중간 | Task 3, 4 |
| `internal/watcher/watcher_test.go` | 신규 생성 | 중간 | Task 2 |
| `internal/server/server.go` | 신규 생성 | 높음 | Task 7, 8 |
| `internal/server/handler.go` | 신규 생성 | 높음 | Task 7, 8 |
| `internal/server/server_test.go` | 신규 생성 | 높음 | Task 6 |
| `web/templates/viewer.html` | 수정 | 중간 | Task 9 |
| `internal/viewer/viewer.go` | 수정 | 낮음 | Task 10 |
| `cmd/winmdview/main.go` | 수정 | 중간 | Task 12 |
| `cmd/winmdview/main_test.go` | 수정 | 중간 | Task 11 |

**총 파일 수**: 10개 (5개 신규 생성 + 5개 수정)
**전체 복잡도**: 높음

---

## 4. 기술적 접근

### 4.1 실시간 미리보기 아키텍처

```
[외부 편집기] ─ 파일 저장 ─> [.md 파일]
                                  |
                        fsnotify Write 이벤트
                                  |
                                  v
                     [Watcher] ─ debounce(100ms) ─> 이벤트 발행
                                                        |
                                                        v
                                              os.ReadFile() + goldmark
                                                        |
                                                        v
                                                 렌더링된 HTML
                                                        |
                                                        v
                                              Server.Broadcast(html)
                                                        |
                                              WebSocket 메시지 전송
                                                        |
                                                        v
                                        [WebView2 / viewer.html]
                                          1. scrollY 저장
                                          2. innerHTML 업데이트
                                          3. scrollTo 복원
```

### 4.2 Debounce 전략

```
이벤트 시간축:
  t=0ms    Write 이벤트 → Timer 시작 (100ms)
  t=50ms   Write 이벤트 → Timer 리셋 (100ms)
  t=80ms   Write 이벤트 → Timer 리셋 (100ms)
  t=180ms  Timer 만료 → 채널에 이벤트 발행 (1회만 처리)
```

### 4.3 WebSocket 연결 관리

- `sync.RWMutex`로 클라이언트 맵 보호
- 클라이언트 등록/해제를 별도 goroutine(hub 패턴)으로 관리
- 각 클라이언트 연결에 대해 읽기 goroutine 유지 (연결 상태 감지)

### 4.4 Graceful Shutdown 체인

```
윈도우 닫기 이벤트
       |
       v
  ctx.Cancel()
       |
       ├── Watcher.Close()     → fsnotify 감시 종료
       ├── Server.Shutdown()   → HTTP 서버 graceful shutdown
       │      └── WebSocket 연결 정리
       └── Viewer 리소스 해제
```

---

## 5. 리스크 및 대응

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| fsnotify가 일부 Windows 환경에서 이벤트 중복 발생 | 중간 | Debounce로 해결, Write 이벤트만 필터링 |
| WebSocket 연결이 WebView2 내부에서 불안정할 수 있음 | 중간 | 자동 재연결 로직으로 복구, exponential backoff |
| gorilla/websocket의 goroutine 경쟁 조건 | 높음 | `go test -race` 필수, sync.RWMutex 사용 |
| 대용량 파일 재렌더링 시 지연 | 낮음 | MVP에서는 동기 렌더링, 추후 비동기 처리 검토 |
| Navigate 전환 시 기존 SetHtml 테스트 실패 가능 | 중간 | Viewer 인터페이스를 추상화하여 두 방식 모두 지원 |

---

## 6. SPEC-UI-001 의존성

이 SPEC은 SPEC-UI-001의 다음 구현에 의존합니다:

| SPEC-UI-001 구현물 | 이 SPEC에서의 활용 |
|--------------------|-------------------|
| `internal/markdown/renderer.go` | 파일 변경 시 재렌더링에 사용 |
| `internal/viewer/viewer.go` | Navigate() 메서드 추가를 위해 수정 |
| `web/templates/viewer.html` | WebSocket JS 추가를 위해 수정 |
| `web/css/github-markdown.css` | 기존 CSS 그대로 유지 |
| `web/embed.go` | 기존 embed 구조 유지 |

---

## 7. 범위 외 (Out of Scope)

| 기능 | 예정 SPEC |
|------|-----------|
| 우클릭 컨텍스트 메뉴 등록 | SPEC-WIN-001 |
| 시스템 트레이 아이콘 | SPEC-WIN-001 |
| 단일 인스턴스 관리 | SPEC-WIN-001 |
| 여러 파일 동시 감시 | 미정 |
| diff 기반 부분 업데이트 | 미정 (Optional 요구사항) |

---

## 8. 다음 단계

1. `/moai:2-run SPEC-WATCH-001` 실행하여 TDD 사이클로 구현 시작
2. 구현 완료 후 `/moai:3-sync SPEC-WATCH-001` 실행하여 문서화
