---
spec_id: SPEC-THEME-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-THEME-001 리뷰

## 리뷰 요약

다크 모드 테마 시스템 SPEC. CSS Custom Properties 기반 전환, 시스템 테마 감지, FOUC 방지 등 세부 기술이 잘 정의됨. SPEC-UI-001의 CSS 파괴적 변경과 Mermaid 테마 연동의 담당 SPEC 불명확이 주요 이슈.

---

## 1. spec.md 이슈

### 🟡 W-1: github-markdown.css 리팩터링의 파괴적 변경

- **위치**: spec.md §1.3, plan.md Task 3
- **문제**: SPEC-UI-001에서 작성한 github-markdown.css를 "CSS 변수 기반으로 리팩터링"하면 SPEC-UI-001의 ACC-001(GitHub 스타일 렌더링) 수용기준의 시각적 결과가 달라질 수 있음
- **권장**: SPEC-UI-001 구현 시 CSS 변수 전환을 고려한 구조 채택, 또는 SPEC-THEME-001에 시각적 회귀 테스트 필수화

### 🟡 W-2: Mermaid 테마 연동 담당 SPEC 불명확

- **위치**: spec.md REQ-THEME-009 (line 161), plan.md Task 17
- **문제**: Mermaid 다이어그램의 테마 연동이 SPEC-THEME-001에서 Optional Goal로, SPEC-RENDER-001에서도 "호환" 관계로만 언급. 어느 SPEC이 구현 책임인지 불명확
- **권장**: Mermaid 테마 연동의 구현 책임을 하나의 SPEC에 배정 (SPEC-THEME-001 권장)

### 🟡 W-3: localStorage fallback과 SPEC-WATCH-001의 관계

- **위치**: spec.md §3.8 REQ-THEME-008 (lines 152-159)
- **문제**: SPEC-CONFIG-001 미구현 시 localStorage를 fallback으로 사용. 그러나 SPEC-UI-001만 구현 시(SetHtml 방식) localStorage가 about:blank origin에서 동작하지 않을 수 있음. Navigate 방식(SPEC-WATCH-001)이면 localhost origin에서 동작 가능
- **권장**: localStorage fallback의 전제 조건으로 SPEC-WATCH-001(HTTP 서버)이 필요한지 명시

### 🟡 W-4: 구문 강조 CSS 생성 도구

- **위치**: plan.md §5.4 (lines 159-168)
- **문제**: chroma CLI로 CSS를 생성한다고 했으나, 이 CSS 생성이 빌드 타임에 수행되는지, 수동으로 한 번만 생성하여 커밋하는 것인지 불명확
- **권장**: "사전 생성하여 소스 코드에 커밋" 방식 명시. 또는 Go 코드에서 chroma API로 런타임 생성 검토

### 🟡 W-5: Ctrl+Shift+T 키보드 단축키 충돌 확인

- **위치**: spec.md §3.6 REQ-THEME-006 (lines 136-142)
- **문제**: Ctrl+Shift+T는 많은 브라우저에서 "닫은 탭 복원" 단축키로 사용됨. WebView2 내부에서 이 키 조합을 가로챌 수 있는지, WebView2 자체가 이 키를 소비하지 않는지 확인 필요
- **권장**: 대안 단축키(예: Ctrl+Shift+D) 검토 또는 WebView2의 키 이벤트 동작 테스트

### 🟡 W-6: JS 테스트 프레임워크 - SPEC-RENDER-001과 동일 이슈

- **위치**: spec.md §5 Traceability, plan.md §3
- **문제**: theme_test.js가 명시되어 있으나 JS 테스트 도구가 기술 스택에 미포함. SPEC-RENDER-001과 동일 이슈
- **권장**: 프로젝트 전체 JS 테스트 전략 수립 필요

---

## 2. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🟡 Warning | 6건 | W-1 ~ W-6 |

### 우선 조치 순서
1. github-markdown.css 리팩터링 전략 (W-1)
2. localStorage fallback 의존성 (W-3)
3. Ctrl+Shift+T 충돌 확인 (W-5)
4. Mermaid 테마 연동 담당 배정 (W-2)
