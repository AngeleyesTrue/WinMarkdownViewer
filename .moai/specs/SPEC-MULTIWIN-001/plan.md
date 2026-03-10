---
id: SPEC-MULTIWIN-001
type: plan
version: 0.2.0
---

# SPEC-MULTIWIN-001: 구현 계획

## 마일스톤 개요

### Phase 0 (PoC): go-webview2 멀티 인스턴스 검증 (MUST)

**목적**: 멀티 윈도우 구현 전, go-webview2가 동일 프로세스에서 여러 WebView2 인스턴스를 지원하는지 검증한다.

**영향 범위**: `cmd/poc/` (임시 디렉토리, 검증 후 제거 가능)

**검증 항목**:
1. 두 개의 WebView2 인스턴스를 별도 goroutine에서 `Run()` 호출 가능한지 확인
2. COM STA(Single-Threaded Apartment) 스레딩 모델 제약 확인
3. `runtime.LockOSThread()`와 COM 초기화 조합 테스트
4. 대안 라이브러리 `webview/webview` 조사 및 비교

**결과 경로 (3가지 시나리오)**:

| 경로 | 조건 | 후속 조치 |
|------|------|----------|
| **Path A** (goroutine-per-window) | 별도 goroutine에서 `Run()` 호출이 정상 동작 | 현재 설계대로 진행. 각 윈도우가 독립 goroutine에서 실행 |
| **Path B** (single-thread multi-window) | goroutine 방식 불가, COM API 직접 사용으로 단일 스레드에서 여러 윈도우 관리 가능 | viewer 계층 재설계. 단일 메시지 루프에서 모든 윈도우를 관리하도록 변경 |
| **Path C** (multi-process) | 동일 프로세스에서 멀티 인스턴스 불가능 | 멀티 프로세스 모델로 전환. 각 윈도우가 별도 프로세스로 실행되고, Named Pipe로 조율 |

**PoC 코드 구조**:
```
cmd/poc/multiwin/
  main.go    # 2개 WebView2 인스턴스를 별도 goroutine에서 실행하는 최소 테스트
```

**완료 기준**: Path A/B/C 중 하나의 결과가 확정되고, 해당 경로에 따라 Secondary Goal의 기술 접근을 확정한다.

---

### Primary Goal: 콘솔 창 억제 (빌드 설정 개선)

**영향 범위**: 프로젝트 루트 (신규 파일)

| 작업 | 파일 | 설명 |
|------|------|------|
| build.ps1 생성 | `build.ps1` (신규) | `build`, `dev`, `test`, `clean` 타겟 정의 |
| 빌드 문서 업데이트 | `README.md` | build.ps1 사용법 추가 |

**기술 접근**:
- `.\build.ps1 build`: `go build -ldflags="-s -w -H windowsgui" -o winmdview.exe ./cmd/winmdview`
- `.\build.ps1 dev`: `go build -o winmdview-dev.exe ./cmd/winmdview` (콘솔 창 유지, 디버그용)
- `.\build.ps1 test`: `go test -race -coverprofile=coverage.out ./...`
- `.\build.ps1 clean`: 빌드 산출물 정리
- 기존 `installer/build-msi.ps1`의 빌드 플래그와 일관성 유지
- 참고: WSL/Git Bash 사용자를 위해 Makefile을 보조 옵션으로 제공 가능

### Secondary Goal: WindowManager 핵심 구현

**영향 범위**: `internal/window/` (신규 패키지), `internal/app/`

| 작업 | 파일 | 설명 |
|------|------|------|
| Window 구조체 정의 | `internal/window/window.go` (신규) | 개별 윈도우 상태 캡슐화 |
| WindowManager 구현 | `internal/window/manager.go` (신규) | 윈도우 생성/추적/정리 |
| Window 테스트 | `internal/window/window_test.go` (신규) | 윈도우 생명주기 테스트 |
| Manager 테스트 | `internal/window/manager_test.go` (신규) | 관리자 동작 테스트 |
| Pipe 프로토콜 확장 | `internal/app/pipe.go` (수정) | `OPEN:<filepath>` 명령 프로토콜 |
| Pipe 테스트 업데이트 | `internal/app/pipe_test.go` (수정) | 새 프로토콜 테스트 |

**기술 접근**:

```
WindowManager
  ├── windows map[int]*Window    // ID로 윈도우 추적
  ├── mu      sync.RWMutex       // 동시성 보호
  ├── nextID  int                // 순차 ID 할당
  ├── OpenFile(path string)      // 새 윈도우 생성 또는 기존 활성화
  ├── CloseWindow(id int)        // 윈도우 정리 및 제거
  ├── GetWindows() []*Window     // 열린 윈도우 목록
  └── Shutdown()                 // 전체 종료

Window
  ├── ID       int
  ├── FilePath string
  ├── Viewer   viewer.Viewer     // WebView2 인스턴스
  ├── Server   *server.Server    // HTTP/WebSocket 서버
  ├── Watcher  *watcher.Watcher  // 파일 감시자
  └── Close()                    // 리소스 정리
```

**핵심 설계 결정**:
- 각 Window가 독립적인 Server 인스턴스를 소유 (포트 분리)
- Window.Close()에서 Server.Shutdown(), Watcher.Close() 순서로 정리
- WindowManager는 thread-safe하게 구현 (sync.RWMutex)

### Tertiary Goal: main.go 리팩토링 및 통합

**영향 범위**: `cmd/winmdview/main.go`, `internal/app/app.go`

| 작업 | 파일 | 설명 |
|------|------|------|
| main.go 리팩토링 | `cmd/winmdview/main.go` (수정) | WindowManager 기반 진입점으로 전환 |
| app.go 리팩토링 | `internal/app/app.go` (수정) | 단일 파일 로직을 WindowManager에 위임 |

**기술 접근**:
- 현재 `main.go`의 선형 흐름(서버 시작 -> 뷰어 생성 -> 감시 시작)을 WindowManager 기반으로 전환
- 첫 번째 인스턴스: WindowManager 생성 + 첫 번째 윈도우 오픈 + Pipe 서버 시작
- 두 번째 인스턴스: Pipe로 `OPEN:<filepath>` 전송 후 종료

**switchFile() 마이그레이션 전략** (단계적 제거):

각 윈도우는 하나의 파일만 표시하며, 윈도우 내 파일 전환 기능은 제공하지 않는다.

| 단계 | 작업 | 설명 |
|------|------|------|
| Phase 1 | 병행 운영 | `switchFile()` 유지 + `WindowManager.OpenFile()` 신규 추가. 기존 동작에 영향 없음 |
| Phase 2 | 호출자 전환 | Pipe 핸들러가 새 파일 요청 시 `switchFile()` 대신 `WindowManager.OpenFile()` 호출 |
| Phase 3 | switchFile() 제거 | 모든 호출자가 마이그레이션된 후 `switchFile()` 및 관련 코드 삭제 |

**설계 결정**: 윈도우 내 파일 전환(switchFile)은 멀티 윈도우 모델에서 불필요하다. 사용자가 다른 파일을 보려면 새 윈도우를 열고, 기존 윈도우는 닫으면 된다.

### Final Goal: 시스템 트레이 및 설정 통합

**영향 범위**: `internal/tray/`, `internal/config/`

| 작업 | 파일 | 설명 |
|------|------|------|
| 트레이 메뉴 확장 | `internal/tray/tray.go` (수정) | 동적 윈도우 목록 메뉴 |
| 트레이 테스트 업데이트 | `internal/tray/tray_test.go` (수정) | 동적 메뉴 테스트 |
| 설정 확장 | `internal/config/config.go` (수정) | 멀티 윈도우 상태 저장 |
| 설정 테스트 업데이트 | `internal/config/config_test.go` (수정) | 새 설정 필드 테스트 |

**기술 접근**:
- `tray.go`에서 WindowManager의 윈도우 목록을 구독하여 동적 메뉴 갱신
- 각 윈도우 메뉴 항목 클릭 시 해당 윈도우 활성화
- 설정에 `WindowStates []WindowState` 추가 (위치/크기/파일 경로)

## 파일 영향 분석

### 신규 파일

| 파일 | 목적 |
|------|------|
| `build.ps1` | PowerShell 빌드 스크립트 (-H windowsgui 포함) |
| `cmd/poc/multiwin/main.go` | go-webview2 멀티 인스턴스 PoC (검증 후 제거 가능) |
| `internal/window/manager.go` | WindowManager 중앙 관리자 |
| `internal/window/window.go` | 개별 윈도우 상태 |
| `internal/window/manager_test.go` | Manager 테스트 |
| `internal/window/window_test.go` | Window 테스트 |

### 수정 파일

| 파일 | 변경 내용 |
|------|----------|
| `cmd/winmdview/main.go` | WindowManager 기반 진입점으로 리팩토링 |
| `internal/app/app.go` | switchFile() 단계적 제거, WindowManager 연동 |
| `internal/app/pipe.go` | OPEN: 프로토콜 명령 추가 |
| `internal/app/pipe_test.go` | 새 프로토콜 테스트 |
| `internal/tray/tray.go` | 동적 윈도우 목록 메뉴 |
| `internal/tray/tray_test.go` | 동적 메뉴 테스트 |
| `internal/config/config.go` | WindowStates 필드 추가 |
| `internal/config/config_test.go` | 새 필드 테스트 |

## 기술 제약 및 위험 요소

### 위험 1: go-webview2 멀티 인스턴스 지원 (높음)

**설명**: go-webview2의 `Run()` 메서드는 블로킹 호출이다. 동일 프로세스에서 여러 WebView2 인스턴스를 실행할 수 있는지 사전 검증 필요.

**완화 전략**:
1. 구현 전 go-webview2 소스 코드 분석 (COM 초기화 방식 확인)
2. PoC (Proof of Concept): 두 개의 WebView2 인스턴스를 별도 goroutine에서 실행하는 테스트 코드 작성
3. 대안 A: WebView2 COM API를 직접 사용하여 단일 메시지 루프에서 여러 윈도우 관리
4. 대안 B: 멀티 프로세스 모델로 전환 (각 윈도우가 별도 프로세스)

### 위험 2: 리소스 누수 (중간)

**설명**: 윈도우 닫기 시 HTTP 서버, WebSocket 연결, 파일 감시자가 올바르게 정리되지 않을 수 있다.

**완화 전략**:
- `context.Context` 기반 생명주기 관리
- Window.Close()에서 정해진 순서로 리소스 해제 (Watcher -> WebSocket -> Server -> Viewer)
- 테스트에서 goroutine 누수 검증 (`goleak` 사용)

### 위험 3: 포트 고갈 (낮음)

**설명**: 윈도우별 독립 HTTP 서버 사용 시 포트를 소모한다.

**완화 전략**:
- 포트 0 바인딩으로 OS가 자동 할당
- 실사용에서 동시 10개 이상 윈도우는 드묾
- 포트 풀 관리는 현 단계에서 불필요

### 위험 4: 하위 호환성 (낮음)

**설명**: Named Pipe 프로토콜 변경 시 이전 버전과의 호환성.

**완화 전략**:
- `OPEN:` 접두사 없는 메시지는 기존 방식(파일 경로 직접 전달)으로 처리
- 새 버전은 항상 `OPEN:<filepath>` 형식으로 전송

## 테스트 전략

- **단위 테스트**: WindowManager, Window, Pipe 프로토콜 각각 독립 테스트
- **통합 테스트**: WindowManager + Server + Watcher 조합 테스트
- **수동 검증**: go-webview2 멀티 인스턴스 PoC
- **회귀 테스트**: 기존 단일 윈도우 동작이 멀티 윈도우에서도 동일하게 작동하는지 확인
- **커버리지 목표**: 새 코드 85% 이상
