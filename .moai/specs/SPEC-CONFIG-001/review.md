---
spec_id: SPEC-CONFIG-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-CONFIG-001 리뷰

## 리뷰 요약

사용자 설정 시스템 SPEC으로, JSON 기반 설정 파일, 검증 로직, 동시성 안전, 기본값 복원 등 견고한 설계가 돋보입니다. Go 표준 라이브러리만 사용하여 외부 의존성이 없는 점이 좋습니다. 아래에 보완이 필요한 사항을 정리합니다.

---

## 1. spec.md 이슈

### 🟡 Warning (보완 권장)

#### W-1: windowX/windowY 검증에서 "화면 크기" 획득 방법 미명시

- **위치**: spec.md §4.4 검증 규칙 (lines 186-195)
- **문제**: windowX는 "-1 또는 0 <= n <= 화면 너비", windowY는 "-1 또는 0 <= n <= 화면 높이"로 검증한다고 했으나, Go에서 화면 크기를 조회하는 방법이 명시되지 않음. 표준 라이브러리만으로는 불가능하며, Windows API(`GetSystemMetrics(SM_CXSCREEN)`) 호출이 필요
- **권장**: 
  1. windowX/windowY 검증을 단순화 - "-1이 아니면 any positive integer"로 완화 (화면 크기 검증은 viewer.go에서 수행)
  2. 또는 `golang.org/x/sys/windows`를 사용하여 화면 크기 조회 API 호출을 validator.go에 추가 (기술 스택에 추가 필요)

#### W-2: 단일 인스턴스(SPEC-WIN-001)와 설정 파일 동시 접근

- **위치**: spec.md §4.5 동시성 전략 (lines 197-202)
- **문제**: `sync.RWMutex`는 동일 프로세스 내 goroutine 간 동시성만 보호. SPEC-WIN-001의 단일 인스턴스 관리가 구현되면 한 프로세스만 실행되므로 문제가 완화되지만, 구현 순서상 SPEC-CONFIG-001이 먼저 구현될 수 있어 다중 프로세스 시나리오를 고려해야 함
- **권장**: spec.md §2 Assumptions에 "SPEC-WIN-001 미구현 시 다중 프로세스가 동시에 설정 파일을 접근할 수 있으며, last-write-wins 정책을 적용한다"를 명시 (현재 acceptance.md EC-005에만 언급)

#### W-3: customCSS 검증의 타이밍

- **위치**: spec.md §4.4 검증 규칙 (line 194)
- **문제**: "빈 문자열 또는 존재하는 파일 경로"로 검증한다고 했으나, 설정 로드 시점에 파일이 존재하더라도 렌더링 시점에는 삭제되었을 수 있음. 또한 상대 경로와 절대 경로 중 어떤 것을 지원하는지 미명시
- **권장**: 
  1. 경로 형식 정책 명시 (절대 경로만 허용 권장)
  2. 파일 존재 검증은 "로드 시점"에서만 수행하고, 렌더링 시점에 파일 미존재 시 graceful fallback

#### W-4: 설정 항목 확장 시 하위 호환성 전략

- **위치**: spec.md §2.2 A7 (line 69)
- **문제**: "설정 마이그레이션은 현재 범위에 포함하지 않는다"로 범위 외 처리. 그러나 plan.md Task 3.2에 "부분 설정 병합 (누락된 필드만 기본값으로 채우기)"가 있어 사실상 하위 호환성을 지원하고 있음
- **권장**: A7을 "설정 파일에 새 필드가 추가되면 기본값으로 자동 채워지며, 스키마 버전 관리(major 변경)는 범위 외"로 명확히 재기술

---

## 2. plan.md 이슈

### 🟡 Warning (보완 권장)

#### W-5: Task 2.1, 2.2의 viewer.go / main.go 의존성 불명확

- **위치**: plan.md §2 Secondary Goal (lines 43-49)
- **문제**: "창 닫기 시 자동 저장" (Task 2.1)이 viewer.go를 수정하지만, viewer.go는 SPEC-UI-001의 구현물. SPEC-UI-001에서 viewer.go에 "창 닫기 콜백" 인터페이스가 정의되어 있지 않으면 이 Task가 viewer.go의 인터페이스 변경을 요구
- **권장**: SPEC-UI-001의 viewer.go가 "창 닫기 이벤트" 콜백을 제공해야 한다는 선행 조건을 명시하거나, config 통합 시 viewer.go 인터페이스 확장을 별도 Task로 분리

#### W-6: 테스트에서 %APPDATA% 격리

- **위치**: plan.md §1.2 (line 27)
- **문제**: "통합 테스트: 실제 파일 시스템 I/O (t.TempDir() 활용)"이라고 했으나, 설정 파일 경로가 `os.UserConfigDir()` 기반이므로 테스트 시 실제 %APPDATA%에 파일을 생성할 수 있음
- **권장**: plan.md Task 4.1의 "ConfigPath 환경변수 오버라이드"를 Primary Goal로 격상하여, 테스트 격리를 먼저 보장

---

## 3. acceptance.md 이슈

### 🟡 Warning (보완 권장)

#### W-7: EC-001 "%APPDATA% 접근 불가" 시나리오 테스트 방법

- **위치**: acceptance.md §6 EC-001 (lines 169-177)
- **문제**: "%APPDATA% 디렉토리에 쓰기 권한이 없을 때"를 테스트하는 방법이 실질적으로 어려움. Windows에서 %APPDATA%는 사용자 프로필 디렉토리의 일부로, 쓰기 권한 제거가 시스템 불안정을 초래할 수 있음
- **권장**: "ConfigPath 오버라이드로 임시 디렉토리를 지정한 후, 해당 디렉토리에 읽기 전용 권한을 설정하여 테스트"하는 방법을 구체적으로 명시

#### W-8: AC-CFG-002의 fontSize 적용 확인 방법

- **위치**: acceptance.md §1 AC-CFG-002 (lines 24-30)
- **문제**: "마크다운 렌더링에 fontSize 20이 적용되어야 한다"의 검증 방법이 불명확. fontSize가 CSS의 어떤 속성에 대응하는지, WebView2에서 이를 어떻게 확인하는지 미명시
- **권장**: spec.md에 fontSize가 HTML 템플릿의 어떤 CSS 속성(예: `body { font-size: {{.FontSize}}px }`)에 매핑되는지 명시, 또는 acceptance에서 "body 요소의 font-size CSS 속성이 20px로 설정된다"로 구체화

---

## 4. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🟡 Warning | 8건 | W-1 ~ W-8 |

### 우선 조치 권장 순서

1. **ConfigPath 오버라이드 격상** (W-6) — 테스트 격리를 위해 Primary Goal로
2. **windowX/windowY 검증 방법 결정** (W-1) — 기술 스택에 영향
3. **viewer.go 인터페이스 선행 조건** (W-5) — SPEC-UI-001과의 조율
4. **customCSS 경로 정책** (W-3) — 절대/상대 경로 결정
5. **나머지 Warning 순차 처리**
