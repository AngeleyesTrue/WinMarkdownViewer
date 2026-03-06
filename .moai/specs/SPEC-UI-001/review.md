---
spec_id: SPEC-UI-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-UI-001 리뷰

## 리뷰 요약

전반적으로 잘 작성된 MVP 스펙 문서입니다. 요구사항 분류(Ubiquitous/Event-Driven/Unwanted/State-Driven/Optional), TDD 기반 구현 계획, Gherkin 스타일 수용 기준 등 체계적으로 구성되어 있습니다. 아래에 수정이 필요하거나 보완이 권장되는 사항들을 정리합니다.

---

## 1. spec.md 이슈

### 🔴 Critical (수정 필요)

#### C-1: 프로젝트 구조 불일치

- **위치**: spec.md §1.3 (lines 38-49)
- **문제**: `docs/structure.md`에 정의된 프로젝트 구조와 SPEC의 프로젝트 구조가 다름
  - `docs/structure.md`: `web/templates/viewer.html`, `web/css/github-markdown.css`, `web/css/dark.css`, `web/css/light.css`, `web/js/websocket.js`, `web/js/scroll.js`
  - SPEC: `web/template.html`, `web/style.css`
- **권장**: MVP 범위에서의 축소된 구조라면 그 근거를 명시하거나, 향후 확장을 고려하여 `docs/structure.md`의 경로 체계와 일치시킬 것. 최소한 네이밍 컨벤션은 통일 권장 (`template.html` vs `templates/viewer.html`)

#### C-2: 모듈 경로 확정 필요

- **위치**: spec.md §4.1 (line 100)
- **현재**: `github.com/user/WinMarkdownViewer` (또는 프로젝트에 맞는 경로)
- **문제**: 모듈 경로가 확정되지 않은 상태. 구현 시점에서 혼동 유발 가능.
- **권장**: 실제 사용할 모듈 경로를 확정하여 명시 (예: `github.com/yourorg/WinMarkdownViewer`)

### 🟡 Warning (보완 권장)

#### W-1: REQ-N-002 (임시 파일 미생성)와 WebView2 동작 충돌 가능성

- **위치**: spec.md §3.3 (line 80)
- **문제**: `go-webview2`는 내부적으로 WebView2 사용자 데이터 폴더(User Data Folder)를 생성할 수 있음. 이는 WebView2 Runtime의 동작이므로 애플리케이션 레벨에서 완전히 제어하기 어려움.
- **권장**: "시스템은 **자체적으로** 임시 파일을 파일시스템에 생성하지 않아야 한다 (WebView2 Runtime이 자체적으로 생성하는 사용자 데이터 폴더는 제외)"로 범위를 명확히 할 것

#### W-2: Traceability에서 REQ-N-002, REQ-N-003, REQ-O-001, REQ-O-002 누락

- **위치**: spec.md §6 (lines 148-159)
- **문제**: Traceability 테이블에 다음 요구사항에 대한 매핑이 빠져 있음:
  - `REQ-N-002` (임시 파일 미생성)
  - `REQ-N-003` (관리자 권한 미요구)
  - `REQ-O-001` (타이틀바 파일명 표시)
  - `REQ-O-002` (빈 파일 안내 메시지)
- **권장**: 모든 REQ에 대한 구현 파일 및 테스트 시나리오 매핑 추가

#### W-3: REQ-E-003의 에러 출력 위치 상세화

- **위치**: spec.md §3.2 (line 75)
- **문제**: "사용법 안내 메시지를 표준 에러에 출력"이라 되어 있으나, GUI 앱(`-H windowsgui` 링커 플래그 사용)인 경우 표준 에러가 콘솔에 표시되지 않음
- **권장**: 릴리스 빌드에서는 메시지 박스(MessageBox) 표시를 고려하거나, 개발/릴리스 빌드 시 동작을 구분하여 명시할 것

---

## 2. plan.md 이슈

### 🟡 Warning (보완 권장)

#### W-4: Task 의존성 그래프 불명확

- **위치**: plan.md §2 전체
- **문제**: Task 10(CLI 테스트)의 의존성이 `Task 6`으로 되어 있는데, Task 6은 Markdown 렌더러 테스트임. CLI 테스트가 렌더러 테스트에 의존하는 이유가 불분명. 아마 Task 1(Go 모듈 초기화)의 오타로 보임.
- **권장**: 의존성 관계 재검토 및 수정. Task 10은 `Task 1` 또는 `Task 7(렌더러 구현)`에 의존하는 것이 적절

#### W-5: Task 9 (WebView2 뷰어)에 테스트 전략 미비

- **위치**: plan.md §2 Task 9 (lines 82-90)
- **문제**: "WebView2는 GUI 컴포넌트로 단위 테스트 대신 빌드 검증 및 수동 테스트"로만 되어 있음. WebView2 래핑 로직 중 테스트 가능한 부분(옵션 검증, 에러 처리 등)에 대한 전략이 없음.
- **권장**: `viewer.go` 내에서 WebView2 초기화 로직을 인터페이스로 분리하여 mock 기반 유닛 테스트가 가능하도록 설계 제안 추가. 최소한 `New()` 함수의 파라미터 검증 테스트는 가능.

#### W-6: 파일 영향 분석에서 go.sum 누락

- **위치**: plan.md §3 (lines 115-128)
- **문제**: `go.sum`이 파일 목록에 없음. `go mod tidy` 실행 시 자동 생성되므로 의도적 누락일 수 있으나, 신규 프로젝트이므로 명시가 바람직.
- **권장**: `go.sum` (자동 생성)을 테이블에 추가하거나 비고로 언급

#### W-7: 후속 SPEC 넘버링 불일치

- **위치**: plan.md §6 (lines 200-209)
- **문제**: 후속 SPEC이 `SPEC-002`, `SPEC-003` 등으로 표기되어 있으나, 현재 SPEC ID 체계는 `SPEC-UI-001`로 도메인 접두사를 포함. 후속 SPEC도 동일 체계를 따라야 일관성 확보.
- **권장**: `SPEC-UI-002`, `SPEC-UI-003` 등 또는 도메인에 맞는 접두사 사용 (예: `SPEC-WATCH-001`, `SPEC-INSTALL-001`)

---

## 3. acceptance.md 이슈

### 🟡 Warning (보완 권장)

#### W-8: ACC-006에 두 개의 시나리오 합침

- **위치**: acceptance.md §2 ACC-006 (lines 80-94)
- **문제**: "파일 미존재"와 "읽기 권한 없음" 시나리오가 하나의 ACC-006에 합쳐져 있음. 각각 다른 REQ(REQ-S-001, REQ-S-002)에 대응하므로 분리가 바람직.
- **권장**: 읽기 권한 관련 시나리오를 별도 ACC-006-B 또는 ACC-008로 분리 (기존 ACC-008~011 번호 재조정 필요)

#### W-9: ACC-010 한글 경로 테스트 보완 필요

- **위치**: acceptance.md §3 ACC-010 (lines 130-136)
- **문제**: "내 문서/readme.md"로 한글 경로만 예시. Windows 환경에서 실제 많이 사용되는 케이스들이 추가되면 좋음:
  - 긴 경로 (260자 초과)
  - UNC 경로 (`\\server\share\file.md`)
  - 드라이브 문자 포함 절대 경로 (`C:\Users\사용자\문서\readme.md`)
- **권장**: Windows 특화 경로 케이스를 별도 수용 기준으로 추가하거나, ACC-010의 예시를 확장

#### W-10: ACC-011 (비 .md 확장자) 정책 검토 필요

- **위치**: acceptance.md §3 ACC-011 (lines 138-145)
- **문제**: "확장자에 관계없이 내용을 Markdown으로 처리한다"는 정책이 의도적인 것인지 확인 필요. 컨텍스트 메뉴가 `.md` 파일에만 등록되지만, 사용자가 CLI로 임의 파일을 전달할 수 있음. `.txt` 파일을 Markdown으로 처리하는 것은 합리적이나, 바이너리 파일 등이 전달되었을 때의 동작도 정의 필요.
- **권장**: 바이너리 파일이나 인코딩이 맞지 않는 파일에 대한 엣지 케이스도 추가 (예: ACC-0XX: 바이너리 파일 전달 시 에러 메시지 또는 raw text 표시)

#### W-11: 테스트 커버리지 85% 기준의 범위 조정

- **위치**: acceptance.md §4 ACC-013 (line 170), §5 Quality Gate (line 189)
- **문제**: ACC-013에서는 "테스트 커버리지가 85% 이상이다"라고 했으나, Quality Gate에서는 "(GUI 제외)"라고 부연. MVP는 파일 수가 적고 viewer.go가 상당 부분을 차지하므로, 85%가 달성 가능한지 사전 검토 필요.
- **권장**: 커버리지 측정 범위를 `internal/markdown/` 패키지로 한정하거나, `internal/viewer/`를 명시적으로 제외하는 것이 현실적

---

## 4. 문서 간 교차 정합성 이슈

### 🟡 Warning (보완 권장)

#### X-1: 제품 정의서(docs/product.md)와 SPEC 범위 확인

- **문제**: `docs/product.md`의 F1에 포함된 **KaTeX 수학 수식**, **Mermaid 다이어그램** 지원이 SPEC-UI-001에서는 명시적으로 "범위 외"로 분류됨 (plan.md §6). 이는 적절한 판단이나, `spec.md`의 Requirements에서 이것이 범위 밖임을 명시적으로 언급하면 더 명확해짐.
- **권장**: spec.md §5 Constraints에 "KaTeX/Mermaid 렌더링은 이 SPEC 범위에 포함되지 않음"을 한 줄 추가

#### X-2: 다크/라이트 테마 관련 범위

- **문제**: `docs/product.md` F5에 "테마 선택 (라이트/다크/시스템 연동)"이 핵심 기능으로 있고, `docs/structure.md`에는 `web/css/dark.css`, `web/css/light.css`가 정의되어 있으나, SPEC-UI-001의 `web/style.css` 하나로 단순화됨.
- **권장**: MVP의 CSS가 라이트 테마만 지원하는지, 시스템 테마 연동(`prefers-color-scheme`) 기본 지원이 포함되는지 명시. 최소한 Requirements에 테마 관련 범위를 명시적으로 기술

#### X-3: docs/tech.md Go 버전 업데이트 필요

- **문제**: `docs/tech.md`에 Go 버전이 `1.23+`으로 되어 있으나, SPEC-UI-001에서는 `Go 1.26.0`을 명시. SPEC이 최신 기준이므로 `docs/tech.md`의 버전 표기가 오래된 상태.
- **권장**: `docs/tech.md`의 Go 버전을 `1.26+` 또는 `1.26.0`으로 업데이트하여 SPEC과 일치시킬 것

---

## 5. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|----|
| 🔴 Critical | 2건 | C-1, C-2 |
| 🟡 Warning | 11건 | W-1 ~ W-7, W-8 ~ W-11 |
| 교차 정합성 | 3건 | X-1, X-2, X-3 |

### 우선 조치 권장 순서

1. **프로젝트 구조 정합** (C-1) — `docs/structure.md`와의 관계 정리
2. **모듈 경로 확정** (C-2) — 구현 착수 전 반드시 결정
3. **Traceability 보완** (W-2) — 누락된 REQ 매핑 추가
4. **후속 SPEC 넘버링** (W-7) — ID 체계 결정
5. **docs/tech.md Go 버전 업데이트** (X-3) — SPEC과 일치시키기
6. 나머지 Warning 항목 순차 처리
