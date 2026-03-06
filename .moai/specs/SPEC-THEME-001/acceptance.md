---
id: SPEC-THEME-001
type: acceptance
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-THEME-001: 수용 기준 - 다크 모드 테마 시스템

## 1. 수용 기준 시나리오

### 시나리오 1: 라이트 테마 기본 표시 [AC-THEME-001]

```gherkin
Given 테마 설정이 "light"로 저장되어 있다
When 사용자가 마크다운 파일을 WinMarkdownViewer로 연다
Then 문서 배경이 흰색 계열(#ffffff)로 표시된다
And 텍스트가 어두운 색(#1f2328)으로 표시된다
And 코드 블록 배경이 밝은 회색(#f6f8fa)으로 표시된다
And 코드 구문 강조가 github 라이트 스타일로 표시된다
```

### 시나리오 2: 다크 테마 표시 [AC-THEME-002]

```gherkin
Given 테마 설정이 "dark"로 저장되어 있다
When 사용자가 마크다운 파일을 WinMarkdownViewer로 연다
Then 문서 배경이 어두운 색(#0d1117)으로 표시된다
And 텍스트가 밝은 색(#e6edf3)으로 표시된다
And 코드 블록 배경이 어두운 색(#161b22)으로 표시된다
And 코드 구문 강조가 github-dark 스타일로 표시된다
```

### 시나리오 3: 시스템 테마 자동 감지 - 다크 모드 [AC-THEME-003]

```gherkin
Given 테마 설정이 "system"으로 설정되어 있다
And Windows 시스템 설정이 다크 모드로 되어 있다
When 사용자가 마크다운 파일을 WinMarkdownViewer로 연다
Then 뷰어가 자동으로 다크 테마를 적용한다
And 문서 배경이 어두운 색으로 표시된다
And 텍스트가 밝은 색으로 표시된다
```

### 시나리오 4: 시스템 테마 실시간 전환 [AC-THEME-004]

```gherkin
Given 테마 설정이 "system"으로 설정되어 있다
And 뷰어에 마크다운 파일이 열려 있다
And 현재 라이트 테마가 적용되어 있다
When Windows 시스템 설정을 다크 모드로 변경한다
Then 뷰어가 재시작 없이 다크 테마로 전환된다
And 전환 시 부드러운 애니메이션이 적용된다
And 문서 콘텐츠가 유지된다
```

### 시나리오 5: 키보드 단축키로 테마 전환 [AC-THEME-005]

```gherkin
Given 뷰어에 마크다운 파일이 열려 있다
And 현재 테마 모드가 "system"이다
When 사용자가 Ctrl+Shift+D를 누른다
Then 테마 모드가 "light"로 변경된다
And 라이트 테마가 적용된다

When 사용자가 다시 Ctrl+Shift+D를 누른다
Then 테마 모드가 "dark"로 변경된다
And 다크 테마가 적용된다

When 사용자가 다시 Ctrl+Shift+D를 누른다
Then 테마 모드가 "system"으로 변경된다
And 시스템 테마에 따른 테마가 적용된다
```

### 시나리오 6: 테마 설정 영구 저장 [AC-THEME-006]

```gherkin
Given 뷰어에 마크다운 파일이 열려 있다
When 사용자가 테마를 "dark"로 변경한다
And 뷰어를 종료한다
And 다시 마크다운 파일을 연다
Then 다크 테마가 자동으로 적용된다
And 테마 전환 없이 처음부터 다크 테마로 표시된다
```

### 시나리오 7: 초기 로딩 시 깜빡임 없음 (FOUC 방지) [AC-THEME-007]

```gherkin
Given 테마 설정이 "dark"로 저장되어 있다
When 사용자가 마크다운 파일을 WinMarkdownViewer로 연다
Then 흰색 배경이 잠깐 보이는 깜빡임이 발생하지 않는다
And 처음부터 다크 테마로 렌더링된다
```

### 시나리오 8: 코드 블록 구문 강조 테마 연동 [AC-THEME-008]

```gherkin
Given 마크다운 파일에 Go 코드 블록이 포함되어 있다
And 현재 라이트 테마가 적용되어 있다
When 사용자가 테마를 다크 모드로 전환한다
Then 코드 블록의 구문 강조 색상이 다크 테마에 맞게 변경된다
And 키워드, 문자열, 주석 등의 색상이 다크 배경에 맞는 밝은 색으로 변경된다
And 코드 블록 배경도 어두운 색으로 변경된다
```

### 시나리오 9: 테마 전환 트랜지션 [AC-THEME-009]

```gherkin
Given 뷰어에 마크다운 파일이 열려 있다
And 현재 라이트 테마가 적용되어 있다
When 사용자가 테마를 다크 모드로 전환한다
Then 배경색이 0.2초~0.3초에 걸쳐 부드럽게 변경된다
And 텍스트 색상이 부드럽게 전환된다
And 급격한 색상 변화(깜빡임)가 발생하지 않는다
```

### 시나리오 10: UI 토글 버튼 동작 [AC-THEME-010]

```gherkin
Given 뷰어에 마크다운 파일이 열려 있다
When 사용자가 뷰어 오른쪽 상단의 테마 토글 버튼을 확인한다
Then 현재 테마 상태를 나타내는 아이콘이 표시된다

When 사용자가 토글 버튼을 클릭한다
Then 테마가 다음 모드로 순환 전환된다
And 아이콘이 새로운 테마 상태를 반영한다
```

---

## 2. Edge Case

### 2.1 테마 전환 관련
- 빠른 연속 전환: Ctrl+Shift+D를 빠르게 여러 번 누를 때 마지막 상태가 적용됨
- 전환 중 페이지 새로고침: 전환 애니메이션이 중단되어도 최종 상태가 올바르게 적용
- config.json과 localStorage가 다른 값을 가질 때: config.json 우선

### 2.2 시스템 연동 관련
- Windows 고대비 모드: CSS `prefers-contrast` 감지는 범위 외이나, 크래시 없이 동작
- 멀티 모니터에서 각 모니터의 DPI/색상 프로파일이 다를 때: WebView2 기본 처리에 위임
- WebView2가 `prefers-color-scheme`을 지원하지 않는 매우 오래된 버전: 라이트 테마 fallback

### 2.3 저장 관련
- config.json 파일이 없을 때: localStorage fallback 사용 (SPEC-WATCH-001의 HTTP 서버가 전제 조건)
- config.json 파일이 읽기 전용일 때: localStorage fallback 사용, 사용자에게 알림 없음 (SPEC-WATCH-001 필요)
- localStorage가 비활성화된 환경 (SPEC-UI-001만, about:blank origin): 매 실행 시 system 모드로 시작 (비저장)

### 2.4 콘텐츠 관련
- 이미지가 포함된 문서: 이미지 자체는 테마 영향 없음 (CSS filter 적용 안 함)
- 인라인 HTML에 하드코딩된 색상: 테마 변경에 반응하지 않음 (허용)
- 매우 긴 문서 (10,000줄+): 테마 전환 시 성능 저하 없이 동작

---

## 3. Quality Gate

### 3.1 테스트 커버리지
- Go 코드 (viewer.go, config.go 변경분): 85% 이상
- JavaScript 코드 (theme.js): JS 테스트는 수동 검증 또는 향후 Playwright 기반 E2E 테스트로 대체
  - 테마 전환 로직
  - 시스템 테마 감지
  - 키보드 단축키 처리
  - 설정 저장/로딩

### 3.2 시각적 품질
- 라이트/다크 테마 모두에서 모든 UI 요소가 읽기 가능해야 한다
- 테마 전환 시 텍스트 대비비(contrast ratio): WCAG AA 기준 (4.5:1) 이상
- 코드 블록 구문 강조가 배경과 충분한 대비를 가져야 한다

### 3.3 성능 기준
- 테마 전환 시간: CSS 트랜지션 포함 500ms 이내
- 초기 로딩 시 FOUC 없음
- 테마 전환이 스크롤 위치에 영향을 주지 않음

### 3.4 TRUST 5 검증
- **Tested**: 모든 테마 모드(system/light/dark)에 대한 테스트 존재
- **Readable**: CSS 변수명이 역할을 명확히 설명
- **Unified**: github-markdown 스타일과 일관된 색상 체계
- **Secured**: XSS 벡터 없음 (CSS만 변경, innerHTML 미사용)
- **Trackable**: 커밋 메시지에 SPEC-THEME-001 참조 포함

---

## 4. Definition of Done

- [ ] 라이트 테마 CSS 변수 정의 및 적용 확인
- [ ] 다크 테마 CSS 변수 정의 및 적용 확인
- [ ] github-markdown.css가 CSS 변수 기반으로 리팩터링됨 (라이트 테마의 시각적 결과가 기존과 동일해야 함 - 시각적 회귀 테스트)
- [ ] 시스템 테마 자동 감지 동작 확인 (Windows 다크 모드)
- [ ] 시스템 테마 변경 시 실시간 반영 동작 확인
- [ ] Ctrl+Shift+D 키보드 단축키로 테마 순환 전환 동작 확인
- [ ] UI 토글 버튼 표시 및 클릭 전환 동작 확인
- [ ] 테마 설정 저장 및 재시작 후 복원 동작 확인
- [ ] 코드 블록 구문 강조 테마 연동 동작 확인
- [ ] 테마 전환 시 부드러운 트랜지션 애니메이션 동작 확인
- [ ] 초기 로딩 시 FOUC(깜빡임) 없음 확인
- [ ] 테스트 커버리지 85% 이상
