---
spec_id: SPEC-WIN-001
type: acceptance-criteria
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-WIN-001 수용 기준

## 1. 컨텍스트 메뉴 시나리오

### ACC-001: 컨텍스트 메뉴 등록

```gherkin
Given winmdview.exe가 빌드되어 있다
  And 현재 Windows 레지스트리에 WinMarkdownViewer 키가 없다
When 사용자가 "winmdview.exe --register"를 실행한다
Then HKCU\Software\Classes\.md\shell\WinMarkdownViewer 레지스트리 키가 생성된다
  And (Default) 값이 "마크다운 뷰어로 열기"로 설정된다
  And command 서브키의 (Default) 값에 실행 파일의 절대 경로가 포함된다
  And HKLM 레지스트리는 변경되지 않는다
  And 등록 완료 메시지가 표시된다
```

### ACC-002: 컨텍스트 메뉴 해제

```gherkin
Given 컨텍스트 메뉴가 등록되어 있다
When 사용자가 "winmdview.exe --unregister"를 실행한다
Then HKCU\Software\Classes\.md\shell\WinMarkdownViewer 레지스트리 키가 삭제된다
  And 하위 키(command)도 모두 삭제된다
  And 해제 완료 메시지가 표시된다
```

### ACC-003: 파일 탐색기에서 컨텍스트 메뉴 사용

```gherkin
Given 컨텍스트 메뉴가 등록되어 있다
  And 파일 탐색기에 "readme.md" 파일이 있다
When 사용자가 "readme.md"를 우클릭한다
Then 컨텍스트 메뉴에 "마크다운 뷰어로 열기" 항목이 표시된다
When 사용자가 "마크다운 뷰어로 열기"를 클릭한다
Then WinMarkdownViewer가 "readme.md"를 열어 렌더링한다
```

### ACC-004: 일반 실행 시 레지스트리 미수정

```gherkin
Given 프로그램이 빌드되어 있다
When 사용자가 "winmdview.exe test.md"를 실행한다 (--register 플래그 없음)
Then 레지스트리에 어떤 변경도 발생하지 않는다
  And 파일이 정상적으로 뷰어에서 열린다
```

---

## 2. 시스템 트레이 시나리오

### ACC-005: 최소화 시 트레이로 이동

```gherkin
Given WinMarkdownViewer가 실행 중이고 윈도우가 표시되어 있다
When 사용자가 윈도우의 최소화 버튼을 클릭한다
Then 윈도우가 태스크바에서 사라진다
  And 시스템 트레이 영역에 WinMarkdownViewer 아이콘이 표시된다
  And 아이콘에 "WinMarkdownViewer" 툴팁이 표시된다
  And 프로세스가 종료되지 않고 백그라운드에서 계속 실행된다
```

### ACC-006: 트레이에서 복원 및 종료

```gherkin
Given 윈도우가 시스템 트레이로 최소화되어 있다
When 사용자가 트레이 아이콘을 더블클릭한다
Then 윈도우가 복원되어 전면에 표시된다
  And 트레이 아이콘은 유지된다

Given 윈도우가 시스템 트레이로 최소화되어 있다
When 사용자가 트레이 아이콘을 우클릭한다
Then 컨텍스트 메뉴에 "열기"와 "종료" 항목이 표시된다
When 사용자가 "열기"를 클릭한다
Then 윈도우가 복원되어 전면에 표시된다

Given 트레이 아이콘의 컨텍스트 메뉴가 열려 있다
When 사용자가 "종료"를 클릭한다
Then 트레이 아이콘이 제거된다
  And 모든 리소스가 정리된다
  And 프로세스가 종료된다
```

---

## 3. 단일 인스턴스 시나리오

### ACC-007: 첫 번째 인스턴스 시작

```gherkin
Given WinMarkdownViewer가 실행 중이지 않다
When 사용자가 "winmdview.exe test.md"를 실행한다
Then Named Mutex "WinMarkdownViewer_SingleInstance"가 생성된다
  And Named Pipe "\\.\pipe\WinMarkdownViewer" 서버가 시작된다
  And 뷰어 윈도우가 "test.md"를 렌더링하여 표시한다
```

### ACC-008: 두 번째 인스턴스에서 파일 전달

```gherkin
Given WinMarkdownViewer가 이미 "test.md"를 열고 실행 중이다
When 사용자가 "winmdview.exe another.md"를 실행한다 (두 번째 인스턴스)
Then 두 번째 인스턴스가 Named Mutex 획득에 실패한다
  And Named Pipe를 통해 "another.md" 경로를 기존 인스턴스에 전송한다
  And 두 번째 인스턴스 프로세스가 즉시 종료된다
  And 기존 인스턴스의 뷰어가 "another.md"로 2초 이내에 전환하여 렌더링한다
  And 기존 인스턴스의 윈도우가 전면에 표시된다
  And 두 번째 뷰어 윈도우는 생성되지 않는다
```

---

## 4. 에러 처리 시나리오

### ACC-009: 이미 등록된 상태에서 재등록

```gherkin
Given 컨텍스트 메뉴가 이미 등록되어 있다
  And 실행 파일이 다른 경로로 이동되었다
When 사용자가 새 경로에서 "winmdview.exe --register"를 실행한다
Then 레지스트리의 command 값이 새 실행 파일 경로로 갱신된다
  And "등록이 업데이트되었습니다" 메시지가 표시된다
```

### ACC-010: 미등록 상태에서 해제 시도

```gherkin
Given 컨텍스트 메뉴가 등록되어 있지 않다
When 사용자가 "winmdview.exe --unregister"를 실행한다
Then "등록된 컨텍스트 메뉴가 없습니다" 메시지가 표시된다
  And 프로세스가 정상적으로 종료된다
  And 레지스트리에 어떤 변경도 발생하지 않는다
```

### ACC-011: Named Pipe 서버 오류

```gherkin
Given 첫 번째 인스턴스가 실행 중이다
When Named Pipe 연결에서 잘못된 데이터(빈 문자열, 경로가 아닌 데이터)가 수신된다
Then 기존 인스턴스는 해당 데이터를 무시한다
  And Pipe 서버는 계속 새 연결을 대기한다
  And 현재 열린 파일은 영향받지 않는다
```

---

## 5. 엣지 케이스

### ACC-012: 첫 번째 인스턴스 종료 후 재시작

```gherkin
Given WinMarkdownViewer의 첫 번째 인스턴스가 실행 후 종료되었다
When 새로운 인스턴스로 "winmdview.exe test.md"를 실행한다
Then Named Mutex가 정상적으로 획득된다
  And Named Pipe 서버가 새로 시작된다
  And 뷰어가 정상적으로 동작한다
```

### ACC-013: 파일 전환 시 감시 대상 변경 (SPEC-WATCH-001 연동)

```gherkin
Given SPEC-WATCH-001이 구현되어 파일 감시가 동작 중이다
  And "test.md"를 감시하며 뷰어에서 표시 중이다
When Named Pipe를 통해 "another.md" 경로가 수신된다
Then 파일 감시가 "test.md"에서 "another.md"로 전환된다
  And 뷰어가 "another.md"를 렌더링하여 표시한다
  And "another.md"의 변경 사항이 실시간으로 반영된다
```

### ACC-014: CLI 플래그와 파일 경로 동시 전달

```gherkin
Given 프로그램이 빌드되어 있다
When 사용자가 "winmdview.exe --register test.md"를 실행한다
Then --register와 파일 경로가 동시에 주어지면 레지스트리 등록만 수행하고 안내 메시지를 표시한다
  And 파일 뷰어는 열리지 않는다
  And "레지스트리 등록이 완료되었습니다. 파일을 열려면 --register 없이 실행하세요." 안내 메시지가 표시된다
```

---

## 6. 빌드 및 실행 검증

### ACC-015: 빌드 성공

```gherkin
Given Go 1.26.0이 설치되어 있다
  And golang.org/x/sys/windows, energye/systray 의존성이 추가되어 있다
When "go build ./cmd/winmdview/" 명령을 실행한다
Then 빌드가 에러 없이 성공한다
  And CGO가 사용되지 않는다 (CGO_ENABLED=0)
```

### ACC-016: 테스트 통과

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test ./..." 명령을 실행한다
Then 모든 테스트가 통과한다
  And 레지스트리 조작 테스트가 포함된다
  And Named Mutex 테스트가 포함된다
  And Named Pipe 통신 테스트가 포함된다
  And 테스트 커버리지가 85% 이상이다
```

### ACC-017: 경쟁 조건 검사

```gherkin
Given 모든 테스트 파일이 작성되어 있다
When "go test -race ./..." 명령을 실행한다
Then 데이터 경쟁(race condition)이 감지되지 않는다
  And Named Pipe 서버의 동시 연결 처리에서 경쟁이 없다
```

---

## 7. Quality Gate (품질 게이트)

| 항목 | 기준 | 검증 방법 |
|------|------|-----------|
| 빌드 성공 | `go build` 에러 없음 | `go build ./cmd/winmdview/` |
| 테스트 통과 | 모든 테스트 PASS | `go test ./...` |
| 테스트 커버리지 | 85% 이상 (`internal/registry/`, `internal/app/` 기준, `internal/tray/`는 GUI 의존으로 제외) | `go test -coverprofile=coverage.out ./internal/registry/... ./internal/app/...` |
| 경쟁 조건 | 감지 없음 | `go test -race ./...` |
| 정적 분석 | 경고 없음 | `go vet ./...` |
| CGO 미사용 | CGO_ENABLED=0 빌드 성공 | `CGO_ENABLED=0 go build ./cmd/winmdview/` |
| 코드 품질 | lint 통과 | `golangci-lint run` (설치 시) |
| 레지스트리 정리 | 테스트 후 레지스트리 원상 복구 | 테스트 코드 내 cleanup defer |

---

## 8. Definition of Done (완료 정의)

- [ ] 모든 요구사항(REQ-*)에 대응하는 구현 코드 존재
- [ ] 모든 수용 기준(ACC-*)의 Given-When-Then 시나리오 검증 완료
- [ ] `go build ./cmd/winmdview/` 성공 (CGO_ENABLED=0)
- [ ] `go test ./...` 모든 테스트 통과
- [ ] `go test -race ./...` 경쟁 조건 미감지
- [ ] `go vet ./...` 경고 없음
- [ ] 테스트 커버리지 85% 이상 (`internal/registry/`, `internal/app/` 기준)
- [ ] `--register` 실행 후 파일 탐색기에서 컨텍스트 메뉴 수동 검증
- [ ] `--unregister` 실행 후 컨텍스트 메뉴 제거 수동 검증
- [ ] 두 번째 인스턴스 실행 시 파일 전달 동작 수동 검증
- [ ] 트레이 최소화/복원/종료 동작 수동 검증
- [ ] 테스트 실행 후 레지스트리가 원상 복구됨을 확인
