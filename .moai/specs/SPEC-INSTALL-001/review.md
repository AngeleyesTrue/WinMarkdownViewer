---
spec_id: SPEC-INSTALL-001
type: review
version: 1.0.0
created: 2026-03-06
reviewer: "Antigravity"
---

# SPEC-INSTALL-001 리뷰

## 리뷰 요약

MSI 설치 프로그램 SPEC으로, WiX Toolset v4 기반의 패키지 정의, 빌드 자동화, GitHub Actions CI/CD를 포함합니다. WiX v4의 새로운 스키마를 잘 반영하고 있으나, **SPEC-WIN-001과의 레지스트리 범위(HKCU vs HKLM) 불일치**라는 Critical 이슈가 있습니다.

---

## 1. spec.md 이슈

### 🔴 Critical (수정 필요)

#### C-1: HKCU vs HKLM 레지스트리 범위 불일치

- **위치**: spec.md §3.5 REQ-REG-001 (lines 138-142)
- **문제**: SPEC-INSTALL-001에서는 `HKLM\SOFTWARE\Classes\.md\shell\WinMarkdownViewer`에 컨텍스트 메뉴를 등록하지만, SPEC-WIN-001에서는 `HKCU\Software\Classes\.md\shell\WinMarkdownViewer`에 등록. 두 SPEC이 동일한 레지스트리 키를 다른 루트(HKLM vs HKCU)에 작성하므로:
  1. MSI 설치 후 `--unregister`를 실행하면 HKCU만 삭제되고 HKLM 키는 남음
  2. MSI 제거 후 `--register`를 실행하면 HKCU에만 등록되어 HKLM과 이중 등록 상태가 아님
  3. 양쪽 모두 등록되면 Windows가 어떤 것을 표시할지 예측 불가
- **권장 해결안**:
  - **방안 A**: MSI도 HKCU(per-user 설치)로 통일. MSI 설치 시에도 관리자 권한 없이 현재 사용자에게만 등록
  - **방안 B**: MSI는 HKLM(per-machine 설치), `--register`는 "MSI 비설치 환경(포터블)에서만 사용"으로 역할 분리. 이 경우 MSI가 설치된 환경에서 `--register` 실행 시 "MSI를 통해 이미 등록되어 있습니다" 메시지 출력
  - **방안 C**: `--register`를 제거하고 MSI 전용으로 통일 (포터블 배포는 범위 외)

### 🟡 Warning (보완 권장)

#### W-1: Frontmatter 키 이름 불일치

- **위치**: spec.md frontmatter (line 2), plan.md frontmatter (line 2), acceptance.md frontmatter (line 2)
- **문제**: spec.md는 `id:`, plan.md/acceptance.md는 `spec-id:` 키를 사용. 다른 SPEC(UI-001, WATCH-001 등)에서는 `id:`와 `spec_id:` 사용. 3가지 키 이름이 혼재
- **권장**: 전체 SPEC에서 frontmatter 키 이름 통일 (예: `spec_id:`)

#### W-2: WiX v4 파일명과 docs/structure.md 불일치

- **위치**: spec.md §1.3 (lines 38-53)
- **문제**: SPEC에서는 `Package.wxs`(WiX v4 신규 스키마)를 사용하지만, docs/structure.md에서는 `Product.wxs`(WiX v3 스타일)를 사용. WiX v4에서는 `<Product>` 대신 `<Package>` 루트 요소를 사용하므로 SPEC이 정확하지만, docs 문서 업데이트 필요
- **권장**: docs/structure.md의 `installer/wix/Product.wxs`를 `installer/wix/Package.wxs`로 수정

#### W-3: Assumptions의 A5 "관리자 권한" vs REQ-REG-001 HKLM

- **위치**: spec.md §2.1 A5 (line 70), §3.5 REQ-REG-001 (lines 139-142)
- **문제**: A5에서 "사용자는 관리자 권한으로 MSI를 실행할 수 있다"라고 가정. HKLM에 쓰려면 관리자 권한이 필수이므로 가정 자체는 맞지만, SPEC-WIN-001의 "관리자 권한 불필요" 설계 철학과 상충 
- **권장**: C-1의 해결안에 따라 정리

#### W-4: Pester 테스트 파일 위치

- **위치**: plan.md §2 Task 2.2 (line 52)
- **문제**: `installer/tests/build-msi.Tests.ps1` 경로가 spec.md §1.3의 디렉토리 구조에 포함되지 않음
- **권장**: spec.md §1.3의 디렉토리 구조에 `installer/tests/` 디렉토리 추가

#### W-5: REQ-UPGRADE-002의 수리(Repair) 모드 UI

- **위치**: spec.md §3.3 REQ-UPGRADE-002 (line 121)
- **문제**: "수리(Repair) 모드를 제공해야 한다"로만 되어 있으나, Windows Installer의 수리 모드는 자동 제공되므로 별도 구현이 불필요할 수 있음. 또는 MSI 다이얼로그 시퀀스에서 Repair/Remove 선택 UI를 제공해야 하는지 불명확
- **권장**: WiX의 기본 Repair 동작에 의존하는 것인지, 커스텀 UI를 구현하는 것인지 명확히 기술

---

## 2. acceptance.md 이슈

### 🟡 Warning (보완 권장)

#### W-6: AC-INST-002 컨텍스트 메뉴 텍스트 불일치

- **위치**: acceptance.md §1 AC-INST-002 (lines 23-33)
- **문제**: "WinMarkdownViewer로 열기"라고 되어 있으나, SPEC-WIN-001에서는 "마크다운 뷰어로 열기"로 등록. 메뉴 텍스트가 불일치
- **권장**: 하나로 통일. "마크다운 뷰어로 열기" 또는 "WinMarkdownViewer로 열기" 중 결정

#### W-7: EC-003 WebView2 미설치 시 동작

- **위치**: acceptance.md §6 EC-003 (lines 177-183)
- **문제**: "MSI 설치를 완료하고 winmdview.exe를 실행하면 오류 메시지 표시"라고 했으나, MSI 설치 단계에서 WebView2 확인을 하지 않고 실행 단계로 미룸. MSI 설치 시 Prerequisites 체크를 할 것인지(Launch Condition), 아니면 실행 시 에러 메시지로 처리할 것인지 결정 필요
- **권장**: MSI에 WebView2 Runtime 존재 여부를 Launch Condition으로 검사하거나, 또는 Bootstrapper(Burn)를 통해 WebView2 Evergreen 설치를 자동으로 포함하는 것을 검토

---

## 3. 리뷰 집계

| 등급 | 건수 | ID |
|------|------|-----|
| 🔴 Critical | 1건 | C-1 |
| 🟡 Warning | 7건 | W-1 ~ W-7 |

### 우선 조치 권장 순서

1. **HKCU vs HKLM 정책 통일** (C-1) — SPEC-WIN-001과 함께 결정
2. **컨텍스트 메뉴 텍스트 통일** (W-6) — SPEC-WIN-001과 일치시키기
3. **WiX 파일명 docs/structure.md 업데이트** (W-2)
4. **Frontmatter 키 이름 통일** (W-1) — 전체 SPEC 대상
5. **나머지 Warning 순차 처리**
