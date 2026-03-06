---
spec_id: SPEC-WIN-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-WIN-001 리뷰

## 리뷰 요약

Windows 통합 SPEC으로, 컨텍스트 메뉴, 시스템 트레이, 단일 인스턴스 관리를 포함합니다. 기능 범위와 요구사항이 명확하게 정의되어 있으나, **CGO 제약과 systray 라이브러리의 충돌**, **systray.Run()과 WebView2 이벤트 루프의 goroutine 조율** 등 기술적 리스크가 높은 영역이 있습니다.

---

## 1. spec.md 이슈

### 🔴 Critical (수정 필요)

#### C-1: getlantern/systray의 CGO 의존성

- **위치**: spec.md §1.2 (line 33), §5 Constraints (line 180)
- **문제**: `github.com/getlantern/systray`는 **Windows에서도 CGO를 필요**로 할 수 있음 (내부적으로 시스템 API를 C 바인딩으로 호출). 이는 프로젝트 전체 제약인 "CGO 사용 금지 (pure Go 빌드)"와 직접 충돌
- **권장**: 다음 대안 중 하나를 채택:
  1. `github.com/energye/systray` (pure Go, CGO 불필요)
  2. `golang.org/x/sys/windows`를 직접 사용하여 Shell_NotifyIcon API 호출
  3. `lxn/walk` 라이브러리의 트레이 기능 사용
  4. CGO 제약을 systray에 한정하여 완화 (비권장)

### 🟡 Warning (보완 권장)

#### W-1: systray.Run()과 WebView2 이벤트 루프 충돌

- **위치**: spec.md §5 (line 185), plan.md §4.4 (lines 237-242)
- **문제**: systray.Run()은 "일부 플랫폼에서 메인 goroutine에서 실행되어야 한다"고 명시되어 있고, go-webview2도 메인 goroutine에서의 이벤트 루프를 필요로 함. 두 라이브러리의 메인 goroutine 점유가 충돌
- **권장**: 구체적인 조율 전략을 Specifications 또는 plan.md에 명시. 예: "systray.Run()을 별도 goroutine에서 실행하고, WebView2를 메인 goroutine에서 실행" 또는 "systray를 사용하지 않고 Windows Shell_NotifyIcon API를 직접 호출"

#### W-2: --register 실행 시 실행 파일 경로의 절대 경로 보장

- **위치**: spec.md §4.1 (lines 122-131)
- **문제**: `Register(exePath string)`에서 exePath를 사용자가 전달하는 것이 아니라 `os.Executable()`로 자동 획득해야 하는데, 이 부분이 명시되어 있지 않음. 심볼릭 링크를 통해 실행 시 `os.Executable()`이 링크 경로를 반환할 수 있음
- **권장**: "exePath는 `os.Executable()` + `filepath.EvalSymlinks()`로 획득하여 절대 경로와 실제 경로를 보장한다"를 Specifications에 추가

#### W-3: Named Pipe 보안 고려

- **위치**: spec.md §4.3 pipe.go (lines 152-156)
- **문제**: Named Pipe `\\.\pipe\WinMarkdownViewer`에 대한 보안 설명자(Security Descriptor)가 미명시. 기본 보안 설정 시 동일 세션의 다른 프로세스가 악의적 데이터를 전송할 수 있음
- **권장**: pipe를 통해 수신하는 데이터에 대한 검증 로직 명시 - 예: "수신 데이터가 유효한 파일 경로인지 검증하고, 파일이 존재하는지 확인 후에만 처리"

#### W-4: SPEC-WATCH-001과의 파일 전환 시 감시 대상 변경

- **위치**: spec.md §3.2 REQ-E-008 (line 95)
- **문제**: "파일 감시를 새 파일로 전환"한다고 했으나, Watcher 모듈의 WatchFile 변경 API가 SPEC-WATCH-001에 정의되어 있지 않음 (SPEC-WATCH-001의 Watcher는 생성 시 파일 경로를 받고, 변경 인터페이스가 없음)
- **권장**: SPEC-WATCH-001의 Watcher에 `SwitchFile(newPath string) error` 메서드 추가를 요구사항으로 명시하거나, 기존 Watcher를 Close하고 새로 생성하는 전략 명시

---

## 2. plan.md 이슈

### 🟡 Warning (보완 권장)

#### W-5: Task 2 레지스트리 테스트의 환경 격리

- **위치**: plan.md §2 Task 2 (lines 47-56)
- **문제**: "테스트 환경에서 HKCU 레지스트리 직접 조작 (테스트 전용 서브키 사용)"이라고 했으나, 테스트 전용 서브키의 경로가 미정의. 테스트 실패 시 레지스트리 잔류 키가 남을 수 있음
- **권장**: 테스트 전용 레지스트리 경로를 명시 (예: `HKCU\Software\Classes\.md-test\shell\WinMarkdownViewer_Test`), 그리고 `t.Cleanup()` 또는 `defer` 로 반드시 정리

#### W-6: Task 14 진입점 통합의 복잡도

- **위치**: plan.md §2 Task 14 (lines 147-158)
- **문제**: CLI 플래그 + Named Mutex + Named Pipe + systray + 뷰어 + 서버 + 파일 감시가 모두 하나의 main.go에 통합됨. 복잡도가 "높음"이지만 하나의 Task로 묶여 있어 분리가 어려움
- **권장**: Task 14를 하위 Task로 분리 - 14a(CLI 플래그 파싱), 14b(인스턴스 관리 로직), 14c(트레이 통합), 14d(전체 통합 테스트)

---

## 3. acceptance.md 이슈

### 🟡 Warning (보완 권장)

#### W-7: ACC-008 두 번째 인스턴스 파일 전달의 타이밍

- **위치**: acceptance.md §3 ACC-008 (lines 107-118)
- **문제**: "기존 인스턴스의 뷰어가 'another.md'로 전환하여 렌더링한다"의 예상 응답 시간이 미정의. Named Pipe 통신 + 파일 읽기 + 재렌더링 + 윈도우 포커스까지 얼마나 걸려야 하는지 기준 필요
- **권장**: "2초 이내에 전환 완료" 등의 응답 시간 기준 추가

#### W-8: ACC-014의 동작 정책 재확인

- **위치**: acceptance.md §5 ACC-014 (lines 179-187)
- **문제**: `--register test.md` 실행 시 "레지스트리 등록만 수행"이라는 정책. 사용자 관점에서는 `--register`와 파일 경로가 동시에 주어졌을 때 "등록 후 파일도 열어주기"를 기대할 수 있음
- **권장**: 이 정책이 의도적인 설계 결정인지 확인하고, 의도적이라면 `--register test.md` 실행 시 사용자에게 "등록이 완료되었습니다. 파일을 열려면 winmdview.exe test.md를 실행하세요" 안내 메시지 표시 고려

---

## 4. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🔴 Critical | 1건 | C-1 |
| 🟡 Warning | 8건 | W-1 ~ W-8 |

### 우선 조치 권장 순서

1. **systray 라이브러리 CGO 의존성 해결** (C-1) — 프로젝트 전체 제약과 직접 충돌
2. **systray.Run() / WebView2 goroutine 조율** (W-1) — 아키텍처 설계에 영향
3. **Watcher 파일 전환 API 정의** (W-4) — SPEC-WATCH-001 수정 필요
4. **Named Pipe 보안 검증** (W-3) — 보안 기본 사항
5. **나머지 Warning 순차 처리**
