---
spec-id: SPEC-INSTALL-001
title: "MSI Installer - Implementation Plan"
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
development-mode: tdd
---

# SPEC-INSTALL-001: MSI 설치 프로그램 - 구현 계획

## 1. TDD 접근 방식

### 1.1 테스트 전략

MSI 설치 프로그램은 전통적인 단위 테스트가 어렵기 때문에, 다음과 같은 다층 테스트 전략을 사용한다:

- **WiX 구조 검증 테스트**: WXS 파일의 XML 구조와 필수 요소를 검증하는 Go 테스트
- **빌드 스크립트 테스트**: PowerShell Pester 테스트로 build-msi.ps1의 매개변수 처리 및 단계 검증
- **MSI 검증 테스트**: 빌드된 MSI 파일의 내용을 검사하는 통합 테스트
- **설치/제거 E2E 테스트**: GitHub Actions에서 VM 또는 컨테이너를 사용한 설치-검증-제거 테스트

### 1.2 RED-GREEN-REFACTOR 사이클 적용

MSI 프로젝트 특성상, 다음과 같이 TDD 사이클을 적용한다:

1. **RED**: WiX XML 구조 검증 테스트 또는 PowerShell Pester 테스트 작성 (실패 확인)
2. **GREEN**: WXS 파일 또는 PowerShell 스크립트 작성 (테스트 통과)
3. **REFACTOR**: WXS 구조 개선, 공통 변수 추출, 중복 제거

---

## 2. 마일스톤 (우선순위 기반)

### Primary Goal: WiX v4 패키지 코어 구조

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 1.1 | Package.wxs 작성 | installer/wix/Package.wxs | 제품 정의, MajorUpgrade, Feature 트리 |
| Task 1.2 | Registry.wxs 작성 | installer/wix/Registry.wxs | 컨텍스트 메뉴 레지스트리 항목 |
| Task 1.3 | Shortcuts.wxs 작성 | installer/wix/Shortcuts.wxs | 시작 메뉴 바로가기 |
| Task 1.4 | Components.wxs 작성 | installer/wix/Components.wxs | 파일 컴포넌트 정의 |
| Task 1.5 | Directories.wxs 작성 | installer/wix/Directories.wxs | 설치 디렉토리 구조 |
| Task 1.6 | Variables.wxi 작성 | installer/wix/Variables.wxi | 공통 변수 (버전, 제품명) |

### Secondary Goal: 빌드 자동화

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 2.1 | build-msi.ps1 작성 | installer/build-msi.ps1 | Go 빌드 + WiX 빌드 + 검증 |
| Task 2.2 | Pester 테스트 작성 | installer/tests/build-msi.Tests.ps1 | 빌드 스크립트 테스트 |

### Tertiary Goal: CI/CD 워크플로우

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 3.1 | release.yml 작성 | .github/workflows/release.yml | v* 태그 기반 릴리스 자동화 |
| Task 3.2 | CI 통합 테스트 | .github/workflows/release.yml | MSI 빌드 검증 단계 추가 |

### Optional Goal: 파일 연결 기능

| Task ID | 작업 | 파일 | 설명 |
|---------|------|------|------|
| Task 4.1 | 파일 연결 Feature | installer/wix/Registry.wxs | .md 파일 연결 선택적 등록 |
| Task 4.2 | 설치 UI 커스터마이징 | installer/wix/Package.wxs | 체크박스 옵션 추가 |

---

## 3. 파일 영향 분석 (File Impact)

### 3.1 신규 생성 파일

| 파일 | 목적 | 크기 예상 |
|------|------|-----------|
| installer/wix/Package.wxs | WiX 메인 패키지 정의 | ~80 줄 |
| installer/wix/Directories.wxs | 디렉토리 구조 | ~20 줄 |
| installer/wix/Components.wxs | 파일 컴포넌트 | ~30 줄 |
| installer/wix/Registry.wxs | 레지스트리 항목 | ~60 줄 |
| installer/wix/Shortcuts.wxs | 시작 메뉴 바로가기 | ~25 줄 |
| installer/wix/Variables.wxi | 공통 변수 | ~15 줄 |
| installer/build-msi.ps1 | MSI 빌드 스크립트 | ~120 줄 |
| installer/tests/build-msi.Tests.ps1 | Pester 테스트 | ~80 줄 |
| .github/workflows/release.yml | CI/CD 워크플로우 | ~80 줄 |

### 3.2 수정 대상 파일

| 파일 | 변경 내용 |
|------|-----------|
| .gitignore | dist/, *.msi 패턴 추가 |
| README.md | 설치 방법 섹션 추가 (sync 단계에서) |

---

## 4. 의존성

### 4.1 선행 SPEC 의존성
- **SPEC-WIN-001**: 컨텍스트 메뉴 레지스트리 키 구조가 확정되어야 Registry.wxs 작성 가능
- **SPEC-UI-001**: winmdview.exe 바이너리가 빌드 가능해야 MSI 패키지 생성 가능

### 4.2 외부 도구 의존성
- WiX Toolset v4: `dotnet tool install --global wix`
- .NET SDK 8.0+: WiX v4 빌드에 필요
- Go 1.26.0: winmdview.exe 빌드
- PowerShell 5.1+: 빌드 스크립트 실행

### 4.3 후속 SPEC 영향
- 없음 (SPEC-INSTALL-001은 현재 로드맵의 리프 노드)

---

## 5. 기술적 접근 방향

### 5.1 WiX v4 패키지 설계

WiX v4는 이전 v3 대비 간소화된 XML 스키마를 사용한다:
- `<Wix>` 루트 대신 `<Package>` 루트 요소 사용
- `<Fragment>` 기반 모듈화로 각 관심사(레지스트리, 바로가기 등) 분리
- GUID 자동 생성(`*`) 활용으로 수동 GUID 관리 최소화
- `<StandardDirectory>` 참조로 ProgramFiles 등 표준 경로 사용

### 5.2 레지스트리 전략

SPEC-WIN-001에서 구현하는 레지스트리 키를 WiX 컴포넌트로 래핑한다:
- 설치 시 WiX가 레지스트리 키를 자동 생성
- 제거 시 WiX가 레지스트리 키를 자동 삭제 (레퍼런스 카운팅)
- 별도의 커스텀 액션 없이 선언적 레지스트리 관리

### 5.3 Major Upgrade 패턴

- UpgradeCode GUID를 고정하여 모든 버전에서 동일하게 유지
- `<MajorUpgrade>` 요소로 이전 버전 자동 제거
- 다운그레이드 시도 시 사용자에게 경고 메시지 표시

---

## 6. 리스크 및 대응 방안

| 리스크 | 발생 확률 | 영향도 | 대응 방안 |
|--------|----------|--------|-----------|
| WiX v4 빌드 환경 이슈 (GitHub Actions) | 중 | 높음 | dotnet tool install 사전 테스트, 캐싱 설정 |
| 레지스트리 권한 부족 (비관리자 실행) | 낮음 | 높음 | MSI는 기본적으로 관리자 권한 요구, UAC 프롬프트 |
| WebView2 Runtime 미설치 시 오류 | 중 | 중 | 설치 전 Prerequisites 체크 또는 README 안내 |
| 기존 .md 파일 연결 충돌 | 중 | 낮음 | 파일 연결을 선택적 Feature로 분리, 기본값 비활성 |
| Go 빌드 + WiX 빌드 시간 초과 (CI) | 낮음 | 중 | Go 빌드 캐시 + WiX 도구 캐시 활용 |
