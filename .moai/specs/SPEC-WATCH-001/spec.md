---
id: SPEC-WATCH-001
title: "Real-time Preview - File Watch and Auto Refresh"
version: 1.0.0
status: draft
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P1
lifecycle: spec-first
tags: [fsnotify, websocket, http-server, live-reload]
---

# SPEC-WATCH-001: 실시간 미리보기 - 파일 감시 및 자동 새로고침

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- Microsoft Edge WebView2 Runtime 설치 필수

### 1.2 기술 스택
- Go 1.26.0 (CGO 불필요)
- github.com/fsnotify/fsnotify - 크로스 플랫폼 파일 시스템 이벤트 감시
- github.com/gorilla/websocket - WebSocket 서버/클라이언트
  - **주의**: gorilla/websocket은 2022년 아카이브 후 재활성화됨. 대안으로 nhooyr.io/websocket 검토 가능. 구현 시작 전 최신 상태 확인 필수
- net/http (stdlib) - 내장 HTTP 서버
- github.com/yuin/goldmark - Markdown 렌더링 (SPEC-UI-001에서 구현 완료)
- github.com/jchv/go-webview2 - WebView2 바인딩 (SPEC-UI-001에서 구현 완료)

### 1.3 프로젝트 구조 변경
```
WinMarkdownViewer/
  cmd/winmdview/main.go              # 진입점 (수정: 서버 시작 로직 추가)
  internal/markdown/renderer.go      # 기존 유지
  internal/watcher/watcher.go        # [신규] 파일 변경 감시
  internal/watcher/watcher_test.go   # [신규] 감시 테스트
  internal/server/server.go          # [신규] HTTP 서버 + WebSocket
  internal/server/server_test.go     # [신규] 서버 테스트
  internal/server/handler.go         # [신규] HTTP/WebSocket 핸들러
  internal/viewer/viewer.go          # 수정: SetHtml -> Navigate 전환
  web/
    templates/viewer.html            # 수정: WebSocket 클라이언트 스크립트 추가
    css/github-markdown.css          # 기존 유지
    embed.go                         # 기존 유지
  go.mod                             # 수정: 의존성 추가
```

### 1.4 아키텍처 변경 (SetHtml -> Navigate 전환)

SPEC-UI-001에서는 `WebView2.SetHtml()`로 HTML을 직접 주입하는 방식이었으나, 이 SPEC에서는 내장 HTTP 서버를 통한 `WebView2.Navigate(localhost URL)` 방식으로 전환한다.

**전환 이유:**
- WebSocket 연결은 HTTP 프로토콜 위에서 동작하므로 실제 HTTP 서버가 필요
- `SetHtml()`은 `about:blank` origin을 사용하여 WebSocket 연결이 불가능
- `Navigate()`는 `http://localhost` origin을 제공하여 WebSocket 핸드셰이크 가능
- 향후 SPEC-WIN-001에서 단일 인스턴스 통신 시 HTTP 서버를 재활용 가능

**데이터 흐름:**
```
[파일 변경] -> fsnotify 이벤트
                  |
                  v (debounce)
            goldmark 재렌더링
                  |
                  v
        WebSocket 메시지 전송
                  |
                  v
   브라우저 JavaScript가 DOM 업데이트
        (스크롤 위치 유지)
```

---

## 2. Assumptions (가정)

- A1: SPEC-UI-001이 완료되어 기본 뷰어(goldmark 렌더링 + WebView2 표시)가 동작한다.
- A2: 감시 대상은 단일 .md 파일이며, 디렉토리 전체 감시는 범위 밖이다.
- A3: 내장 HTTP 서버는 localhost의 랜덤 포트에서 실행되어 외부 접근이 불가능하다.
- A4: 파일 변경 이벤트는 fsnotify의 Write 이벤트를 기준으로 감지한다.
- A5: Debounce 간격은 100ms로 설정하여 빠른 연속 저장 시 마지막 변경만 반영한다.
- A6: WebSocket 연결은 단일 클라이언트(WebView2)만 지원하면 충분하다.

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous (항상 활성)

- **REQ-U-001**: 시스템은 **항상** 내장 HTTP 서버를 localhost에서만 바인딩해야 한다 (127.0.0.1).
- **REQ-U-002**: 시스템은 **항상** WebSocket 연결 시 파일의 현재 내용을 즉시 전송해야 한다 (초기 로드).
- **REQ-U-003**: 시스템은 **항상** 파일 변경 감지 후 debounce를 적용하여 100ms 이내의 연속 변경은 마지막 변경만 반영해야 한다.
- **REQ-U-004**: 시스템은 **항상** DOM 업데이트 시 현재 스크롤 위치를 보존해야 한다 (최선 노력 방식). 콘텐츠 길이가 크게 변경된 경우 정확한 위치 복원을 보장하지 않음.

### 3.2 Event-Driven (이벤트 기반)

- **REQ-E-001**: **WHEN** 사용자가 .md 파일 경로를 인자로 프로그램을 실행하면, **THEN** 시스템은 내장 HTTP 서버를 랜덤 포트에서 시작하고, 해당 파일을 렌더링한 HTML을 서빙하며, WebView2를 `http://localhost:{port}` 로 Navigate해야 한다.
- **REQ-E-002**: **WHEN** 감시 중인 .md 파일이 외부 편집기에서 저장(Write 이벤트)되면, **THEN** 시스템은 파일을 다시 읽어 goldmark으로 재렌더링하고, WebSocket을 통해 새 HTML을 클라이언트에 전송해야 한다.
- **REQ-E-003**: **WHEN** WebSocket 클라이언트가 새 HTML 메시지를 수신하면, **THEN** JavaScript가 현재 스크롤 위치를 기록한 후 DOM을 업데이트하고, 기록한 스크롤 위치로 복원해야 한다.
- **REQ-E-004**: **WHEN** WebSocket 연결이 끊어지면, **THEN** 클라이언트 JavaScript는 자동으로 재연결을 시도해야 한다 (exponential backoff, 최대 30초).
- **REQ-E-005**: **WHEN** 사용자가 WebView2 윈도우를 닫으면, **THEN** 시스템은 파일 감시를 중지하고, WebSocket 연결을 정리하며, HTTP 서버를 종료하고, 프로세스를 종료해야 한다.

### 3.3 Unwanted (금지 행위)

- **REQ-N-001**: 시스템은 HTTP 서버를 0.0.0.0이나 외부 IP에 바인딩**하지 않아야 한다**.
- **REQ-N-002**: 시스템은 JavaScript location.reload() 또는 WebView2의 Navigate 재호출을 **하지 않아야 한다**. innerHTML 교체 방식으로 콘텐츠를 업데이트한다.
- **REQ-N-003**: 시스템은 debounce 없이 모든 파일 시스템 이벤트에 즉시 반응**하지 않아야 한다**.

### 3.4 State-Driven (상태 기반)

- **REQ-S-001**: **IF** 감시 중인 파일이 삭제되면, **THEN** 시스템은 마지막 렌더링 결과를 유지하고 파일 재생성 시 자동으로 감시를 재개해야 한다.
- **REQ-S-002**: **IF** 감시 중인 파일의 읽기 중 에러가 발생하면 (권한 변경 등), **THEN** 시스템은 WebSocket을 통해 에러 메시지를 클라이언트에 전송하고 마지막 성공 렌더링을 유지해야 한다.
- **REQ-S-003**: **IF** HTTP 서버의 포트 바인딩이 실패하면, **THEN** 시스템은 다른 랜덤 포트를 재시도하고, 3회 실패 시 에러 메시지를 출력하고 종료해야 한다.

### 3.5 Optional (선택 사항)

- **REQ-O-001**: **가능하면** WebSocket 메시지에 변경된 콘텐츠의 diff만 전송하여 네트워크 효율성을 높여야 한다 (MVP에서는 전체 HTML 전송 허용).
- **REQ-O-002**: **가능하면** 파일 변경 감지 시 윈도우 타이틀바에 업데이트 시각을 표시해야 한다.

---

## 4. Specifications (상세 명세)

### 4.1 파일 감시 모듈 (`internal/watcher/`)

- `NewWatcher(filePath string) (*Watcher, error)`: fsnotify 기반 파일 감시 인스턴스 생성
- `Watch(ctx context.Context) <-chan struct{}`: 파일 변경 이벤트를 debounce 처리하여 채널로 전달
- Debounce 로직: `time.AfterFunc` 또는 `time.Timer`를 사용하여 100ms 간격으로 통합
- fsnotify Write 이벤트만 감시 (Create, Remove는 파일 복원 감지용으로 처리)
- 파일 삭제 시 감시 유지, 재생성 시 자동 복구
- 파일 삭제 시 부모 디렉토리를 감시하여 파일 재생성을 감지하고 자동으로 감시를 재개한다. fsnotify 감시 해제 시 100ms 폴링으로 fallback한다
- `SwitchFile(newPath string) error`: 감시 대상 파일을 변경하는 메서드. 기존 감시를 중지하고 새 파일로 전환한다 (SPEC-WIN-001 단일 인스턴스 연동용)

### 4.2 HTTP 서버 + WebSocket (`internal/server/`)

- `NewServer(renderer *markdown.Renderer) (*Server, error)`: HTTP 서버 인스턴스 생성
- `Start() (int, error)`: 랜덤 포트에서 서버 시작, 할당된 포트 번호 반환
- `Shutdown(ctx context.Context) error`: graceful shutdown
- HTTP 엔드포인트:
  - `GET /` : 현재 렌더링된 HTML 페이지 서빙 (viewer.html 템플릿 + WebSocket 클라이언트 JS)
  - `GET /ws` : WebSocket 업그레이드 엔드포인트
- WebSocket 메시지 형식: JSON `{"type": "update", "html": "<렌더링된 HTML>"}`
- `Broadcast(html string)`: 연결된 모든 WebSocket 클라이언트에 HTML 전송

### 4.3 WebSocket 클라이언트 JavaScript (`web/templates/viewer.html` 수정)

- WebSocket 연결 (`ws://localhost:{port}/ws`)
- 메시지 수신 시:
  1. `window.scrollY` 기록
  2. `document.getElementById('content').innerHTML = msg.html` 업데이트
  3. `window.scrollTo(0, savedScrollY)` 복원
- 연결 끊김 시 자동 재연결 (1s, 2s, 4s, ... 최대 30s exponential backoff)

### 4.4 진입점 수정 (`cmd/winmdview/main.go`)

- 기존 `SetHtml()` 호출을 `Navigate(fmt.Sprintf("http://localhost:%d", port))` 로 변경
- 서버 시작 -> 감시 시작 -> Navigate -> 이벤트 루프 순서
- goroutine으로 파일 변경 이벤트 수신 후 재렌더링 + Broadcast
- `context.Context`를 활용한 graceful shutdown 체인

---

## 5. Constraints (제약사항)

- CGO 사용 금지 (pure Go 빌드 유지)
- HTTP 서버는 반드시 localhost(127.0.0.1)에만 바인딩
- 단일 파일 감시만 지원 (여러 파일 동시 감시는 범위 밖)
- WebSocket 메시지는 text 프레임 사용 (binary 불필요)
- 이 SPEC에서 컨텍스트 메뉴, 시스템 트레이, 단일 인스턴스, MSI 인스톨러는 범위 밖

---

## 6. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 시나리오 |
|-------------|-----------|-----------------|
| REQ-U-001 | internal/server/server.go | ACC-001 |
| REQ-U-002 | internal/server/handler.go | ACC-002 |
| REQ-U-003 | internal/watcher/watcher.go | ACC-003 |
| REQ-U-004 | web/templates/viewer.html (JS) | ACC-004 |
| REQ-E-001 | cmd/winmdview/main.go, internal/server/ | ACC-001 |
| REQ-E-002 | internal/watcher/, internal/server/ | ACC-002 |
| REQ-E-003 | web/templates/viewer.html (JS) | ACC-004 |
| REQ-E-004 | web/templates/viewer.html (JS) | ACC-005 |
| REQ-E-005 | cmd/winmdview/main.go | ACC-006 |
| REQ-N-001 | internal/server/server.go | ACC-007 |
| REQ-N-002 | web/templates/viewer.html (JS) | ACC-004 |
| REQ-N-003 | internal/watcher/watcher.go | ACC-003 |
| REQ-S-001 | internal/watcher/watcher.go | ACC-008 |
| REQ-S-002 | internal/server/handler.go | ACC-009 |
| REQ-S-003 | internal/server/server.go | ACC-010 |
| REQ-O-001 | internal/server/handler.go | - |
| REQ-O-002 | web/templates/viewer.html (JS) | - |
