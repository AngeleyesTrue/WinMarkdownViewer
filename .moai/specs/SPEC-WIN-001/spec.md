---
id: SPEC-WIN-001
title: "Windows Integration - Context Menu, System Tray, Single Instance"
version: 1.0.0
status: completed
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P2
lifecycle: spec-first
tags: [windows, registry, systray, single-instance, named-pipe]
---

# SPEC-WIN-001: Windows 통합 - 컨텍스트 메뉴, 시스템 트레이, 단일 인스턴스

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- 관리자 권한 불필요 (HKCU 레지스트리 사용)

### 1.2 기술 스택
- Go 1.26.0 (CGO 불필요)
- golang.org/x/sys/windows - Windows API 바인딩 (레지스트리, Named Mutex, Named Pipe)
- github.com/energye/systray - 크로스 플랫폼 시스템 트레이 (energye/systray는 pure Go 구현으로 CGO 불필요)
- github.com/jchv/go-webview2 - WebView2 바인딩 (SPEC-UI-001에서 구현 완료)

### 1.3 프로젝트 구조 변경
```
WinMarkdownViewer/
  cmd/winmdview/main.go              # 수정: CLI 플래그, 단일 인스턴스 체크 추가
  internal/registry/registry.go      # [신규] Windows 레지스트리 조작
  internal/registry/registry_test.go # [신규] 레지스트리 테스트
  internal/tray/tray.go              # [신규] 시스템 트레이 관리
  internal/tray/tray_test.go         # [신규] 트레이 테스트
  internal/app/instance.go           # [신규] 단일 인스턴스 관리
  internal/app/instance_test.go      # [신규] 단일 인스턴스 테스트
  internal/app/pipe.go               # [신규] Named Pipe 통신
  internal/app/pipe_test.go          # [신규] Pipe 테스트
  assets/icon.ico                    # [신규] 트레이 아이콘 (16x16, 32x32)
  go.mod                             # 수정: 의존성 추가
```

---

## 2. Assumptions (가정)

- A1: SPEC-UI-001이 완료되어 기본 뷰어가 동작한다.
- A2: SPEC-WATCH-001이 완료되어 내장 HTTP 서버가 존재한다 (권장, 필수는 아님).
- A3: 레지스트리 변경은 HKCU (현재 사용자) 범위에서 수행하여 관리자 권한이 불필요하다.
- A4: 프로그램의 실행 파일 경로는 레지스트리 등록 시점에 확정된다.
- A5: 시스템 트레이 아이콘 리소스(.ico)는 go:embed로 바이너리에 포함된다.
- A6: 단일 인스턴스 관리는 Windows Named Mutex + Named Pipe를 사용한다.
- A7: 사용자가 WinMarkdownViewer를 이동하거나 이름을 변경하면 레지스트리를 재등록해야 한다.

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous (항상 활성)

- **REQ-U-001**: 시스템은 **항상** 레지스트리 작업을 HKCU 범위에서만 수행해야 한다 (관리자 권한 불필요).
- **REQ-U-002**: 시스템은 **항상** 시스템 트레이 아이콘에 "WinMarkdownViewer" 툴팁을 표시해야 한다.
- **REQ-U-003**: 시스템은 **항상** Named Mutex 이름으로 `WinMarkdownViewer_SingleInstance`를 사용해야 한다.

### 3.2 Event-Driven (이벤트 기반)

- **REQ-E-001**: **WHEN** 사용자가 `--register` CLI 플래그로 프로그램을 실행하면, **THEN** 시스템은 Windows 레지스트리에 `.md` 파일 우클릭 컨텍스트 메뉴를 등록하고 완료 메시지를 표시해야 한다.
  - 레지스트리 경로: `HKCU\Software\Classes\.md\shell\WinMarkdownViewer`
  - 메뉴 텍스트: "마크다운 뷰어로 열기"
  - 명령: `"{실행파일 절대경로}" "%1"`

- **REQ-E-002**: **WHEN** 사용자가 `--unregister` CLI 플래그로 프로그램을 실행하면, **THEN** 시스템은 등록된 컨텍스트 메뉴 레지스트리 키를 제거하고 완료 메시지를 표시해야 한다.

- **REQ-E-003**: **WHEN** 사용자가 파일 탐색기에서 .md 파일을 우클릭하고 "마크다운 뷰어로 열기"를 선택하면, **THEN** 시스템은 해당 파일을 뷰어에서 열어야 한다.

- **REQ-E-004**: **WHEN** 뷰어가 이미 실행 중인 상태에서 새로운 .md 파일이 열리면, **THEN** 새 인스턴스는 Named Pipe를 통해 기존 인스턴스에 파일 경로를 전달하고, 기존 인스턴스가 해당 파일로 전환해야 한다.

- **REQ-E-005**: **WHEN** 사용자가 WebView2 윈도우의 최소화 버튼을 클릭하면, **THEN** 윈도우가 태스크바에서 사라지고 시스템 트레이로 최소화되어야 한다.

- **REQ-E-006**: **WHEN** 사용자가 시스템 트레이 아이콘을 더블클릭하면, **THEN** 윈도우가 트레이에서 복원되어 전면에 표시되어야 한다.

- **REQ-E-007**: **WHEN** 사용자가 시스템 트레이 아이콘을 우클릭하면, **THEN** 컨텍스트 메뉴가 표시되어야 한다:
  - "열기" - 윈도우 복원
  - "종료" - 프로그램 종료

- **REQ-E-008**: **WHEN** 기존 인스턴스가 Named Pipe를 통해 새 파일 경로를 수신하면, **THEN** 새 파일을 렌더링하고, 파일 감시를 새 파일로 전환하며, 윈도우를 전면에 표시해야 한다.

### 3.3 Unwanted (금지 행위)

- **REQ-N-001**: 시스템은 HKLM(전체 시스템) 레지스트리를 수정**하지 않아야 한다**.
- **REQ-N-002**: 시스템은 `--register`/`--unregister` 없이 일반 실행 시 레지스트리를 수정**하지 않아야 한다**.
- **REQ-N-003**: 시스템은 트레이로 최소화 시 프로세스를 종료**하지 않아야 한다** (백그라운드 유지).
- **REQ-N-004**: 시스템은 이미 실행 중인 인스턴스가 있을 때 두 번째 뷰어 윈도우를 생성**하지 않아야 한다**.

### 3.4 State-Driven (상태 기반)

- **REQ-S-001**: **IF** Named Mutex 획득에 성공하면 (첫 번째 인스턴스), **THEN** 시스템은 Named Pipe 서버를 시작하고 정상적으로 뷰어를 실행해야 한다.
- **REQ-S-002**: **IF** Named Mutex 획득에 실패하면 (두 번째+ 인스턴스), **THEN** 시스템은 Named Pipe 클라이언트로 파일 경로를 기존 인스턴스에 전송하고 즉시 종료해야 한다.
- **REQ-S-003**: **IF** `--register` 실행 시 동일한 레지스트리 키가 이미 존재하면, **THEN** 시스템은 기존 키를 업데이트하고 (실행 파일 경로 갱신) 성공 메시지를 표시해야 한다.
- **REQ-S-004**: **IF** `--unregister` 실행 시 레지스트리 키가 존재하지 않으면, **THEN** 시스템은 "등록된 컨텍스트 메뉴가 없습니다" 메시지를 표시해야 한다.

### 3.5 Optional (선택 사항)

- **REQ-O-001**: **가능하면** `--set-default` CLI 플래그로 .md 파일의 기본 프로그램으로 등록할 수 있어야 한다.
- **REQ-O-002**: **가능하면** 트레이 아이콘에 현재 열린 파일명을 툴팁으로 표시해야 한다.
- **REQ-O-003**: **가능하면** 레지스트리 등록 시 커스텀 아이콘을 컨텍스트 메뉴에 표시해야 한다.

---

## 4. Specifications (상세 명세)

### 4.1 레지스트리 관리 모듈 (`internal/registry/`)

- `Register(exePath string) error`: 컨텍스트 메뉴 레지스트리 등록
  - `HKCU\Software\Classes\.md\shell\WinMarkdownViewer` 키 생성
  - `(Default)` 값: "마크다운 뷰어로 열기"
  - `command` 서브키의 `(Default)` 값: `"{exePath}" "%1"`
  - `Icon` 값: `"{exePath}",0` (Optional)
- `Unregister() error`: 레지스트리 키 제거
  - `WinMarkdownViewer` 키와 하위 키 전체 삭제
- `IsRegistered() (bool, error)`: 현재 등록 상태 확인
- `SetDefault(exePath string) error`: .md 파일 기본 프로그램 설정 (Optional)

### 4.2 시스템 트레이 모듈 (`internal/tray/`)

- `NewTray(iconData []byte) (*Tray, error)`: 트레이 인스턴스 생성
- `Run(onOpen func(), onQuit func())`: 트레이 이벤트 루프 시작
  - 메뉴 항목: "열기", "종료"
  - 더블클릭: onOpen 콜백 호출
- `SetTooltip(text string)`: 툴팁 텍스트 변경
- `Quit()`: 트레이 제거 및 종료
- energye/systray 라이브러리를 래핑하여 사용

### 4.3 단일 인스턴스 관리 (`internal/app/`)

**instance.go:**
- `TryLock() (bool, error)`: Named Mutex (`WinMarkdownViewer_SingleInstance`) 획득 시도
  - 성공: true 반환, 첫 번째 인스턴스
  - 실패: false 반환, 이미 실행 중
- `Unlock() error`: Mutex 해제

**pipe.go:**
- `ListenPipe(ctx context.Context, handler func(filePath string)) error`: Named Pipe 서버 시작
  - Pipe 이름: `\\.\pipe\WinMarkdownViewer`
  - 새 연결 수신 시 파일 경로 읽기 -> handler 콜백 호출
- `SendPath(filePath string) error`: Named Pipe 클라이언트로 파일 경로 전송
  - Pipe에 연결 -> 파일 경로 쓰기 -> 연결 종료

### 4.4 진입점 수정 (`cmd/winmdview/main.go`)

- CLI 플래그 파싱: `--register`, `--unregister`, `--set-default`
- 실행 흐름:
  1. CLI 플래그 체크 -> 레지스트리 작업 후 종료
  2. Named Mutex 획득 시도
     - 실패: Named Pipe로 파일 경로 전송 후 종료
     - 성공: Named Pipe 서버 시작
  3. 시스템 트레이 초기화
  4. 뷰어 + HTTP 서버 + 파일 감시 시작 (기존 로직)
  5. Pipe에서 새 파일 수신 -> 뷰어 전환

### 4.5 아이콘 리소스 (`assets/`)

- `icon.ico`: 16x16, 32x32 해상도의 ICO 파일
- go:embed로 바이너리에 포함
- 시스템 트레이 아이콘 및 컨텍스트 메뉴 아이콘으로 사용

---

## 5. Constraints (제약사항)

- CGO 사용 금지 (pure Go 빌드 유지)
- 관리자 권한 불필요 (HKCU 범위만 사용)
- Windows 전용 (Linux/macOS 빌드 제외 가능, build tags 사용)
- Named Pipe는 Windows API 전용 기능
- 이 SPEC에서 MSI 설치 프로그램, 설정 UI는 범위 밖
- systray.Run()을 별도 goroutine에서 실행하고, WebView2를 메인 goroutine에서 실행한다. systray와 WebView2 간 통신은 channel을 통해 수행한다
- exePath는 os.Executable() + filepath.EvalSymlinks()로 획득하여 절대 경로와 실제 경로를 보장한다
- Named Pipe 수신 데이터가 유효한 파일 경로인지 검증하고, 파일이 존재하는지 확인 후에만 처리한다. 최대 수신 크기를 4096바이트로 제한한다
- **Cross-SPEC 의존성**: SPEC-WATCH-001의 Watcher가 SwitchFile API를 필요로 한다. 파일 전환 시 Watcher의 감시 대상 변경을 위한 인터페이스를 제공해야 한다

---

## 6. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 시나리오 |
|-------------|-----------|-----------------|
| REQ-U-001 | internal/registry/registry.go | ACC-001 |
| REQ-U-002 | internal/tray/tray.go | ACC-005 |
| REQ-U-003 | internal/app/instance.go | ACC-007 |
| REQ-E-001 | internal/registry/registry.go | ACC-001 |
| REQ-E-002 | internal/registry/registry.go | ACC-002 |
| REQ-E-003 | internal/registry/registry.go | ACC-003 |
| REQ-E-004 | internal/app/instance.go, internal/app/pipe.go | ACC-007, ACC-008 |
| REQ-E-005 | internal/tray/tray.go, internal/viewer/viewer.go | ACC-005 |
| REQ-E-006 | internal/tray/tray.go | ACC-006 |
| REQ-E-007 | internal/tray/tray.go | ACC-006 |
| REQ-E-008 | internal/app/pipe.go, cmd/winmdview/main.go | ACC-008 |
| REQ-N-001 | internal/registry/registry.go | ACC-001 |
| REQ-N-002 | cmd/winmdview/main.go | ACC-004 |
| REQ-N-003 | internal/tray/tray.go | ACC-005 |
| REQ-N-004 | internal/app/instance.go | ACC-007 |
| REQ-S-001 | internal/app/instance.go, internal/app/pipe.go | ACC-007 |
| REQ-S-002 | internal/app/instance.go, internal/app/pipe.go | ACC-008 |
| REQ-S-003 | internal/registry/registry.go | ACC-009 |
| REQ-S-004 | internal/registry/registry.go | ACC-010 |
| REQ-O-001 | internal/registry/registry.go | - |
| REQ-O-002 | internal/tray/tray.go | - |
| REQ-O-003 | internal/registry/registry.go | - |

---

## Implementation Notes

### 구현 완료 요약 (2026-03-06)

**구현된 기능:**
- REQ-E-001 ~ REQ-E-008: 컨텍스트 메뉴 등록/해제, 시스템 트레이, 단일 인스턴스 전체 구현
- REQ-U-001 ~ REQ-U-003: HKCU 전용, 트레이 툴팁, Named Mutex 이름 준수
- REQ-N-001 ~ REQ-N-004: 금지 행위 전체 준수
- REQ-S-001 ~ REQ-S-004: 상태 기반 요구사항 전체 구현
- REQ-O-001: Open With 프로그램 선택 목록 등록 지원 (--set-default)

**계획 대비 변경사항:**
- 레지스트리 경로를 `HKCU\Software\Classes\.md\shell\` 에서 `HKCU\Software\Classes\SystemFileAssociations\.md\shell\`로 변경 (Windows 최신 파일 연결 모범 사례 적용)
- `internal/app/constants.go` 추가 (리팩토링 단계에서 공통 상수 분리)
- `internal/viewer/viewer.go` 확장 (트레이 최소화/복원 지원)

**보안 강화 (2026-03-09):**
- Registry `validateExePath()` 추가: 빈 경로, null 바이트, 상대 경로 거부
- Pipe 입력 검증: null 바이트 트리밍, `filepath.Clean`, `filepath.IsAbs`, `os.Stat` 검증
- ListenPipe 지수 백오프: 100ms~5s 백오프 + 연속 에러 10회 제한으로 리소스 고갈 방지
- errgroup 고루틴 관리: 파일 감시, Named Pipe, 시스템 트레이 고루틴을 errgroup으로 조율
- `golang.org/x/sync` 의존성 추가

**테스트 결과:**
- 전체 테스트: 8 패키지 통과
- 커버리지: internal/app 85.3%, internal/registry 83.1%, internal/config 88.9%, internal/server 89.5%
- go vet: 클린
- CGO: 미사용 (순수 Go 빌드)

**제한사항:**
- CGO 비활성으로 `-race` 플래그 테스트 불가
- `internal/tray` 커버리지 32.0% (GUI 의존 코드)
- `cmd/winmdview` 커버리지 5.9% (통합 테스트 한계)
