# WinMarkdownViewer SPEC 전체 검토 리포트

> **검토 일시**: 2026-03-06
> **검토 범위**: MASTER-PLAN.md + 7개 SPEC (SPEC-UI-001 ~ SPEC-THEME-001)
> **참조 문서**: docs/product.md, docs/structure.md, docs/tech.md

---

## 1. 전체 평가

### ✅ 잘 된 점

| 항목 | 설명 |
|------|------|
| **체계적 구성** | 모든 SPEC이 spec.md / plan.md / acceptance.md 3파일 구조로 일관됨 |
| **요구사항 분류** | Ubiquitous / Event-Driven / Unwanted / State-Driven / Optional 패턴이 일관 적용 |
| **Gherkin 스타일** | Acceptance Criteria가 Given-When-Then 형식으로 명확하게 작성됨 |
| **TDD 전략** | 모든 plan.md에 RED-GREEN-REFACTOR 사이클이 명시됨 |
| **Traceability** | 요구사항 → 구현 파일 → 테스트 시나리오 추적 매핑이 존재 |
| **의존성 그래프** | MASTER-PLAN에 SPEC 간 의존성이 명확히 정의됨 |
| **SPEC-UI-001 피드백 반영** | 이전 리뷰(review.md)의 주요 피드백이 spec.md에 상당 부분 반영됨 |
| **범위 관리** | 각 SPEC의 Out of Scope 섹션이 명확 |

### ⚠️ 공통 이슈

| 등급 | 건수 | 요약 |
|------|------|------|
| 🔴 Critical | 3건 | HKCU↔HKLM 불일치, CGO 제약과 systray 충돌, 바이너리 크기 초과 |
| 🟡 Warning | 12건 | 문서 간 정합성, 누락된 연동 사항, 테스트 전략 미비 |
| 🔵 Info | 5건 | 스타일/권장 사항 |

---

## 2. MASTER-PLAN 이슈

### 🟡 MP-1: 진행 추적표의 상태 업데이트 필요

- **위치**: MASTER-PLAN.md 진행 추적 테이블 (lines 117-125)
- **문제**: 모든 SPEC이 "📋 SPEC 완료 | 구현 대기"로 동일 상태. "권장 실행 순서" 섹션에서 "← 지금 여기"가 SPEC-UI-001을 가리키고 있는데, 실제 구현 상태와 동기화 필요
- **권장**: 구현 시작 시 상태를 자동 업데이트하는 규칙 명시 (예: 🚧 구현중, ✅ 구현완료)

### 🟡 MP-2: Phase별 병렬 가능 SPEC 명확화

- **문제**: Phase 2에서 SPEC-WATCH-001과 SPEC-WIN-001의 관계가 "SPEC-WATCH-001 권장" 수준인데, MASTER-PLAN의 실행 순서 제약에서는 "SPEC-WATCH-001 → SPEC-WIN-001보다 먼저"로 강한 제약으로 기술
- **권장**: "필수"와 "권장"의 구분을 실행 순서 제약 표에서도 일관되게 표기

---

## 3. 각 SPEC별 주요 이슈 (요약)

> 상세 리뷰는 각 SPEC 폴더의 `review.md`에 작성됨

### SPEC-UI-001
이전 리뷰 피드백 반영 확인. 대부분 반영됨 (W-1, W-2, W-5, W-7, W-8, W-9, W-10, W-11, X-1, X-2). **C-1(프로젝트 구조)과 C-2(모듈 경로)**는 반영 완료 상태.

### SPEC-WATCH-001
- 🟡 **gorilla/websocket의 아카이브 상태** 확인 필요 (deprecated 여부)
- 🟡 **SPEC-UI-001의 SetHtml→Navigate 전환**이 파괴적 변경인데, 하위호환성 전략 미비

### SPEC-WIN-001
- 🔴 **getlantern/systray가 CGO를 필요**로 할 수 있음 → "CGO 사용 금지" 제약과 충돌
- 🟡 **시스템 트레이와 WebView2 이벤트 루프 충돌** 가능성 높음

### SPEC-INSTALL-001
- 🔴 **HKCU vs HKLM 불일치**: SPEC-WIN-001은 HKCU 사용, SPEC-INSTALL-001은 HKLM 사용
- 🟡 **WiX v4 파일 구조**가 docs/structure.md의 파일명과 불일치

### SPEC-CONFIG-001
- 🟡 **SPEC-WIN-001 단일 인스턴스**와의 설정 파일 동시 접근 시나리오 미정의
- 🟡 **windowX/windowY 검증 범위**가 "화면 크기"로 되어 있으나, Go에서 화면 크기 조회 방법 미명시

### SPEC-RENDER-001
- 🔴 **바이너리 크기 증가 ~1.8MB**가 product.md의 "15MB 이하" 제약에 영향
- 🟡 **KaTeX 폰트의 go:embed 로딩**이 HTTP 서버 없이는 불가 (SPEC-WATCH-001 필수?)

### SPEC-THEME-001
- 🟡 **SPEC-UI-001의 github-markdown.css 리팩터링** 범위를 SPEC-UI-001 spec.md에 선반영 필요
- 🟡 **SPEC-RENDER-001의 Mermaid 테마 연동**이 양쪽 SPEC에서 모두 Optional로 정의, 담당 SPEC 불명확

---

## 4. 문서 간 교차 정합성 이슈 (Cross-SPEC)

### 🔴 XS-1: HKCU vs HKLM 레지스트리 범위 불일치

| SPEC | 레지스트리 범위 | 근거 |
|------|----------------|------|
| SPEC-WIN-001 | **HKCU** | REQ-U-001: "HKCU 범위에서만 수행", REQ-N-001: "HKLM 수정 금지" |
| SPEC-INSTALL-001 | **HKLM** | REQ-REG-001: `HKLM\SOFTWARE\Classes\.md\shell\WinMarkdownViewer` |

- **문제**: MSI 설치 프로그램(SPEC-INSTALL-001)은 관리자 권한으로 HKLM에 쓰고, 수동 등록(SPEC-WIN-001)은 HKCU에 쓴다. 이 두 가지가 공존하면 레지스트리 충돌 가능
- **권장**: MSI 설치 시 HKLM에 등록, `--register`는 MSI 비설치 환경(포터블)에서만 사용하는 것으로 역할 분리. 또는 MSI도 per-user 설치(HKCU)로 통일

### 🟡 XS-2: docs/structure.md vs SPEC 프로젝트 구조 불일치

| 항목 | docs/structure.md | SPEC 문서들 |
|------|------------------|-------------|
| 서버 모듈 | `internal/server/websocket.go` | `internal/server/handler.go` |
| 레지스트리 모듈 | `internal/registry/contextmenu.go`, `fileassoc.go` | `internal/registry/registry.go` |
| CSS 테마 | `web/css/dark.css`, `light.css` | `web/css/theme-dark.css`, `theme-light.css` |
| JS | `web/js/websocket.js`, `scroll.js` | viewer.html 내 인라인 JS |
| WiX 파일명 | `installer/wix/Product.wxs` | `installer/wix/Package.wxs` |
| App 모듈 | `internal/app/app.go` | 미정의 (SPEC-WIN-001에서 instance.go, pipe.go만 정의) |

- **권장**: docs/structure.md를 SPEC 문서들과 일치시키거나, docs/structure.md에 "이 문서는 전체 비전을 나타내며, 각 SPEC의 프로젝트 구조가 정확한 구현 명세"임을 명시

### 🟡 XS-3: docs/tech.md와 SPEC 기술 스택 불일치

| 항목 | docs/tech.md | SPEC 문서들 |
|------|-------------|-------------|
| Go 버전 | 1.26+ | 1.26.0 (일관됨) ✅ |
| 코드 하이라이트 | **highlight.js** (임베디드) | **goldmark-highlighting (chroma)** |
| WiX 버전 | 미명시 | v4 (SPEC-INSTALL-001) |

- **문제**: tech.md에 "highlight.js (임베디드)"로 되어 있으나, SPEC-UI-001에서는 "goldmark-highlighting (chroma 기반)"으로 구현. 완전히 다른 접근
- **권장**: docs/tech.md의 코드 하이라이트 항목을 goldmark-highlighting/chroma로 수정

### 🟡 XS-4: SPEC-RENDER-001의 SPEC-WATCH-001 의존성 강도

- **문제**: SPEC-RENDER-001은 SPEC-UI-001만 필수로 하고 SPEC-WATCH-001은 선택으로 정의. 그러나 KaTeX 폰트 파일(woff2)은 HTTP 서버 경유로 로딩해야 가능한데, SPEC-UI-001만으로는 `SetHtml()` 방식이므로 폰트 로딩 경로가 불명확
- **권장**: SPEC-RENDER-001의 KaTeX 폰트 로딩 전략을 재검토하고, SPEC-WATCH-001의 HTTP 서버가 실질적으로 필수인지 판단. 또는 KaTeX 폰트를 CSS `@font-face`의 `data:` URI로 인라인 임베딩하는 대안 검토

### 🟡 XS-5: SPEC-THEME-001이 SPEC-UI-001의 CSS를 파괴적 변경

- **문제**: SPEC-THEME-001이 github-markdown.css를 "CSS 변수 기반으로 리팩터링"하므로 SPEC-UI-001에서 작성한 CSS가 전면 교체됨. 이 변경이 SPEC-UI-001의 ACC-001(GitHub 스타일 렌더링) 수용 기준에 영향
- **권장**: SPEC-UI-001의 CSS 작성 시 향후 CSS 변수 전환을 고려한 구조로 작성하거나, SPEC-THEME-001 구현 시 시각적 회귀 테스트 수행 필수화

---

## 5. docs/ 문서 업데이트 권장 사항

| 파일 | 권장 변경 |
|------|----------|
| docs/tech.md | 코드 하이라이트를 "highlight.js" → "goldmark-highlighting (chroma)" 수정 |
| docs/structure.md | SPEC 문서의 프로젝트 구조와 일치시키거나, 비전 문서임을 명시 |
| docs/product.md | 비기능 요구사항 "바이너리 크기 15MB 이하"를 KaTeX+Mermaid 임베딩(~1.8MB) 반영하여 재확인 |

---

## 6. 우선 조치 순서

1. **HKCU vs HKLM 정합** (XS-1) — SPEC-WIN-001 / SPEC-INSTALL-001 간 레지스트리 정책 통일
2. **getlantern/systray CGO 의존성 확인** (SPEC-WIN-001 C-1) — CGO 제약과의 충돌 해결
3. **docs/tech.md 코드 하이라이트 수정** (XS-3) — 기술 스택 문서 정합
4. **KaTeX 폰트 로딩 전략** (XS-4) — HTTP 서버 의존성 명확화
5. **gorilla/websocket 상태 확인** (SPEC-WATCH-001) — 대안 검토
6. **나머지 Warning 항목 순차 처리**
