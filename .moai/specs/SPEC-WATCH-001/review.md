---
spec_id: SPEC-WATCH-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-WATCH-001 리뷰

## 리뷰 요약

실시간 미리보기 SPEC으로, SPEC-UI-001의 아키텍처를 SetHtml → Navigate(localhost) 방식으로 전환하는 중요한 변경을 포함합니다. 전체적으로 잘 구성되어 있으며, Debounce 전략, WebSocket 재연결, Graceful Shutdown 체인 등 세부 기술 결정이 명확합니다. 아래에 보완이 필요한 사항을 정리합니다.

---

## 1. spec.md 이슈

### 🟡 Warning (보완 권장)

#### W-1: gorilla/websocket 라이브러리 상태 확인 필요

- **위치**: spec.md §1.2 (line 33)
- **문제**: gorilla/websocket은 2022년 12월에 유지보수 중단(archived) 선언된 바 있음. 이후 다시 활성화되었으나, 장기적 안정성에 리스크가 있음
- **권장**: 대안 라이브러리(nhooyr.io/websocket, gobwas/ws 등) 검토를 리스크 대응에 포함하거나, gorilla/websocket의 현재 상태를 구현 전에 확인

#### W-2: REQ-U-004 스크롤 위치 보존의 기술적 한계

- **위치**: spec.md §3.1 (line 101)
- **문제**: `innerHTML` 교체 후 `window.scrollTo()` 복원 방식은 콘텐츠 길이가 크게 변경된 경우 의미 있는 위치 복원이 불가능. 예: 문서 상단에 대량 텍스트를 추가하면 기존 스크롤 위치의 콘텐츠가 달라짐
- **권장**: 스크롤 위치 보존의 "최선 노력(best-effort)" 성격을 명시하거나, 향후 개선 방향(앵커 기반 위치 추적 등)을 Optional 요구사항으로 언급

#### W-3: REQ-N-002 "전체 페이지 새로고침 금지"의 범위

- **위치**: spec.md §3.3 (line 114)
- **문제**: `innerHTML`로 전체 콘텐츠 교체도 사실상 "전체 교체"에 해당. "location.reload()를 사용하지 않는다"의 의미인지, "변경된 부분만 diff 기반으로 업데이트한다"의 의미인지 불명확
- **권장**: 요구사항의 의도를 명확히 - 예: "JavaScript location.reload() 또는 WebView2의 Navigate 재호출을 하지 않아야 한다"

#### W-4: 파일 삭제 후 재생성 시 감시 복구 메커니즘

- **위치**: spec.md §3.4 REQ-S-001 (line 119), §4.1 (line 138)
- **문제**: fsnotify는 파일이 삭제되면 감시가 해제되는 경우가 있음. "자동으로 감시를 재개"한다고 했으나, 재생성을 감지하려면 디렉토리를 감시해야 할 수 있음. 이에 대한 기술적 접근이 불명확
- **권장**: Specifications에 "파일 삭제 시 부모 디렉토리를 대신 감시하여 파일 재생성을 감지" 또는 "폴링 방식으로 파일 존재 여부를 주기적으로 확인" 등 구체적인 복구 메커니즘 명시

---

## 2. plan.md 이슈

### 🟡 Warning (보완 권장)

#### W-5: SPEC-UI-001의 기존 테스트 깨짐 가능성

- **위치**: plan.md §2 Task 10 (lines 119-124)
- **문제**: SetHtml → Navigate 전환은 SPEC-UI-001에서 작성한 viewer.go의 인터페이스와 테스트를 파괴할 수 있음. 기존 테스트의 수정 범위가 plan.md에 언급되지 않음
- **권장**: Task에 "기존 SPEC-UI-001 테스트 업데이트" 항목 추가, 또는 viewer.go의 인터페이스를 추상화하여 SetHtml/Navigate 모두 지원하는 설계 반영

#### W-6: 통합 테스트의 실현 가능성

- **위치**: plan.md §2 Task 11 (lines 126-132)
- **문제**: "파일 변경 → WebSocket 업데이트 흐름 테스트"가 통합 테스트로 명시되어 있으나, WebSocket 클라이언트가 WebView2 내부에서 실행되므로 Go 테스트에서 이 흐름을 자동화하기 어려움
- **권장**: 테스트 범위를 명확히 - Go 레벨에서는 "watcher → server.Broadcast()" 흐름만 테스트하고, "WebSocket → DOM 업데이트"는 수동 검증으로 분류

---

## 3. acceptance.md 이슈

### 🟡 Warning (보완 권장)

#### W-7: ACC-005 WebSocket 재연결 테스트 방법 미비

- **위치**: acceptance.md §2 ACC-005 (lines 62-71)
- **문제**: "네트워크 일시 장애 등으로 WebSocket 연결이 끊어진다"를 테스트하는 방법이 불명확. localhost 연결에서 네트워크 장애를 시뮬레이션하기 어려움
- **권장**: 테스트 방법을 구체적으로 명시 - 예: "서버 측에서 WebSocket 연결을 강제 종료하여 재연결을 유발" 또는 "JavaScript에서 ws.close()를 호출하여 시뮬레이션"

#### W-8: ACC-007을 별도 시나리오로 분리했는데, ACC-002와 논리적 중복

- **위치**: acceptance.md ACC-002 (lines 26-35) vs ACC-007 (lines 86-93)
- **문제**: ACC-002에서 이미 "WebSocket 연결이 수립되어 있다"를 Given 조건으로 하고, ACC-007에서 "처음 연결 시 현재 내용 전송"을 별도로 정의. 둘 다 WebSocket 초기 연결 시 동작을 다루고 있어 부분 중복
- **권장**: ACC-007의 시나리오가 더 명확하므로 유지하되, ACC-002의 Given 조건에서 "WebSocket 연결 수립" 과정이 ACC-007의 동작을 포함함을 참조 표기

---

## 4. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🟡 Warning | 8건 | W-1 ~ W-8 |

### 우선 조치 권장 순서

1. **gorilla/websocket 상태 확인** (W-1) — 대안 검토
2. **파일 삭제 시 감시 복구 메커니즘** (W-4) — 구현 복잡도에 영향
3. **기존 SPEC-UI-001 테스트 업데이트 계획** (W-5) — 파괴적 변경 관리
4. **나머지 Warning 순차 처리**
