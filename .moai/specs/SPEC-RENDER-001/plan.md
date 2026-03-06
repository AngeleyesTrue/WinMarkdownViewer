---
id: SPEC-RENDER-001
type: plan
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
methodology: TDD
---

# SPEC-RENDER-001: 구현 계획 - KaTeX + Mermaid 확장 렌더링

## 1. TDD 접근 방식

본 SPEC은 TDD (RED-GREEN-REFACTOR) 방식으로 구현한다.

### 1.1 RED-GREEN-REFACTOR 전략

- **RED**: 각 요구사항에 대한 실패하는 테스트를 먼저 작성
- **GREEN**: 테스트를 통과하는 최소한의 코드 구현
- **REFACTOR**: 중복 제거 및 코드 품질 개선

---

## 2. 마일스톤 (우선순위 기반)

### Primary Goal: KaTeX 수식 렌더링

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 1 | KaTeX 라이브러리 파일 다운로드 및 배치 | web/js/katex.min.js, web/css/katex.min.css, web/fonts/ | 파일 존재 확인 |
| 2 | go:embed 선언 업데이트 | web/embed.go | embed.FS 접근 테스트 |
| 3 | HTML 템플릿에 KaTeX CSS/JS 로딩 추가 | web/templates/viewer.html | 템플릿 렌더링 테스트 |
| 4 | render-extensions.js 수식 감지 로직 구현 | web/js/render-extensions.js | 인라인/블록 수식 감지 테스트 |
| 5 | KaTeX 오류 처리 구현 | web/js/render-extensions.js | 잘못된 수식 오류 표시 테스트 |

### Secondary Goal: Mermaid 다이어그램 렌더링

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 6 | Mermaid 라이브러리 파일 다운로드 및 배치 | web/js/mermaid.min.js | 파일 존재 확인 |
| 7 | go:embed 선언에 Mermaid 추가 | web/embed.go | embed.FS 접근 테스트 |
| 8 | goldmark mermaid 코드 블록 출력 확인/설정 | internal/markdown/renderer.go | mermaid 블록 HTML 출력 테스트 |
| 9 | render-extensions.js Mermaid 초기화 로직 구현 | web/js/render-extensions.js | Mermaid 블록 감지 및 변환 테스트 |
| 10 | Mermaid 오류 처리 구현 | web/js/render-extensions.js | 잘못된 구문 오류 표시 테스트 |

### Final Goal: 통합 및 호환성

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 11 | 수식과 다이어그램이 동시에 포함된 문서 통합 테스트 | web/js/render-extensions.js | 복합 문서 렌더링 테스트 |
| 12 | 성능 보호 (다이어그램 최대 50개, 수식 타임아웃) | web/js/render-extensions.js | 대량 요소 제한 테스트 |
| 13 | 실시간 미리보기 재렌더링 지원 | web/js/render-extensions.js | 재렌더링 함수 호출 테스트 |

---

## 3. 파일 영향 분석

### 신규 파일

| 파일 경로 | 목적 |
|-----------|------|
| web/js/katex.min.js | KaTeX 렌더링 엔진 (외부 라이브러리) |
| web/js/mermaid.min.js | Mermaid 렌더링 엔진 (외부 라이브러리) |
| web/js/render-extensions.js | 수식/다이어그램 감지 및 렌더링 초기화 |
| web/css/katex.min.css | KaTeX 스타일시트 (외부 라이브러리) |
| web/fonts/KaTeX_*.woff2 | KaTeX 수학 폰트 파일 |

### 수정 파일

| 파일 경로 | 변경 내용 |
|-----------|-----------|
| web/embed.go | KaTeX/Mermaid 리소스에 대한 go:embed 선언 추가 |
| web/templates/viewer.html | KaTeX CSS/JS, Mermaid JS, render-extensions.js 로딩 추가 |
| internal/markdown/renderer.go | goldmark mermaid 코드 블록 class 설정 확인/조정 |

### 테스트 파일

| 파일 경로 | 테스트 범위 |
|-----------|-------------|
| internal/markdown/renderer_test.go | mermaid 코드 블록 HTML 출력 검증 |
| internal/viewer/viewer_test.go | embed.FS 리소스 접근 검증 |
| web/js/render-extensions_test.js | 수식/다이어그램 감지, 렌더링, 오류 처리 검증 |

---

## 4. 의존성

### 외부 라이브러리 (go:embed로 임베딩)
- **KaTeX**: v0.16.x (major 버전 고정, go:embed에 포함. 버전 업데이트는 별도 PR로 관리)
  - katex.min.js
  - katex.min.css
  - KaTeX 폰트 파일 (woff2)
- **Mermaid**: v10.x (major 버전 고정, go:embed에 포함. 버전 업데이트는 별도 PR로 관리)
  - mermaid.min.js

### SPEC 의존성
- **필수**: SPEC-UI-001 (기본 뷰어 MVP) - goldmark 렌더러와 WebView2 윈도우 기반
- **필수**: SPEC-WATCH-001 (HTTP 서버) - KaTeX CSS의 @font-face가 상대 경로로 폰트를 로딩하므로 HTTP 서버가 필수. SetHtml 방식(about:blank origin)에서는 폰트 로딩 불가

---

## 5. 기술적 접근 방식

### 5.1 수식 감지 전략: JavaScript 후처리

goldmark 레벨에서 수식을 감지하는 대신, JavaScript에서 렌더링된 HTML의 텍스트 노드를 순회하여 `$...$` / `$$...$$` 패턴을 감지한다.

**장점:**
- goldmark 커스텀 확장 개발 불필요 (구현 복잡도 감소)
- KaTeX API와 직접 통합 가능
- 향후 다른 수식 라이브러리로 교체 용이

**주의점:**
- `<code>`, `<pre>` 태그 내부의 `$`는 반드시 건너뛰어야 한다
- 정규식 기반 감지 시 edge case 처리 필요 (이스케이프, 중첩 등)

### 5.2 Mermaid 통합 전략: goldmark 코드 블록 + JS 초기화

goldmark이 \`\`\`mermaid 코드 블록을 `<pre><code class="language-mermaid">` 태그로 출력하면, JavaScript에서 이를 `<div class="mermaid">` 태그로 변환하고 mermaid.run()을 호출한다.

**장점:**
- goldmark의 기본 코드 블록 처리 재활용
- Mermaid의 표준 초기화 패턴 활용
- 다이어그램 유형별 별도 처리 불필요

### 5.3 리소스 로딩 순서

```
1. <head>에서 katex.min.css 로딩 (렌더링 차단, FOUC 방지)
2. 마크다운 HTML 콘텐츠 렌더링
3. katex.min.js 로딩 (defer)
4. mermaid.min.js 로딩 (defer)
5. render-extensions.js 로딩 (defer, 의존성 파일 뒤에 배치)
6. DOMContentLoaded 이벤트에서 수식 감지 + Mermaid 초기화 실행
```

---

## 6. 리스크 및 대응 방안

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| KaTeX 폰트 파일이 embed에서 로딩되지 않음 | 높음 | KaTeX CSS의 @font-face가 상대 경로로 폰트를 로딩하므로 HTTP 서버(SPEC-WATCH-001)가 필수. SetHtml 방식에서는 폰트 로딩 불가 |
| Mermaid.js 크기(~1.5MB)로 바이너리 크기 증가 | 중간 | 수용 가능한 수준이지만, 필요 시 Mermaid lite 버전 검토 |
| 수식 감지 정규식이 edge case를 놓침 | 중간 | 테스트 케이스를 충분히 확보하고 반복 개선 |
| WebView2의 CSP(Content Security Policy)가 인라인 스크립트 차단 | 높음 | webview2.Settings에서 CSP를 완화하거나 외부 JS 파일로 분리하여 대응 |
| 실시간 미리보기 시 Mermaid 재렌더링 깜빡임 | 낮음 | 다이어그램 10개 이상 문서에서는 변경된 Mermaid 블록만 선택적으로 재렌더링. 전체 재렌더링은 최후 수단 |
