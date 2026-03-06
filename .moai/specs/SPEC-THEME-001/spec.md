---
id: SPEC-THEME-001
title: "Dark Mode Theme System"
version: 1.0.0
status: draft
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P3
lifecycle: spec-first
tags: [theme, dark-mode, css, prefers-color-scheme, toggle]
---

# SPEC-THEME-001: 다크 모드 테마 시스템

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
- CSS Custom Properties (CSS 변수) - 테마 색상 정의
- CSS `prefers-color-scheme` 미디어 쿼리 - 시스템 테마 감지
- JavaScript `matchMedia` API - 시스템 테마 변경 실시간 감지
- github.com/yuin/goldmark-highlighting (chroma 기반) - 코드 블록 구문 강조 테마

### 1.3 프로젝트 구조 (변경 사항)
```
WinMarkdownViewer/
  internal/viewer/viewer.go          # 초기 테마 class 설정
  internal/config/config.go          # 테마 설정 읽기 (SPEC-CONFIG-001 연동)
  web/
    templates/viewer.html            # 테마 class 적용, 토글 버튼 추가
    js/
      theme.js                       # 테마 전환 로직, 시스템 감지, 설정 저장
    css/
      theme-light.css                # 라이트 테마 CSS 변수 정의
      theme-dark.css                 # 다크 테마 CSS 변수 정의
      github-markdown.css            # CSS 변수 기반으로 리팩터링
      syntax-light.css               # 코드 구문 강조 라이트 테마 (chroma)
      syntax-dark.css                # 코드 구문 강조 다크 테마 (chroma)
    embed.go                         # go:embed 선언 업데이트
```

---

## 2. Assumptions (가정)

### 2.1 전제 조건
- SPEC-UI-001이 완료되어 기본 마크다운 렌더링과 WebView2 표시가 동작한다
- WebView2가 CSS Custom Properties와 `prefers-color-scheme` 미디어 쿼리를 지원한다
- CSS 전환 애니메이션이 WebView2에서 정상 동작한다

### 2.2 설계 결정
- CSS Custom Properties(CSS 변수) 기반으로 테마를 정의한다
  - 이유: `<html>` 태그의 class 전환만으로 전체 테마를 변경할 수 있다
  - 이유: 별도의 CSS 파일 교체 없이 동적 테마 전환이 가능하다
- 테마 모드는 "system" (기본), "light", "dark" 3가지를 제공한다
- 테마 설정 저장은 SPEC-CONFIG-001의 config.json을 활용한다
  - SPEC-CONFIG-001이 미구현인 경우, localStorage fallback을 사용한다
  - 주의: localStorage fallback은 SPEC-WATCH-001(HTTP 서버, localhost origin)이 전제 조건. SPEC-UI-001만(SetHtml, about:blank)에서는 localStorage 사용 불가

### 2.3 범위 외 (Out of Scope)
- 커스텀 테마 파일 로딩 (사용자가 직접 CSS 파일 작성)
- 테마 에디터 (GUI로 색상 편집)
- 색각 이상(color blindness) 대응 테마
- 시스템 트레이에서의 테마 전환 (SPEC-WIN-001 영역)

---

## 3. Requirements (요구사항)

### 3.1 CSS 테마 변수 시스템 [REQ-THEME-001]

시스템은 **항상** CSS Custom Properties 기반으로 테마 색상을 정의해야 한다

정의해야 하는 CSS 변수:
- `--bg-primary`: 문서 배경색
- `--bg-secondary`: 사이드바/헤더 배경색
- `--text-primary`: 본문 텍스트 색상
- `--text-secondary`: 부제목/메타 텍스트 색상
- `--text-link`: 링크 색상
- `--border-color`: 테두리/구분선 색상
- `--code-bg`: 인라인 코드 배경색
- `--code-block-bg`: 코드 블록 배경색
- `--blockquote-border`: 인용문 좌측 테두리 색상
- `--table-border`: 테이블 테두리 색상
- `--table-stripe`: 테이블 줄무늬 배경색

### 3.2 라이트 테마 [REQ-THEME-002]

**WHEN** 라이트 테마가 활성화되어 있을 때
**THEN** 시스템은 github-markdown-light 스타일을 기반으로 한 밝은 색상 스킴을 적용해야 한다

- 배경: 흰색 계열 (#ffffff 또는 #f6f8fa)
- 텍스트: 어두운 색 (#1f2328)
- 코드 블록: 밝은 회색 배경 (#f6f8fa)
- 코드 구문 강조: github 라이트 테마 (chroma)

### 3.3 다크 테마 [REQ-THEME-003]

**WHEN** 다크 테마가 활성화되어 있을 때
**THEN** 시스템은 github-markdown-dark 스타일을 기반으로 한 어두운 색상 스킴을 적용해야 한다

- 배경: 어두운 색 (#0d1117 또는 #161b22)
- 텍스트: 밝은 색 (#e6edf3)
- 코드 블록: 어두운 배경 (#161b22)
- 코드 구문 강조: github-dark 또는 monokai 테마 (chroma)

### 3.4 시스템 테마 자동 감지 [REQ-THEME-004]

**WHEN** 테마 모드가 "system"으로 설정되어 있을 때
**THEN** 시스템은 CSS `prefers-color-scheme` 미디어 쿼리를 통해 Windows 시스템 테마를 감지하고 자동으로 적용해야 한다

- Windows 다크 모드 → 다크 테마 적용
- Windows 라이트 모드 → 라이트 테마 적용

### 3.5 시스템 테마 실시간 반영 [REQ-THEME-005]

**WHILE** 테마 모드가 "system"으로 설정되어 있는 동안
**WHEN** Windows 시스템 설정에서 테마가 변경될 때
**THEN** 시스템은 뷰어를 재시작하지 않고 실시간으로 테마를 전환해야 한다

- JavaScript `matchMedia('(prefers-color-scheme: dark)')` 변경 이벤트 리스닝
- 테마 전환 시 전체 페이지를 새로고침하지 않는다

### 3.6 수동 테마 전환 - 키보드 단축키 [REQ-THEME-006]

**WHEN** 사용자가 Ctrl+Shift+D 키보드 단축키를 누를 때
**THEN** 시스템은 현재 테마를 순환 전환해야 한다

- 전환 순서: system → light → dark → system
- 전환 시 부드러운 트랜지션 애니메이션 적용 (CSS transition, 0.2s~0.3s)

### 3.7 수동 테마 전환 - UI 토글 버튼 [REQ-THEME-007]

**가능하면** 뷰어 UI에 테마 전환 토글 버튼을 제공한다

- 현재 테마 상태를 아이콘으로 표시 (해/달/시스템 아이콘)
- 클릭 시 테마 순환 전환 (REQ-THEME-006과 동일 동작)
- 뷰어 오른쪽 상단에 배치

### 3.8 테마 설정 저장 [REQ-THEME-008]

**WHEN** 사용자가 테마를 변경할 때
**THEN** 시스템은 변경된 테마 설정을 저장하여 다음 실행 시에도 유지해야 한다

- SPEC-CONFIG-001이 구현된 경우: config.json의 `theme` 필드에 저장
- SPEC-CONFIG-001이 미구현인 경우: `localStorage`에 저장 (fallback)
- 저장 값: "system" (기본), "light", "dark"

### 3.9 코드 블록 구문 강조 테마 연동 [REQ-THEME-009]

**WHEN** 테마가 전환될 때
**THEN** 시스템은 코드 블록의 구문 강조 색상도 해당 테마에 맞게 전환해야 한다

- 라이트 테마 → github 스타일 구문 강조 (syntax-light.css)
- 다크 테마 → github-dark 또는 monokai 스타일 구문 강조 (syntax-dark.css)
- goldmark-highlighting의 chroma 스타일을 CSS로 생성하여 go:embed
- 참고: chroma CSS는 chroma CLI로 사전 생성하여 소스 코드에 커밋. 런타임 생성 아님

### 3.10 테마 전환 트랜지션 [REQ-THEME-010]

**WHEN** 테마가 전환될 때
**THEN** 시스템은 부드러운 트랜지션 애니메이션을 적용해야 한다

- CSS transition 속성으로 배경색, 텍스트 색상의 전환을 0.2s~0.3s로 설정
- 시스템은 깜빡임(flash) 없이 자연스러운 전환을 제공해야 한다

### 3.11 초기 로딩 시 깜빡임 방지 [REQ-THEME-011]

시스템은 초기 로딩 시 테마 깜빡임(FOUC - Flash of Unstyled Content)이 발생**하지 않아야 한다**

- Go에서 HTML 템플릿 생성 시 저장된 테마 설정을 읽어 `<html>` 태그에 초기 class를 설정한다
- JavaScript의 테마 초기화보다 CSS가 먼저 적용되어야 한다

---

## 4. Specifications (사양)

### 4.1 CSS 테마 구조

```css
/* theme-light.css */
html.theme-light, html.theme-system {
  --bg-primary: #ffffff;
  --text-primary: #1f2328;
  /* ... 기타 변수 ... */
}

/* theme-dark.css */
html.theme-dark {
  --bg-primary: #0d1117;
  --text-primary: #e6edf3;
  /* ... 기타 변수 ... */
}

/* prefers-color-scheme 자동 감지 */
@media (prefers-color-scheme: dark) {
  html.theme-system {
    --bg-primary: #0d1117;
    --text-primary: #e6edf3;
    /* ... 다크 변수 오버라이드 ... */
  }
}
```

### 4.2 CSS 변수 리팩터링 시 시각적 회귀 테스트

SPEC-UI-001의 CSS를 CSS 변수 기반으로 리팩터링 시, 라이트 테마의 시각적 결과가 기존과 동일해야 한다. 시각적 회귀 테스트 수행 필수.

### 4.3 JavaScript 테마 관리

```
1. 페이지 로딩 시:
   a. localStorage 또는 config.json에서 저장된 테마 모드 읽기
   b. <html> 태그에 theme-{mode} class 설정
   c. 구문 강조 CSS 활성화 (라이트/다크)

2. 시스템 테마 변경 감지:
   a. matchMedia('(prefers-color-scheme: dark)').addEventListener('change', callback)
   b. theme-system 모드일 때만 반응

3. 수동 전환 시:
   a. 현재 모드에서 다음 모드로 순환 (system → light → dark → system)
   b. <html> class 업데이트
   c. 구문 강조 CSS 전환
   d. localStorage 또는 config.json에 저장

4. 키보드 단축키:
   a. document.addEventListener('keydown', handler)
   b. Ctrl+Shift+D 감지
   c. 수동 전환 로직 호출
```

### 4.3 Go 템플릿 초기 테마 설정

```
1. viewer.go에서 HTML 템플릿 렌더링 시:
   a. config.json에서 theme 설정 읽기 (기본값: "system")
   b. HTML 템플릿의 <html> 태그에 class="theme-{mode}" 설정
   c. 저장된 설정이 없으면 class="theme-system" 사용
```

### 4.4 구문 강조 CSS 생성

```
1. goldmark-highlighting의 chroma 스타일을 CSS로 생성:
   a. 라이트: chroma style "github" → syntax-light.css
   b. 다크: chroma style "github-dark" → syntax-dark.css
2. 두 CSS 파일 모두 go:embed로 임베딩
3. 테마 전환 시 <link> 태그의 href를 교체하거나 CSS class로 전환
```

---

## 5. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 파일 |
|------------|-----------|-------------|
| REQ-THEME-001 | web/css/theme-light.css, web/css/theme-dark.css | web/js/theme_test.js |
| REQ-THEME-002 | web/css/theme-light.css | web/js/theme_test.js |
| REQ-THEME-003 | web/css/theme-dark.css | web/js/theme_test.js |
| REQ-THEME-004 | web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-005 | web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-006 | web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-007 | web/templates/viewer.html, web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-008 | web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-009 | web/css/syntax-light.css, web/css/syntax-dark.css, web/js/theme.js | web/js/theme_test.js |
| REQ-THEME-010 | web/css/theme-light.css, web/css/theme-dark.css | web/js/theme_test.js |
| REQ-THEME-011 | internal/viewer/viewer.go | internal/viewer/viewer_test.go |

### 관련 SPEC
- **선행**: SPEC-UI-001 (기본 마크다운 렌더링 및 WebView2 표시)
- **권장**: SPEC-CONFIG-001 (테마 설정의 영구 저장)
- **담당**: SPEC-RENDER-001의 Mermaid 다이어그램 테마 연동은 SPEC-THEME-001이 담당. SPEC-RENDER-001은 Mermaid 렌더링만 담당하고 테마 적용은 SPEC-THEME-001에 위임
