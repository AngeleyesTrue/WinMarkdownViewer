---
id: SPEC-MULTIWIN-001
version: 0.2.0
status: completed
created: 2026-03-09
updated: 2026-03-09
author: MoAI
priority: high
related_specs:
  - SPEC-WIN-001
  - SPEC-WATCH-001
  - SPEC-CONFIG-001
  - SPEC-UI-001
---

## 변경 이력 (History)

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 0.1.0 | 2026-03-09 | 초기 SPEC 작성 |
| 0.2.0 | 2026-03-09 | 리뷰 반영: Phase 0 PoC 추가, 서버 아키텍처 근거 문서화, 빌드 스크립트 Makefile->build.ps1 변경, 윈도우 위치 전략 명확화, 범위 외 항목 명시, 최대 윈도우 제한/오류 처리 AC 추가 |

# SPEC-MULTIWIN-001: 멀티 윈도우 지원 및 빌드 개선

## 개요

WinMarkdownViewer의 두 가지 핵심 문제를 해결한다:
1. **콘솔 창 표시 문제**: `go build` 직접 실행 시 `-H windowsgui` 플래그 누락으로 콘솔 창이 나타남
2. **멀티 윈도우 지원**: 현재 단일 인스턴스/단일 윈도우 설계를 단일 프로세스/멀티 윈도우로 전환

## 환경 (Environment)

- Windows 10 21H2 이상, Windows 11
- Go 1.26+ 빌드 환경
- WebView2 Runtime (Edge 기반)
- go-webview2 라이브러리 (`github.com/jchv/go-webview2`)
- fsnotify 파일 감시 라이브러리
- gorilla/websocket WebSocket 라이브러리
- energye/systray 시스템 트레이 라이브러리

## 가정 (Assumptions)

- A1: go-webview2의 `Run()` 호출은 블로킹이며, 단일 goroutine에서만 호출 가능하다. 멀티 윈도우 구현 시 go-webview2가 여러 WebView 인스턴스를 동일 프로세스에서 지원하는지 사전 검증이 필요하다.
- A2: 각 윈도우는 독립적인 HTTP 서버 포트를 사용한다 (현재 단일 서버에서 멀티 윈도우용으로 확장).
- A3: Named Mutex 기반 단일 프로세스 제약은 유지하되, 새 파일 요청 시 기존 윈도우를 교체하지 않고 새 윈도우를 생성한다.
- A4: 메모리 사용량은 윈도우당 약 50MB를 기준으로 하며, 동시 윈도우 수에 비례하여 증가한다.
- A5: PowerShell 빌드 스크립트(`build.ps1`)를 통해 `-H windowsgui` 플래그가 표준 빌드에 항상 포함된다.

## 요구사항 (Requirements)

### R1: 콘솔 창 억제 (빌드 설정)

- **R1.1** [Ubiquitous]: 시스템은 **항상** 릴리스 빌드 시 `-ldflags="-H windowsgui"` 플래그를 포함하여 콘솔 창이 표시되지 않아야 한다.
- **R1.2** [Event-Driven]: **WHEN** 개발자가 `.\build.ps1 build` 명령을 실행하면 **THEN** `-H windowsgui` 플래그가 자동으로 적용된 실행 파일이 생성되어야 한다.
- **R1.3** [Optional]: **가능하면** `.\build.ps1 dev` 명령으로 콘솔 창이 표시되는 디버그 빌드를 별도로 제공한다.

### R2: 멀티 윈도우 생성

- **R2.1** [Event-Driven]: **WHEN** 사용자가 새로운 .md 파일을 열면 (파일 연결, 컨텍스트 메뉴) **THEN** 기존 윈도우를 유지한 채 새 윈도우가 생성되어야 한다.
  - 참고: 드래그앤드롭은 v1.0 범위 외(OUT OF SCOPE)이다.
- **R2.2** [Event-Driven]: **WHEN** Named Pipe를 통해 파일 경로가 전달되면 **THEN** 기존 프로세스가 해당 파일을 새 윈도우에서 열어야 한다.
- **R2.3** [Unwanted]: 시스템은 새 파일 요청 시 기존 윈도우의 콘텐츠를 **교체하지 않아야 한다**.
- **R2.4** [State-Driven]: **IF** 동일한 파일이 이미 열려 있으면 **THEN** 해당 윈도우를 포그라운드로 활성화해야 한다.
- **R2.5** [Ubiquitous]: 각 윈도우는 **항상** 하나의 파일만 표시해야 한다. 윈도우 내에서 다른 파일로 전환하는 기능은 제공하지 않는다.
- **R2.6** [State-Driven]: **IF** 열린 윈도우가 10개에 도달하면 **THEN** 사용자에게 경고 대화상자를 표시해야 하며, 10개를 초과하는 윈도우 생성 요청은 거부해야 한다.
- **R2.7** [Event-Driven]: **WHEN** 파일 열기에 실패하면 (파일 미존재, 읽기 권한 없음 등) **THEN** 오류 대화상자를 표시하고 빈 윈도우를 생성하지 않아야 한다.

### R3: 윈도우 생명주기 관리

- **R3.1** [Event-Driven]: **WHEN** 사용자가 특정 윈도우를 닫으면 **THEN** 해당 윈도우와 연관된 리소스(HTTP 서버, 파일 감시자, WebSocket 연결)만 정리되어야 한다.
- **R3.2** [Unwanted]: 하나의 윈도우를 닫는 것이 다른 열린 윈도우에 영향을 **미치지 않아야 한다**.
- **R3.3** [Event-Driven]: **WHEN** 마지막 윈도우가 닫히면 **THEN** 시스템 트레이 아이콘만 유지하고 프로세스는 계속 실행되어야 한다.
- **R3.4** [Event-Driven]: **WHEN** 시스템 트레이 메뉴에서 "종료"를 선택하면 **THEN** 모든 윈도우를 닫고 프로세스를 종료해야 한다.

### R4: 윈도우별 파일 감시

- **R4.1** [Ubiquitous]: 시스템은 **항상** 각 윈도우에 독립적인 파일 감시자(fsnotify Watcher)를 할당해야 한다.
- **R4.2** [Event-Driven]: **WHEN** 감시 중인 파일이 변경되면 **THEN** 해당 파일을 표시하는 윈도우만 업데이트되어야 한다.
- **R4.3** [Unwanted]: 파일 변경 이벤트가 관련 없는 윈도우에 전파되지 **않아야 한다**.

### R5: 시스템 트레이 통합

- **R5.1** [Ubiquitous]: 시스템 트레이 메뉴는 **항상** 현재 열려 있는 모든 윈도우의 목록을 표시해야 한다.
- **R5.2** [Event-Driven]: **WHEN** 시스템 트레이 메뉴에서 특정 윈도우 항목을 클릭하면 **THEN** 해당 윈도우가 포그라운드로 활성화되어야 한다.
- **R5.3** [Event-Driven]: **WHEN** 새 윈도우가 생성되거나 닫히면 **THEN** 시스템 트레이 메뉴가 자동으로 업데이트되어야 한다.

### R6: Named Pipe 프로토콜 확장

- **R6.1** [Event-Driven]: **WHEN** 두 번째 인스턴스가 실행되면 **THEN** Named Pipe를 통해 `OPEN:<filepath>` 형식의 명령을 기존 프로세스에 전송해야 한다.
- **R6.2** [State-Driven]: **IF** Named Pipe 프로토콜에서 `OPEN:` 접두사가 없는 메시지를 수신하면 **THEN** 하위 호환성을 위해 기존 방식(파일 경로 직접 전달)으로 처리해야 한다.

## 사양 (Specifications)

### S1: WindowManager 구조체

```
internal/window/
  manager.go      # WindowManager: 윈도우 생성, 추적, 정리
  window.go       # Window: 개별 윈도우 상태 (viewer, server, watcher)
  manager_test.go
  window_test.go
```

- `WindowManager`는 모든 윈도우 인스턴스를 추적하는 중앙 관리자
- 각 `Window`는 자체 `viewer.Viewer`, `server.Server`, `watcher.Watcher`를 소유
- 윈도우 ID는 순차 정수 (1, 2, 3...)로 할당

### S2: HTTP 서버 라우팅 변경

- 현재: 단일 `Server`가 하나의 콘텐츠를 모든 클라이언트에 브로드캐스트
- 변경: 각 윈도우가 자체 `Server` 인스턴스를 소유하거나, 단일 서버에서 경로 기반 라우팅 (`/window/{id}/ws`) 사용
- **권장: 윈도우별 독립 서버 (포트 분리)**

#### 서버 아키텍처 결정 근거

**윈도우별 독립 서버를 권장하는 이유:**

| 비교 항목 | 윈도우별 독립 서버 (권장) | 단일 서버 + 경로 라우팅 |
|-----------|-------------------------|----------------------|
| 구현 복잡도 | 낮음 (기존 Server 코드 재사용) | 높음 (라우팅 로직, 세션 관리 추가) |
| 격리성 | 높음 (윈도우 독립, 장애 전파 없음) | 낮음 (서버 장애 시 전체 영향) |
| 리소스 정리 | 단순 (윈도우별 Server.Shutdown()) | 복잡 (경로별 핸들러 해제 필요) |
| 포트 사용 | 윈도우당 1포트 (port 0 자동 할당) | 단일 포트 |
| 보안 영향 | 무시 가능 | 무시 가능 |

**보안 고려사항:**
- 모든 서버는 `127.0.0.1`에 바인딩하므로 외부 네트워크 접근 불가
- 외부 방화벽은 loopback 트래픽에 관여하지 않음
- 로컬 보안 소프트웨어(Windows Defender, 백신)가 이론적으로 간섭 가능하나, 실사용에서 문제 발생 확률은 극히 낮음
- 포트 0 자동 할당으로 포트 충돌 위험 최소화

**결론:** 기존 `server.Server` 코드를 변경 없이 재사용할 수 있고, 윈도우 간 완전한 격리를 제공하므로 독립 서버 방식을 채택한다.

### S3: PowerShell 빌드 스크립트 추가

프로젝트 루트에 `build.ps1` PowerShell 스크립트 생성 (기존 `installer/build-msi.ps1` 컨벤션과 일치):
- `.\build.ps1 build`: 릴리스 빌드 (`-H windowsgui` 포함)
- `.\build.ps1 dev`: 개발 빌드 (콘솔 창 표시, 디버그 로깅)
- `.\build.ps1 test`: 테스트 실행
- `.\build.ps1 clean`: 빌드 산출물 정리

**PowerShell을 선택한 이유:**
- Windows 전용 프로젝트이므로 Windows 기본 도구 활용
- `make`는 Windows에 기본 설치되어 있지 않음
- 기존 `installer/build-msi.ps1`과 일관된 스크립트 형식

**참고:** WSL/Git Bash 사용자를 위해 Makefile을 보조 옵션으로 제공할 수 있으나, 필수는 아님.

### S4: 설정 확장

`config.Config` 구조체에 멀티 윈도우 관련 설정 추가:
- 기존 단일 윈도우 설정(`WindowWidth`, `WindowHeight`, `WindowX`, `WindowY`)은 기본값으로 유지
- 파일별 윈도우 위치/크기 저장은 v1.0에서 구현하지 않음 (복잡도 과다)

**윈도우 위치/크기 전략:**
- 첫 번째 윈도우: 설정 파일에 저장된 위치/크기 사용
- 이후 윈도우: 마지막 윈도우 위치에서 캐스케이드 오프셋 적용 (우측 30px, 하단 30px)
- 애플리케이션 재시작 시: 첫 번째 윈도우는 저장된 설정 사용, 이후 윈도우는 캐스케이드
- 화면 경계를 벗어나면 (0, 0) 위치로 리셋

## 추적성 (Traceability)

| 요구사항 | 관련 파일 | 관련 SPEC |
|---------|----------|----------|
| R1 | build.ps1, installer/build-msi.ps1 | SPEC-INSTALL-001 |
| R2 | internal/window/manager.go, internal/app/pipe.go | SPEC-WIN-001 |
| R3 | internal/window/manager.go, internal/tray/tray.go | SPEC-WIN-001 |
| R4 | internal/window/window.go, internal/watcher/watcher.go | SPEC-WATCH-001 |
| R5 | internal/tray/tray.go | SPEC-WIN-001 |
| R6 | internal/app/pipe.go, internal/app/constants.go | SPEC-WIN-001 |

## 구현 노트 (Implementation Notes)

- **PoC 결과**: Path A (goroutine-per-window) 성공 확인 (`cmd/poc/multiwin/` 검증)
- **윈도우 위치**: 캐스케이드 방식에서 OS 기본 배치(`CW_USEDEFAULT`)로 최종 변경 (S4 전략 수정)
- **build.ps1 출력 경로**: `dist/` 디렉토리로 변경
- **테스트 커버리지**: `internal/window` 88.2%, `internal/app` 87.3%
