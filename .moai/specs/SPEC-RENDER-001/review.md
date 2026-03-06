---
spec_id: SPEC-RENDER-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-RENDER-001 리뷰

## 리뷰 요약

KaTeX + Mermaid 확장 렌더링 SPEC. JavaScript 후처리 수식 감지와 goldmark 코드 블록 활용 Mermaid 초기화 전략이 잘 설계됨. KaTeX 폰트 로딩의 HTTP 서버 의존성이 Critical.

---

## 1. spec.md 이슈

### 🔴 C-1: KaTeX 폰트 로딩과 HTTP 서버 의존성

- **위치**: spec.md §1.3, §3.4 REQ-RENDER-004
- **문제**: KaTeX CSS의 @font-face가 상대 경로로 폰트 로딩. SPEC-UI-001만(SetHtml 방식)으로는 origin이 about:blank이라 폰트 로딩 불가. SPEC-WATCH-001의 HTTP 서버가 사실상 필수
- **권장**: SPEC-WATCH-001을 필수 의존성으로 변경, 또는 KaTeX 폰트를 data: URI로 CSS에 인라인 포함

### 🟡 W-1: 바이너리 크기 증가 영향 평가
- Mermaid ~1.5MB + KaTeX ~300KB = ~1.8MB 증가. product.md의 "15MB 이하" 제약과의 총합 검증 필요

### 🟡 W-2: 수식 감지 정규식 Edge Case
- `$100 + $200` 같은 통화 표현, 홀수 개 달러 기호 등 edge case 처리 정책 필요

### 🟡 W-3: JS 테스트 프레임워크 미선택
- render-extensions_test.js가 명시되어 있으나 JS 테스트 도구가 기술 스택에 미포함

### 🟡 W-4: WebView2 CSP 리스크 대응 구체화 필요
- 인라인 스크립트 차단 리스크가 높음이지만 대응방안이 불충분

### 🟡 W-5: KaTeX/Mermaid 라이브러리 버전 고정 전략 부재
- "최신 안정 버전"이 아닌 구체적 버전 고정 필요

### 🟡 W-6: 실시간 미리보기 시 Mermaid 재렌더링 성능
- 다이어그램 많은 문서에서 "1초 이내 업데이트" 충족 어려울 수 있음

---

## 2. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🔴 Critical | 1건 | C-1 |
| 🟡 Warning | 6건 | W-1 ~ W-6 |

### 우선 조치 순서
1. KaTeX 폰트 로딩 전략 (C-1)
2. 바이너리 크기 사전 측정 (W-1)
3. CSP 사전 조사 (W-4)
4. JS 테스트 프레임워크 선정 (W-3)
