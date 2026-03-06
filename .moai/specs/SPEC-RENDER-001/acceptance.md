---
id: SPEC-RENDER-001
type: acceptance
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-RENDER-001: 수용 기준 - KaTeX + Mermaid 확장 렌더링

## 1. 수용 기준 시나리오

### 시나리오 1: 인라인 수학 수식 렌더링 [AC-RENDER-001]

```gherkin
Given 마크다운 파일에 "아인슈타인의 공식은 $E=mc^2$ 입니다" 텍스트가 포함되어 있다
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then "E=mc^2" 부분이 KaTeX로 렌더링된 수학 기호로 표시된다
And 수식 앞뒤의 일반 텍스트는 정상적으로 표시된다
And 수식은 인라인(텍스트 흐름에 포함)으로 표시된다
```

### 시나리오 2: 블록 수학 수식 렌더링 [AC-RENDER-002]

```gherkin
Given 마크다운 파일에 다음 블록 수식이 포함되어 있다:
  """
  $$\sum_{i=1}^{n} x_i = x_1 + x_2 + \cdots + x_n$$
  """
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 수식이 KaTeX로 렌더링된 수학 기호로 표시된다
And 수식은 별도의 줄에 중앙 정렬로 표시된다
And 시그마 기호, 상첨자, 하첨자가 올바르게 표시된다
```

### 시나리오 3: Mermaid flowchart 다이어그램 렌더링 [AC-RENDER-003]

```gherkin
Given 마크다운 파일에 다음 Mermaid 코드 블록이 포함되어 있다:
  """
  ```mermaid
  flowchart TD
    A[시작] --> B{조건}
    B -->|예| C[처리]
    B -->|아니오| D[종료]
  ```
  """
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then Mermaid 코드 블록이 시각적 흐름도로 렌더링된다
And "시작", "조건", "처리", "종료" 노드가 표시된다
And 화살표와 분기 조건 레이블이 표시된다
```

### 시나리오 4: Mermaid sequence 다이어그램 렌더링 [AC-RENDER-004]

```gherkin
Given 마크다운 파일에 다음 Mermaid sequence 다이어그램이 포함되어 있다:
  """
  ```mermaid
  sequenceDiagram
    Client->>Server: HTTP GET /api/users
    Server->>Database: SELECT * FROM users
    Database-->>Server: Result Set
    Server-->>Client: 200 OK (JSON)
  ```
  """
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then Mermaid 코드 블록이 시각적 시퀀스 다이어그램으로 렌더링된다
And Client, Server, Database 참여자가 표시된다
And 메시지 화살표와 레이블이 올바르게 표시된다
```

### 시나리오 5: 코드 블록 내 달러 기호 보호 [AC-RENDER-005]

```gherkin
Given 마크다운 파일에 다음 코드 블록이 포함되어 있다:
  """
  인라인 코드: `$variable = 100`

  코드 블록:
  ```bash
  echo "$HOME"
  price=$((100 * 2))
  ```
  """
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 코드 블록 내의 `$variable`, `$HOME`, `$((100 * 2))` 는 수식으로 처리되지 않는다
And 코드 블록의 텍스트가 원본 그대로 표시된다
```

### 시나리오 6: 수식 구문 오류 처리 [AC-RENDER-006]

```gherkin
Given 마크다운 파일에 잘못된 수식 "$\invalidcommand$"이 포함되어 있다
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 해당 수식은 빨간색 오류 스타일로 표시된다
And 오류 메시지가 tooltip으로 제공된다
And 문서의 나머지 콘텐츠는 정상적으로 렌더링된다
And 다른 올바른 수식은 정상적으로 KaTeX 렌더링된다
```

### 시나리오 7: Mermaid 구문 오류 처리 [AC-RENDER-007]

```gherkin
Given 마크다운 파일에 잘못된 Mermaid 구문이 포함되어 있다:
  """
  ```mermaid
  invalid diagram syntax >>>
  ```
  """
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 해당 다이어그램 위치에 오류 메시지가 표시된다
And 문서의 나머지 콘텐츠는 정상적으로 렌더링된다
And 다른 올바른 다이어그램은 정상적으로 Mermaid 렌더링된다
```

### 시나리오 8: 오프라인 환경 동작 [AC-RENDER-008]

```gherkin
Given 사용자의 컴퓨터가 네트워크에 연결되어 있지 않다
And 마크다운 파일에 수식과 Mermaid 다이어그램이 포함되어 있다
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 수식이 KaTeX로 정상 렌더링된다
And 다이어그램이 Mermaid로 정상 렌더링된다
And 외부 네트워크 요청이 발생하지 않는다
```

### 시나리오 9: 수식과 다이어그램 혼합 문서 [AC-RENDER-009]

```gherkin
Given 마크다운 파일에 일반 텍스트, 인라인 수식, 블록 수식, 코드 블록, Mermaid 다이어그램이 모두 포함되어 있다
When 사용자가 해당 파일을 WinMarkdownViewer로 연다
Then 각 요소가 해당 렌더러로 올바르게 렌더링된다
And 요소 간 레이아웃이 깨지지 않는다
And 문서 스크롤이 정상 동작한다
```

---

## 2. Edge Case

### 2.1 수식 관련
- 빈 수식: `$$` 또는 `$$$$` - 무시하거나 빈 블록 표시
- 여러 줄 블록 수식: 줄바꿈이 포함된 `$$...$$` 정상 처리
- 이스케이프된 달러 기호: `\$100` - 리터럴 달러 기호로 표시
- 통화 표현과의 충돌: `$100` - 인라인 수식은 `$` 뒤에 공백이 없는 경우만 감지. 또는 `\(...\)` 구문 우선 사용 권장
- 연속 수식: `$a$ 텍스트 $b$` - 각각 독립적으로 렌더링
- 중첩 달러 기호: `$$내부에 $인라인$이 있는 블록$$` - 블록 수식으로 처리

### 2.2 다이어그램 관련
- 매우 큰 다이어그램 (100+ 노드): 렌더링되되 성능 저하 허용
- 지원하지 않는 다이어그램 유형: Mermaid가 처리하므로 추가 검증 불필요
- 빈 mermaid 코드 블록: 오류 표시 또는 무시

### 2.3 성능 관련
- 50개 초과 다이어그램: 최대 50개만 렌더링하고 나머지는 코드 블록으로 표시
- 500자 초과 수식: 5초 타임아웃 후 원본 텍스트 표시
- 매우 큰 파일 (1MB+ 마크다운): 렌더링 지연 허용, 크래시 방지

---

## 3. Quality Gate

### 3.1 테스트 커버리지
- Go 코드 (renderer.go, embed.go, viewer.go 변경분): 85% 이상
- Go 테스트에서는 HTML 출력의 구조적 정확성만 검증
- JavaScript 코드 (render-extensions.js): JS 테스트는 수동 검증 또는 향후 Playwright 기반 E2E 테스트로 대체

### 3.2 성능 기준
- 수식 10개 + 다이어그램 5개 포함 문서: 초기 렌더링 3초 이내
- 추가 수식/다이어그램 없는 문서: 기존 대비 렌더링 시간 증가 500ms 미만

### 3.3 바이너리 크기
- KaTeX + Mermaid 임베딩 후 바이너리 크기 증가: 2MB 이내

### 3.4 TRUST 5 검증
- **Tested**: 모든 요구사항에 대한 테스트 존재
- **Readable**: 수식 감지 로직에 명확한 주석 포함
- **Unified**: 기존 코드 스타일과 일관성 유지
- **Secured**: 외부 네트워크 요청 없음, XSS 방지 (사용자 입력을 직접 innerHTML에 삽입하지 않음)
- **Trackable**: 커밋 메시지에 SPEC-RENDER-001 참조 포함

---

## 4. Definition of Done

- [ ] KaTeX 인라인 수식 ($...$) 렌더링 동작 확인
- [ ] KaTeX 블록 수식 ($$...$$) 렌더링 동작 확인
- [ ] Mermaid flowchart 다이어그램 렌더링 동작 확인
- [ ] Mermaid sequence 다이어그램 렌더링 동작 확인
- [ ] Mermaid class/state/gantt/pie 다이어그램 렌더링 동작 확인
- [ ] 코드 블록 내 달러 기호가 수식으로 처리되지 않음 확인
- [ ] 수식 구문 오류 시 오류 메시지 표시 및 문서 렌더링 유지 확인
- [ ] 다이어그램 구문 오류 시 오류 메시지 표시 및 문서 렌더링 유지 확인
- [ ] 오프라인 환경에서 정상 동작 확인
- [ ] 테스트 커버리지 85% 이상
- [ ] 바이너리 크기 증가 2MB 이내
