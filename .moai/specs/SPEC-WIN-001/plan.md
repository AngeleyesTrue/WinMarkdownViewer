---
spec_id: SPEC-WIN-001
type: implementation-plan
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-WIN-001 구현 계획

## 1. 구현 전략

### 1.1 개발 방법론

TDD (Test-Driven Development) - RED-GREEN-REFACTOR 사이클에 따라 구현합니다.

### 1.2 전체 접근 방식

기능별 모듈 분리 후 독립적으로 구현, 최종 통합합니다:

1. 레지스트리 관리 모듈 (internal/registry/) - 컨텍스트 메뉴 등록/해제
2. 단일 인스턴스 관리 모듈 (internal/app/) - Named Mutex + Named Pipe
3. 시스템 트레이 모듈 (internal/tray/) - 트레이 아이콘 및 메뉴
4. 진입점 통합 (cmd/winmdview/main.go) - CLI 플래그, 인스턴스 관리, 트레이

### 1.3 선행 조건

- SPEC-UI-001 구현 완료 (필수)
  - 기본 뷰어 (goldmark + WebView2)
  - CLI 진입점 구조
- SPEC-WATCH-001 구현 완료 (권장)
  - 내장 HTTP 서버 (단일 인스턴스 파일 전환 시 서버 활용 가능)
  - 파일 감시 모듈 (새 파일 전환 시 감시 대상 변경)

---

## 2. 마일스톤

### Primary Goal: 레지스트리 관리 모듈

**Task 1: golang.org/x/sys/windows 의존성 추가**
- 파일: `go.mod`
- 작업: `go get golang.org/x/sys/windows`
- 의존성: 없음

**Task 2: Registry 테스트 작성 (RED)**
- 파일: `internal/registry/registry_test.go`
- 작업:
  - Register 호출 시 HKCU 레지스트리 키 생성 확인
  - Unregister 호출 시 레지스트리 키 삭제 확인
  - IsRegistered 상태 확인 테스트
  - 이미 등록된 상태에서 재등록 시 경로 갱신 테스트
  - 미등록 상태에서 Unregister 시 동작 테스트
- 참고: 테스트 전용 레지스트리: `HKCU\Software\Classes\.md-test\shell\WinMarkdownViewer_Test`. t.Cleanup()으로 반드시 정리
- 의존성: Task 1

**Task 3: Registry 구현 (GREEN)**
- 파일: `internal/registry/registry.go`
- 작업:
  - `Register(exePath string) error` - 컨텍스트 메뉴 등록
  - `Unregister() error` - 레지스트리 키 삭제
  - `IsRegistered() (bool, error)` - 등록 상태 확인
  - golang.org/x/sys/windows/registry 패키지 사용
  - 레지스트리 경로: `HKCU\Software\Classes\.md\shell\WinMarkdownViewer`
- 의존성: Task 2

**Task 4: Registry 리팩토링 (REFACTOR)**
- 파일: `internal/registry/registry.go`
- 작업: 레지스트리 경로 상수화, 에러 메시지 개선
- 의존성: Task 3

### Secondary Goal: 단일 인스턴스 관리 모듈

**Task 5: Instance 테스트 작성 (RED)**
- 파일: `internal/app/instance_test.go`
- 작업:
  - Named Mutex 획득 성공 테스트 (첫 번째 인스턴스)
  - Named Mutex 획득 실패 테스트 (두 번째 인스턴스)
  - Mutex 해제 후 재획득 테스트
- 의존성: Task 1

**Task 6: Instance 구현 (GREEN)**
- 파일: `internal/app/instance.go`
- 작업:
  - `TryLock() (bool, error)` - Named Mutex 획득
  - `Unlock() error` - Mutex 해제
  - Windows API: `CreateMutex`, `WaitForSingleObject`
  - Mutex 이름: `WinMarkdownViewer_SingleInstance`
- 의존성: Task 5

**Task 7: Pipe 테스트 작성 (RED)**
- 파일: `internal/app/pipe_test.go`
- 작업:
  - Named Pipe 서버/클라이언트 통신 테스트
  - 파일 경로 전송 및 수신 테스트
  - 서버 context 취소 시 정상 종료 테스트
  - 동시 다수 클라이언트 연결 테스트
- 의존성: Task 1

**Task 8: Pipe 구현 (GREEN)**
- 파일: `internal/app/pipe.go`
- 작업:
  - `ListenPipe(ctx context.Context, handler func(filePath string)) error`
  - `SendPath(filePath string) error`
  - Pipe 이름: `\\.\pipe\WinMarkdownViewer`
  - Windows API: `CreateNamedPipe`, `ConnectNamedPipe`
- 의존성: Task 7

**Task 9: Instance + Pipe 리팩토링 (REFACTOR)**
- 파일: `internal/app/instance.go`, `internal/app/pipe.go`
- 작업: 공통 상수 분리, 에러 처리 일관성, 타임아웃 설정
- 의존성: Task 6, Task 8

### Final Goal: 시스템 트레이 + 통합

**Task 10: energye/systray 의존성 추가**
- 파일: `go.mod`
- 작업: `go get github.com/energye/systray`
- 의존성: 없음

**Task 11: 아이콘 리소스 준비**
- 파일: `assets/icon.ico`, `assets/embed.go`
- 작업:
  - 16x16, 32x32 ICO 파일 생성 (단순 "MD" 텍스트 아이콘)
  - go:embed 디렉티브로 바이너리 포함
- 의존성: Task 10

**Task 12: Tray 테스트 작성 (RED)**
- 파일: `internal/tray/tray_test.go`
- 작업:
  - Tray 인스턴스 생성 테스트 (아이콘 데이터)
  - 콜백 설정 테스트
  - 참고: systray.Run()은 GUI 의존으로 인터페이스 기반 mock 테스트
- 의존성: Task 11

**Task 13: Tray 구현 (GREEN)**
- 파일: `internal/tray/tray.go`
- 작업:
  - `NewTray(iconData []byte) (*Tray, error)` 생성자
  - `Run(onOpen func(), onQuit func())` 이벤트 루프
  - `SetTooltip(text string)` 툴팁 변경
  - `Quit()` 트레이 종료
  - systray.Run() 래핑, 메뉴 항목 ("열기", "종료") 설정
- 의존성: Task 12

**Task 14a: CLI 플래그 파싱**
- 파일: `cmd/winmdview/main.go`
- 작업:
  - CLI 플래그 추가: `--register`, `--unregister`, `--set-default`
  - 플래그 파싱 로직 분리 및 테스트
- 의존성: Task 3

**Task 14b: 인스턴스 관리 로직**
- 파일: `cmd/winmdview/main.go`
- 작업:
  - TryLock() -> 실패 시 SendPath() 후 종료
  - 성공 시: ListenPipe() 서버 시작
  - Pipe 수신 -> 파일 전환 로직
- 의존성: Task 8, Task 14a

**Task 14c: 트레이 통합**
- 파일: `cmd/winmdview/main.go`
- 작업:
  - 시스템 트레이 초기화 및 이벤트 연결
  - 최소화 시 트레이로 이동 이벤트 처리
  - 트레이 "열기" -> 윈도우 복원, "종료" -> shutdown
- 의존성: Task 13, Task 14b

**Task 14d: 전체 통합 테스트**
- 파일: `cmd/winmdview/main_test.go`
- 작업:
  - CLI 플래그 파싱 + 인스턴스 관리 + 트레이 통합 흐름 테스트
  - 실행 흐름 전체 시나리오 검증
- 의존성: Task 14a, Task 14b, Task 14c

**Task 15: 빌드 및 통합 검증**
- 작업:
  - `go build ./cmd/winmdview/` 성공 확인
  - `go test -race ./...` 경쟁 조건 검사
  - `--register` 실행 후 파일 탐색기에서 컨텍스트 메뉴 확인
  - 이중 실행 시 단일 인스턴스 동작 확인
  - 트레이 최소화/복원 동작 확인
- 의존성: Task 14

---

## 3. 파일 영향 분석

| 파일 | 작업 유형 | 복잡도 | 관련 Task |
|------|-----------|--------|-----------|
| `go.mod` | 수정 | 낮음 | Task 1, 10 |
| `internal/registry/registry.go` | 신규 생성 | 중간 | Task 3, 4 |
| `internal/registry/registry_test.go` | 신규 생성 | 중간 | Task 2 |
| `internal/app/instance.go` | 신규 생성 | 중간 | Task 6, 9 |
| `internal/app/instance_test.go` | 신규 생성 | 중간 | Task 5 |
| `internal/app/pipe.go` | 신규 생성 | 높음 | Task 8, 9 |
| `internal/app/pipe_test.go` | 신규 생성 | 높음 | Task 7 |
| `internal/tray/tray.go` | 신규 생성 | 중간 | Task 13 |
| `internal/tray/tray_test.go` | 신규 생성 | 낮음 | Task 12 |
| `assets/icon.ico` | 신규 생성 | 낮음 | Task 11 |
| `assets/embed.go` | 신규 생성 | 낮음 | Task 11 |
| `cmd/winmdview/main.go` | 수정 | 높음 | Task 14 |

**총 파일 수**: 12개 (10개 신규 생성 + 2개 수정)
**전체 복잡도**: 높음

---

## 4. 기술적 접근

### 4.1 레지스트리 구조

```
HKCU\Software\Classes\.md\
  shell\
    WinMarkdownViewer\
      (Default) = "마크다운 뷰어로 열기"
      Icon = "C:\path\to\winmdview.exe",0     [Optional]
      command\
        (Default) = "\"C:\path\to\winmdview.exe\" \"%1\""
```

### 4.2 단일 인스턴스 흐름

```
[프로그램 시작]
      |
      v
 Named Mutex 획득 시도
      |
      ├─ 성공 (첫 번째 인스턴스) ──────────────────────┐
      │   └── Named Pipe 서버 시작                      |
      │       └── 새 연결 대기                          |
      │                                                 v
      └─ 실패 (두 번째+ 인스턴스)           [뷰어 + 트레이 실행]
          └── Named Pipe 클라이언트                      ^
              └── 파일 경로 전송 ──pipe──> 기존 인스턴스  |
              └── 즉시 종료                 파일 전환 ────┘
```

### 4.3 시스템 트레이 상태 전이

```
[정상 표시] ──최소화 버튼──> [트레이 최소화]
     ^                            |
     |                            |
  트레이 "열기"              트레이 "종료"
  또는 더블클릭                    |
     |                            v
     └──────────────────── [프로그램 종료]
```

### 4.4 systray.Run() 통합 주의사항

`systray.Run()`은 일부 플랫폼에서 메인 goroutine에서 실행되어야 한다. 구현 시:
- `systray.Run(onReady, onExit)` 패턴 사용 (energye/systray)
- systray.Run()을 별도 goroutine에서 실행
- WebView2를 메인 goroutine에서 실행
- systray와 WebView2 간 통신은 channel을 통해 수행

---

## 5. 리스크 및 대응

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| energye/systray가 최신 Go 1.26과 호환되지 않을 수 있음 | 높음 | 호환 버전 확인, pure Go 구현이므로 CGO 이슈 없음 |
| systray.Run()과 WebView2 이벤트 루프의 goroutine 충돌 | 높음 | systray.Run()을 별도 goroutine에서 실행, WebView2를 메인 goroutine에서 실행, channel로 통신 |
| Named Pipe의 보안 (다른 프로세스가 접근 가능) | 중간 | 기본 보안 설명자 사용, 파일 경로 입력만 수용 |
| 레지스트리 테스트가 실제 시스템 레지스트리를 오염시킬 수 있음 | 중간 | 테스트 전용 서브키 사용, cleanup defer 필수 |
| 실행 파일 이동 시 레지스트리의 경로가 무효화됨 | 낮음 | `--register` 재실행 안내, 설치 프로그램(SPEC-INSTALL-001)에서 해결 |

---

## 6. SPEC 의존성

### SPEC-UI-001 의존성 (필수)

| SPEC-UI-001 구현물 | 이 SPEC에서의 활용 |
|--------------------|-------------------|
| `internal/viewer/viewer.go` | 트레이 최소화/복원, 파일 전환 제어 |
| `cmd/winmdview/main.go` | CLI 플래그 추가, 인스턴스 관리 통합 |
| `internal/markdown/renderer.go` | 파일 전환 시 재렌더링 |

### SPEC-WATCH-001 의존성 (권장)

| SPEC-WATCH-001 구현물 | 이 SPEC에서의 활용 |
|----------------------|-------------------|
| `internal/watcher/watcher.go` | 파일 전환 시 감시 대상 변경 |
| `internal/server/server.go` | 파일 전환 시 새 파일 내용 서빙 |

---

## 7. 범위 외 (Out of Scope)

| 기능 | 예정 SPEC |
|------|-----------|
| MSI 설치 프로그램 | SPEC-INSTALL-001 |
| 설정 UI / 환경설정 | SPEC-CONFIG-001 |
| 다크 모드 / 테마 | SPEC-THEME-001 |
| 자동 업데이트 | 미정 |

---

## 8. 다음 단계

1. `/moai:2-run SPEC-WIN-001` 실행하여 TDD 사이클로 구현 시작
2. 구현 완료 후 `/moai:3-sync SPEC-WIN-001` 실행하여 문서화
