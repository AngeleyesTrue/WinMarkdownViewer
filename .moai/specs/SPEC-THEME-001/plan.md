---
id: SPEC-THEME-001
type: plan
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
methodology: TDD
---

# SPEC-THEME-001: 구현 계획 - 다크 모드 테마 시스템

## 1. TDD 접근 방식

본 SPEC은 TDD (RED-GREEN-REFACTOR) 방식으로 구현한다.

### 1.1 RED-GREEN-REFACTOR 전략

- **RED**: CSS 변수 적용, 테마 전환, 시스템 감지에 대한 실패하는 테스트를 먼저 작성
- **GREEN**: 테스트를 통과하는 최소한의 CSS/JS/Go 코드 구현
- **REFACTOR**: CSS 변수 정리, JavaScript 모듈화, 중복 제거

---

## 2. 마일스톤 (우선순위 기반)

### Primary Goal: CSS 테마 변수 시스템 구축

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 1 | CSS Custom Properties 정의 (라이트 기본값) | web/css/theme-light.css | CSS 변수 존재 검증 |
| 2 | 다크 테마 CSS 변수 정의 | web/css/theme-dark.css | CSS 변수 값 검증 |
| 3 | github-markdown.css를 CSS 변수 기반으로 리팩터링 | web/css/github-markdown.css | 하드코딩된 색상 없음 검증. 라이트 테마의 시각적 결과가 기존과 동일해야 함 (시각적 회귀 테스트) |
| 4 | HTML 템플릿에 테마 CSS 로딩 추가 | web/templates/viewer.html | 테마 CSS 포함 검증 |
| 5 | go:embed 선언 업데이트 | web/embed.go | embed.FS 접근 테스트 |

### Secondary Goal: 테마 전환 기능

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 6 | theme.js 기본 구조 구현 (class 전환) | web/js/theme.js | class 전환 동작 테스트 |
| 7 | 시스템 테마 감지 구현 (matchMedia) | web/js/theme.js | prefers-color-scheme 감지 테스트 |
| 8 | 시스템 테마 변경 실시간 리스닝 | web/js/theme.js | 이벤트 리스너 동작 테스트 |
| 9 | 키보드 단축키 (Ctrl+Shift+D) 구현 | web/js/theme.js | 키보드 이벤트 테스트 |
| 10 | UI 토글 버튼 추가 | web/templates/viewer.html, web/js/theme.js | 버튼 클릭 전환 테스트 |
| 11 | 테마 전환 트랜지션 애니메이션 | web/css/theme-light.css, web/css/theme-dark.css | 트랜지션 속성 검증 |

### Final Goal: 설정 저장 및 구문 강조 연동

**태스크 분해:**

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 12 | 테마 설정 저장/로딩 (localStorage fallback) | web/js/theme.js | 저장/복원 테스트 |
| 13 | SPEC-CONFIG-001 연동 (config.json 읽기/쓰기) | internal/config/config.go, web/js/theme.js | config 연동 테스트 |
| 14 | 코드 구문 강조 CSS 생성 (chroma github/github-dark) | web/css/syntax-light.css, web/css/syntax-dark.css | 스타일 적용 검증 |
| 15 | 구문 강조 테마 전환 연동 | web/js/theme.js | 테마 전환 시 구문 강조 변경 검증 |
| 16 | 초기 로딩 FOUC 방지 (Go 템플릿 초기 class 설정) | internal/viewer/viewer.go | 초기 class 설정 테스트 |

### Optional Goal: Mermaid 테마 연동

| 순서 | 태스크 | 파일 영향 | 테스트 대상 |
|------|--------|-----------|-------------|
| 17 | Mermaid 다이어그램 테마 연동 (SPEC-THEME-001 담당) | web/js/theme.js, web/js/render-extensions.js | 다이어그램 테마 전환 테스트. SPEC-RENDER-001은 Mermaid 렌더링만 담당, 테마 적용은 본 SPEC에서 처리 |

---

## 3. 파일 영향 분석

### 신규 파일

| 파일 경로 | 목적 |
|-----------|------|
| web/css/theme-light.css | 라이트 테마 CSS Custom Properties 정의 |
| web/css/theme-dark.css | 다크 테마 CSS Custom Properties 정의 |
| web/css/syntax-light.css | chroma github 스타일 구문 강조 CSS |
| web/css/syntax-dark.css | chroma github-dark 스타일 구문 강조 CSS |
| web/js/theme.js | 테마 전환 로직, 시스템 감지, 설정 저장 |

### 수정 파일

| 파일 경로 | 변경 내용 |
|-----------|-----------|
| web/css/github-markdown.css | 하드코딩된 색상을 CSS 변수로 교체 |
| web/templates/viewer.html | 테마 CSS 로딩, 토글 버튼, theme.js 스크립트 추가 |
| web/embed.go | 신규 CSS/JS 파일에 대한 go:embed 선언 추가 |
| internal/viewer/viewer.go | HTML 템플릿에 초기 테마 class 설정 로직 추가 |
| internal/markdown/renderer.go | 코드 구문 강조 CSS class 설정 조정 |

### 테스트 파일

| 파일 경로 | 테스트 범위 |
|-----------|-------------|
| web/js/theme_test.js | 테마 전환, 시스템 감지, 키보드 단축키, 설정 저장 |
| internal/viewer/viewer_test.go | 초기 테마 class 설정 검증 |
| internal/markdown/renderer_test.go | 구문 강조 CSS class 검증 |

---

## 4. 의존성

### SPEC 의존성
- **필수**: SPEC-UI-001 (기본 뷰어 MVP) - HTML 템플릿과 CSS 구조 기반
- **권장**: SPEC-CONFIG-001 (사용자 설정) - 테마 설정의 영구 저장
  - 미구현 시 localStorage fallback 사용 (단, SPEC-WATCH-001의 HTTP 서버가 전제 조건. SPEC-UI-001만으로는 about:blank origin에서 localStorage 사용 불가)
- **호환**: SPEC-RENDER-001 (확장 렌더링) - Mermaid 다이어그램 테마 연동

### 외부 의존성
- goldmark-highlighting (chroma) - 구문 강조 CSS 생성
  - chroma CLI로 사전 생성하여 소스 코드에 커밋 (런타임 생성 아님)

---

## 5. 기술적 접근 방식

### 5.1 CSS Custom Properties 기반 테마 전환

**전략**: `<html>` 태그의 class를 변경하여 CSS 변수 값을 전환한다.

```
<html class="theme-light"> → 라이트 테마 변수 활성화
<html class="theme-dark">  → 다크 테마 변수 활성화
<html class="theme-system"> → 미디어 쿼리 기반 자동 선택
```

**장점:**
- JavaScript에서 단일 class 변경만으로 전체 테마 전환
- CSS 파일 교체나 동적 스타일 삽입 불필요
- 트랜지션 애니메이션 자연스럽게 적용

### 5.2 시스템 테마 감지 전략

```javascript
// 시스템 다크 모드 감지
const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');

// 실시간 변경 감지
darkModeQuery.addEventListener('change', (e) => {
  if (currentMode === 'system') {
    applySystemTheme(e.matches);
  }
});
```

### 5.3 테마 설정 저장 전략

**우선순위:**
1. SPEC-CONFIG-001의 config.json (Go에서 읽기/쓰기)
2. localStorage (JavaScript fallback)

**Go 연동:**
- viewer.go에서 config.json의 theme 값을 읽어 HTML 템플릿에 전달
- FOUC 방지를 위해 서버 사이드에서 초기 class를 설정

### 5.4 구문 강조 CSS 생성 전략

chroma CLI 또는 Go 코드로 CSS를 사전 생성하여 go:embed에 포함한다:

```
chroma --style=github --html-styles > web/css/syntax-light.css
chroma --style=github-dark --html-styles > web/css/syntax-dark.css
```

테마 전환 시 CSS class 또는 `<link>` 태그 전환으로 구문 강조 테마를 교체한다.

---

## 6. 리스크 및 대응 방안

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| FOUC (Flash of Unstyled Content) 발생 | 높음 | Go 템플릿에서 초기 class 설정으로 해결 |
| SPEC-CONFIG-001 미구현 시 설정 저장 | 중간 | localStorage fallback 구현으로 독립 동작 보장 |
| github-markdown.css 리팩터링 시 기존 스타일 깨짐 | 높음 | 리팩터링 전후 시각적 회귀 테스트 수행. 라이트 테마의 시각적 결과가 기존과 동일해야 함 |
| CSS 트랜지션이 일부 요소에서 깜빡임 | 낮음 | transition 대상 속성을 명시적으로 지정 (all 사용 금지) |
| Mermaid 다이어그램이 테마 전환에 반응하지 않음 | 중간 | Mermaid.initialize()에서 theme 옵션을 동적으로 변경 |
| goldmark-highlighting CSS 출력이 테마 전환과 호환되지 않음 | 중간 | class 기반 토글로 라이트/다크 CSS를 선택적 적용 |
