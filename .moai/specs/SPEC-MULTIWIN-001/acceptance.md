---
id: SPEC-MULTIWIN-001
type: acceptance
version: 0.2.0
---

# SPEC-MULTIWIN-001: 인수 기준

## AC1: 콘솔 창 억제

### AC1.1: 릴리스 빌드 시 콘솔 창 미표시

```gherkin
Given build.ps1이 프로젝트 루트에 존재하고
When 개발자가 `.\build.ps1 build` 명령을 실행하면
Then winmdview.exe가 `-H windowsgui` 플래그로 빌드되고
And 생성된 실행 파일을 더블클릭하면 콘솔 창이 표시되지 않는다
```

### AC1.2: 개발 빌드 시 콘솔 창 표시

```gherkin
Given build.ps1이 프로젝트 루트에 존재하고
When 개발자가 `.\build.ps1 dev` 명령을 실행하면
Then winmdview-dev.exe가 콘솔 창과 함께 실행 가능하고
And 디버그 로그가 콘솔에 출력된다
```

### AC1.3: 파일 연결을 통한 실행

```gherkin
Given winmdview.exe가 `.\build.ps1 build`로 빌드되었고
And .md 파일이 winmdview.exe에 연결되어 있을 때
When 사용자가 .md 파일을 더블클릭하면
Then 마크다운 뷰어 창만 표시되고
And 콘솔 창은 전혀 나타나지 않는다
```

## AC2: 멀티 윈도우 생성

### AC2.1: 새 파일을 별도 윈도우에서 열기

```gherkin
Given WinMarkdownViewer가 file1.md를 표시하고 있을 때
When 사용자가 file2.md를 파일 연결 또는 컨텍스트 메뉴로 열면
Then file1.md 윈도우는 변경 없이 유지되고
And file2.md가 새로운 별도 윈도우에서 열린다
And 두 윈도우 모두 독립적으로 상호작용할 수 있다
```

### AC2.2: Named Pipe를 통한 새 윈도우 생성

```gherkin
Given 첫 번째 인스턴스가 실행 중일 때
When 두 번째 인스턴스가 file3.md 경로와 함께 실행되면
Then 두 번째 인스턴스는 Named Pipe로 "OPEN:file3.md" 명령을 전송하고
And 두 번째 인스턴스는 즉시 종료되고
And 첫 번째 인스턴스가 file3.md를 새 윈도우에서 연다
```

### AC2.3: 동일 파일 중복 열기 방지

```gherkin
Given WinMarkdownViewer가 file1.md를 표시하고 있을 때
When 사용자가 동일한 file1.md를 다시 열면
Then 새 윈도우가 생성되지 않고
And 기존 file1.md 윈도우가 포그라운드로 활성화된다
```

### AC2.4: 기존 콘텐츠 교체 금지

```gherkin
Given WinMarkdownViewer가 file1.md를 표시하고 있을 때
When 새 파일 file2.md가 요청되면
Then file1.md 윈도우의 콘텐츠는 변경되지 않고
And file2.md는 새 윈도우에서 표시된다
```

## AC3: 윈도우 생명주기

### AC3.1: 개별 윈도우 닫기

```gherkin
Given file1.md와 file2.md가 각각 별도 윈도우에서 열려 있을 때
When 사용자가 file1.md 윈도우를 닫으면
Then file1.md 윈도우만 닫히고
And file2.md 윈도우는 영향 없이 계속 표시되고
And file1.md의 HTTP 서버와 파일 감시자가 정리된다
```

### AC3.2: 마지막 윈도우 닫기

```gherkin
Given file1.md 윈도우만 열려 있을 때
When 사용자가 file1.md 윈도우를 닫으면
Then 윈도우는 닫히고
And 시스템 트레이 아이콘은 유지되고
And 프로세스는 계속 실행된다 (새 파일 요청 대기)
```

### AC3.3: 시스템 트레이에서 종료

```gherkin
Given 여러 윈도우가 열려 있고 시스템 트레이 아이콘이 표시될 때
When 사용자가 시스템 트레이 메뉴에서 "종료"를 선택하면
Then 모든 열린 윈도우가 닫히고
And 모든 리소스(서버, 감시자, WebSocket)가 정리되고
And 프로세스가 종료된다
```

### AC3.4: 리소스 누수 없음

```gherkin
Given 10개의 윈도우가 순차적으로 열렸다가 닫혔을 때
When 모든 윈도우가 닫힌 후
Then goroutine 수가 윈도우 열기 전 수준으로 복원되고
And 열린 파일 디스크립터가 정리되고
And HTTP 포트가 해제된다
```

## AC4: 윈도우별 파일 감시

### AC4.1: 해당 윈도우만 업데이트

```gherkin
Given file1.md와 file2.md가 각각 별도 윈도우에서 열려 있을 때
When file1.md 파일이 외부 편집기에서 수정되면
Then file1.md 윈도우의 콘텐츠만 자동으로 업데이트되고
And file2.md 윈도우는 변경되지 않는다
```

### AC4.2: 독립적 파일 감시

```gherkin
Given file1.md와 file2.md가 각각 별도 윈도우에서 열려 있을 때
When file1.md 윈도우를 닫으면
Then file2.md의 파일 감시는 계속 동작하고
And file2.md가 수정되면 정상적으로 업데이트된다
```

### AC4.3: 스크롤 위치 유지

```gherkin
Given file1.md 윈도우에서 문서 중간으로 스크롤한 상태에서
When file1.md 파일이 외부에서 수정되면
Then 콘텐츠가 업데이트된 후 스크롤 위치가 유지된다
```

## AC5: 시스템 트레이 통합

### AC5.1: 윈도우 목록 표시

```gherkin
Given file1.md와 file2.md가 각각 별도 윈도우에서 열려 있을 때
When 사용자가 시스템 트레이 아이콘을 우클릭하면
Then 메뉴에 "file1.md"와 "file2.md" 항목이 표시되고
And "종료" 메뉴 항목이 표시된다
```

### AC5.2: 윈도우 활성화

```gherkin
Given file1.md가 최소화되어 있고 시스템 트레이 메뉴가 표시될 때
When 사용자가 메뉴에서 "file1.md" 항목을 클릭하면
Then file1.md 윈도우가 복원되고 포그라운드로 활성화된다
```

### AC5.3: 동적 메뉴 갱신

```gherkin
Given 시스템 트레이 메뉴에 "file1.md"만 표시되고 있을 때
When 새 윈도우에서 file2.md가 열리면
Then 시스템 트레이 메뉴에 "file2.md" 항목이 추가된다
```

## AC6: Named Pipe 프로토콜

### AC6.1: 새 프로토콜 명령

```gherkin
Given 첫 번째 인스턴스의 Named Pipe 서버가 실행 중일 때
When "OPEN:C:\docs\readme.md" 메시지가 수신되면
Then "C:\docs\readme.md" 파일이 새 윈도우에서 열린다
```

### AC6.2: 하위 호환성

```gherkin
Given 첫 번째 인스턴스의 Named Pipe 서버가 실행 중일 때
When "C:\docs\readme.md" 메시지가 접두사 없이 수신되면
Then 하위 호환성을 위해 해당 파일이 새 윈도우에서 열린다
```

## AC7: 윈도우 제한 및 오류 처리

### AC7.1: 최대 윈도우 수 경고

```gherkin
Given 8개의 윈도우가 열려 있을 때
When 사용자가 새 파일을 열면
Then 새 윈도우가 정상적으로 생성되고
And "윈도우 제한에 가까워지고 있습니다 (8/10)" 경고가 표시된다
```

### AC7.2: 최대 윈도우 수 초과 거부

```gherkin
Given 10개의 윈도우가 이미 열려 있을 때
When 사용자가 새 파일을 열려고 하면
Then 새 윈도우가 생성되지 않고
And "최대 윈도우 수(10개)에 도달했습니다. 기존 윈도우를 닫은 후 다시 시도하세요." 오류 대화상자가 표시된다
```

### AC7.3: 파일 열기 실패 시 오류 처리

```gherkin
Given WinMarkdownViewer가 실행 중일 때
When 존재하지 않는 파일 경로로 열기를 요청하면
Then 빈 윈도우가 생성되지 않고
And "파일을 열 수 없습니다: [파일 경로]" 오류 대화상자가 표시된다
```

### AC7.4: 읽기 권한 없는 파일 열기

```gherkin
Given WinMarkdownViewer가 실행 중일 때
When 읽기 권한이 없는 파일을 열려고 하면
Then 빈 윈도우가 생성되지 않고
And "파일을 읽을 수 없습니다: 접근 권한을 확인하세요." 오류 대화상자가 표시된다
```

## 품질 게이트

| 항목 | 기준 |
|------|------|
| 테스트 커버리지 | 새 코드 85% 이상 |
| goroutine 누수 | goleak 검증 통과 |
| 경합 조건 | `go test -race` 통과 |
| 기존 테스트 | 모든 기존 테스트 통과 (회귀 없음) |
| 린트 | `golangci-lint run` 경고 0개 |
| 빌드 | `.\build.ps1 build` 성공, 콘솔 창 미표시 |

## 완료 정의 (Definition of Done)

- [ ] 모든 인수 기준(AC1~AC7)이 통과
- [ ] build.ps1이 프로젝트 루트에 존재하고 `.\build.ps1 build`가 콘솔 창 없는 바이너리를 생성
- [ ] Phase 0 PoC 완료: go-webview2 멀티 인스턴스 결과 경로(A/B/C) 확정
- [ ] WindowManager가 여러 윈도우를 독립적으로 관리
- [ ] 윈도우 닫기 시 관련 리소스만 정리 (다른 윈도우 영향 없음)
- [ ] 파일 변경 시 해당 윈도우만 업데이트
- [ ] 시스템 트레이에 열린 윈도우 목록 표시
- [ ] Named Pipe가 `OPEN:` 프로토콜을 지원하면서 하위 호환성 유지
- [ ] 새 코드 테스트 커버리지 85% 이상
- [ ] `go test -race ./...` 통과
- [ ] 기존 모든 테스트 통과 (회귀 없음)
