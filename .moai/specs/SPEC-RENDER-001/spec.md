---
id: SPEC-RENDER-001
title: "Extended Rendering - KaTeX + Mermaid"
version: 1.0.0
status: completed
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P3
lifecycle: spec-first
tags: [katex, mermaid, math, diagram, rendering, go-embed]
---

# SPEC-RENDER-001: 확장 렌더링 - KaTeX 수학 수식 + Mermaid 다이어그램

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- Microsoft Edge WebView2 Runtime (Evergreen 배포)

### 1.2 기술 스택
- Go 1.26.0 (CGO 불필요)
- github.com/yuin/goldmark - GFM Markdown 파서
- KaTeX (JavaScript 라이브러리) - 수학 수식 렌더링
- Mermaid (JavaScript 라이브러리) - 다이어그램 렌더링
- go:embed - KaTeX/Mermaid JS/CSS 리소스 임베딩

### 1.3 리소스 크기 예상
- KaTeX: ~300KB (katex.min.js + katex.min.css + 폰트 파일)
- Mermaid: ~1.5MB (mermaid.min.js, minified)
- 총 바이너리 크기 증가: ~1.8MB
- 참고: product.md의 15MB 이하 제약 내에서 충분한 여유 있음 (총 예상 ~5-8MB)

### 1.4 프로젝트 구조 (변경 사항)
```
WinMarkdownViewer/
  internal/markdown/renderer.go      # goldmark 확장 설정 추가
  web/
    templates/viewer.html            # KaTeX/Mermaid 스크립트 로딩 추가
    js/
      katex.min.js                   # KaTeX 렌더링 엔진 (go:embed)
      mermaid.min.js                 # Mermaid 렌더링 엔진 (go:embed)
      render-extensions.js           # 수식/다이어그램 감지 및 렌더링 스크립트
    css/
      katex.min.css                  # KaTeX 스타일 (go:embed)
    fonts/
      KaTeX_*.woff2                  # KaTeX 수학 폰트 (go:embed)
    embed.go                         # go:embed 선언 업데이트
```

---

## 2. Assumptions (가정)

### 2.1 전제 조건
- SPEC-UI-001이 완료되어 기본 마크다운 렌더링과 WebView2 표시가 동작한다
- goldmark 파서가 커스텀 확장 또는 HTML 후처리를 지원한다
- WebView2가 KaTeX/Mermaid JavaScript를 정상 실행한다

### 2.2 설계 결정
- 외부 CDN을 사용하지 않고 모든 JS/CSS/폰트를 go:embed로 바이너리에 임베딩한다
- 수식 감지는 goldmark 레벨이 아닌 JavaScript 후처리 방식으로 구현한다
  - 이유: goldmark에 수식 전용 확장을 만드는 것보다 JS 기반 감지가 구현 복잡도가 낮다
  - 이유: KaTeX 자체가 DOM 기반 렌더링이므로 JS 레벨에서 통합하는 것이 자연스럽다
- Mermaid 블록은 goldmark이 \`\`\`mermaid 코드 블록을 \<pre class="language-mermaid"\> 태그로 출력하면, JS에서 mermaid.init()으로 렌더링한다

### 2.3 범위 외 (Out of Scope)
- 수식/다이어그램 편집 기능 (뷰어 전용)
- 수식/다이어그램 이미지 내보내기
- PlantUML 지원
- LaTeX 전체 문법 지원 (KaTeX 지원 범위만)

---

## 3. Requirements (요구사항)

### 3.1 인라인 수학 수식 렌더링 [REQ-RENDER-001]

**WHEN** 마크다운 문서에 `$E=mc^2$` 형태의 인라인 수식이 포함되어 있을 때
**THEN** 시스템은 해당 텍스트를 KaTeX로 렌더링하여 수학 기호로 표시해야 한다

- 인라인 수식 구분자: `$...$` (단일 달러 기호)
- 통화 표현($100 등)과의 충돌을 방지하기 위해 인라인 수식은 `$` 뒤에 공백이 없는 경우만 감지. 또는 `\(...\)` 구문을 우선 사용 권장
- 코드 블록(`\`...\``, `\`\`\`...\`\`\``) 내부의 `$`는 수식으로 처리하지 않는다
- 이스케이프된 `\$`는 리터럴 달러 기호로 표시한다

### 3.2 블록 수학 수식 렌더링 [REQ-RENDER-002]

**WHEN** 마크다운 문서에 `$$\sum_{i=1}^{n} x_i$$` 형태의 블록 수식이 포함되어 있을 때
**THEN** 시스템은 해당 텍스트를 KaTeX로 렌더링하여 중앙 정렬된 수학 수식 블록으로 표시해야 한다

- 블록 수식 구분자: `$$...$$` (이중 달러 기호)
- 블록 수식은 별도의 줄에 중앙 정렬로 표시한다
- 여러 줄에 걸친 수식을 지원한다

### 3.3 Mermaid 다이어그램 렌더링 [REQ-RENDER-003]

**WHEN** 마크다운 문서에 \`\`\`mermaid 코드 블록이 포함되어 있을 때
**THEN** 시스템은 해당 블록을 Mermaid.js로 렌더링하여 시각적 다이어그램으로 표시해야 한다

지원 다이어그램 유형:
- flowchart (흐름도)
- sequence (시퀀스 다이어그램)
- class (클래스 다이어그램)
- state (상태 다이어그램)
- gantt (간트 차트)
- pie (원형 차트)

### 3.4 로컬 리소스 임베딩 [REQ-RENDER-004]

시스템은 **항상** KaTeX 및 Mermaid의 모든 JS/CSS/폰트 파일을 go:embed로 바이너리에 임베딩해야 한다

- 외부 CDN 또는 네트워크 요청이 발생**하지 않아야 한다**
- 오프라인 환경에서도 수식과 다이어그램이 정상 렌더링되어야 한다

### 3.5 수식 렌더링 오류 처리 [REQ-RENDER-005]

**IF** KaTeX가 수식 구문을 파싱할 수 없는 경우
**THEN** 시스템은 원본 수식 텍스트를 빨간색 오류 스타일로 표시하고, 렌더링 오류 메시지를 tooltip으로 제공해야 한다

- 하나의 수식 오류가 다른 수식이나 문서 전체 렌더링에 영향을 주**지 않아야 한다**

### 3.6 다이어그램 렌더링 오류 처리 [REQ-RENDER-006]

**IF** Mermaid가 다이어그램 구문을 파싱할 수 없는 경우
**THEN** 시스템은 원본 코드 블록 텍스트를 오류 스타일로 표시하고, 파싱 오류 메시지를 표시해야 한다

- 하나의 다이어그램 오류가 다른 다이어그램이나 문서 전체 렌더링에 영향을 주**지 않아야 한다**

### 3.7 HTML 템플릿 통합 [REQ-RENDER-007]

**WHEN** 마크다운 파일이 열릴 때
**THEN** 시스템은 HTML 템플릿에 KaTeX CSS/JS 및 Mermaid JS를 로딩하여 수식과 다이어그램 렌더링을 준비해야 한다

- KaTeX CSS는 \<head\>에서 로딩한다 (FOUC 방지)
- KaTeX JS, Mermaid JS는 콘텐츠 렌더링 후 로딩한다 (defer/async)
- render-extensions.js가 DOM 로딩 완료 후 수식 감지 및 다이어그램 초기화를 실행한다

### 3.8 CSP 대응 [REQ-RENDER-008-CSP]

**IF** WebView2 CSP 설정이 인라인 스크립트를 차단하는 경우
**THEN** 시스템은 webview2.Settings에서 CSP를 완화하거나 외부 JS 파일로 분리하여 대응해야 한다

### 3.9 버전 고정 [REQ-RENDER-008-VER]

시스템은 **항상** KaTeX v0.16.x, Mermaid v10.x 등 major 버전을 고정하여 go:embed에 포함해야 한다
- 버전 업데이트는 별도 PR로 관리한다

### 3.10 실시간 미리보기 호환 [REQ-RENDER-008]

**IF** SPEC-WATCH-001이 구현되어 실시간 미리보기가 활성화된 상태에서
**WHEN** 마크다운 파일이 수정되어 WebSocket으로 업데이트가 전달될 때
**THEN** 시스템은 업데이트된 HTML에 대해 KaTeX 및 Mermaid 재렌더링을 실행해야 한다

---

## 4. Specifications (사양)

### 4.1 KaTeX 수식 감지 알고리즘

```
1. DOM 로딩 완료 후 render-extensions.js 실행
2. 텍스트 노드를 순회하며 $...$ 및 $$...$$ 패턴 감지
3. 감지된 패턴에 대해:
   - 코드 블록(<code>, <pre>) 내부인지 확인 → 내부이면 건너뛴다
   - 이스케이프(\$)인지 확인 → 이스케이프이면 리터럴로 변환
   - $$...$$이면 KaTeX.render(tex, element, {displayMode: true}) 호출
   - $...$이면 KaTeX.render(tex, element, {displayMode: false}) 호출
4. 오류 발생 시 throwOnError: false 옵션으로 오류 메시지 인라인 표시
```

### 4.2 Mermaid 다이어그램 초기화

```
1. DOM 로딩 완료 후 render-extensions.js 실행
2. <pre class="language-mermaid"> 요소를 모두 검색
3. 각 요소에 대해:
   - <pre> 태그를 <div class="mermaid">로 변환
   - 내부 텍스트를 Mermaid 구문으로 유지
4. mermaid.initialize({ startOnLoad: false, theme: 'default' }) 호출
5. mermaid.run({ querySelector: '.mermaid' }) 호출
6. 렌더링 오류 시 해당 요소에 오류 메시지 표시
```

### 4.3 go:embed 리소스 구성

```go
//go:embed js/katex.min.js js/mermaid.min.js js/render-extensions.js
//go:embed css/katex.min.css
//go:embed fonts/KaTeX_*.woff2
var ExtensionAssets embed.FS
```

### 4.4 성능 고려사항

- KaTeX/Mermaid JS는 defer 속성으로 로딩하여 초기 렌더링을 차단하지 않는다
- Mermaid는 문서당 최대 50개 다이어그램까지 렌더링한다 (성능 보호)
- 다이어그램 10개 이상 문서에서는 변경된 Mermaid 블록만 선택적으로 재렌더링. 전체 재렌더링은 최후 수단
- 대형 수식(500자 이상)에 대해 렌더링 타임아웃(5초)을 설정한다

---

## 5. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 파일 |
|------------|-----------|-------------|
| REQ-RENDER-001 | web/js/render-extensions.js | internal/markdown/renderer_test.go, web/js/render-extensions_test.js |
| REQ-RENDER-002 | web/js/render-extensions.js | internal/markdown/renderer_test.go, web/js/render-extensions_test.js |
| REQ-RENDER-003 | web/js/render-extensions.js | web/js/render-extensions_test.js |
| REQ-RENDER-004 | web/embed.go | internal/viewer/viewer_test.go |
| REQ-RENDER-005 | web/js/render-extensions.js | web/js/render-extensions_test.js |
| REQ-RENDER-006 | web/js/render-extensions.js | web/js/render-extensions_test.js |
| REQ-RENDER-007 | web/templates/viewer.html | internal/viewer/viewer_test.go |
| REQ-RENDER-008 | web/js/render-extensions.js | web/js/render-extensions_test.js |

### 관련 SPEC
- **선행**: SPEC-UI-001 (기본 마크다운 렌더링 및 WebView2 표시)
- **필수**: SPEC-WATCH-001 (HTTP 서버 - KaTeX 폰트 로딩 필수, 실시간 미리보기 시 재렌더링)
- **호환**: SPEC-THEME-001 (테마 변경 시 Mermaid 테마 동기화)

---

## 6. Implementation Notes (구현 완료 노트)

완료일: 2026-03-06

### 6.1 실제 구현 방식 요약

KaTeX 및 Mermaid 라이브러리 파일을 `go:embed`로 바이너리에 임베딩하고, HTTP 서버에 `/static/` 라우트를 추가하여 브라우저에서 직접 접근하도록 구현하였다. HTML 템플릿에서 KaTeX CSS를 `<head>`에 먼저 로딩하여 FOUC를 방지하고, JS 파일은 `defer` 속성으로 로딩하여 초기 렌더링을 차단하지 않도록 하였다. `render-extensions.js`가 DOM 로딩 완료 후 수식 감지 및 다이어그램 초기화를 실행한다.

### 6.2 계획 대비 실제 변경 사항

**계획에 없었으나 추가된 파일:**
- `internal/server/server.go`: 임베딩된 JS/CSS/폰트 파일을 브라우저에 제공하기 위한 `/static/` 라우트가 필요하여 수정하였다. 원래 계획에는 포함되지 않았으나, go:embed 파일에 HTTP를 통해 접근하려면 서버 측 라우팅이 필수적이었다.
- `internal/server/server_test.go`: 정적 파일 서빙 라우트에 대한 테스트 추가.
- `web/embed_test.go`: `ExtensionAssets` embed.FS 접근 테스트 추가.

**계획과 달리 변경되지 않은 파일:**
- `internal/markdown/renderer.go`: 계획에서는 goldmark 확장 설정 추가가 필요할 것으로 예상하였으나, 기존 goldmark 설정이 ` ```mermaid` 코드 블록을 이미 `<pre class="language-mermaid">` 태그로 올바르게 출력하고 있어 수정이 불필요하였다. Mermaid 렌더링은 전적으로 클라이언트 사이드 JS에서 처리한다.
- `internal/markdown/renderer_test.go`: renderer.go는 수정하지 않았으나, mermaid 코드 블록 HTML 출력 형태를 검증하는 테스트를 추가하였다.

### 6.3 테스트 커버리지

전체 7개 패키지 모두 테스트 통과:

| 패키지 | 커버리지 |
|--------|---------|
| app | 84.2% |
| config | 88.9% |
| markdown | 83.3% |
| server | 89.7% |
| viewer | 71.4% |
| watcher | 84.7% |

acceptance.md의 모든 Definition of Done 항목 충족.
