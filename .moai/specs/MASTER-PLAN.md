---
project: WinMarkdownViewer
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
repository: https://github.com/AngeleyesTrue/WinMarkdownViewer
---

# WinMarkdownViewer - 마스터 플랜

## 프로젝트 비전

Windows 환경에서 .md 파일을 우클릭 또는 더블클릭으로 즉시 렌더링하는 경량 마크다운 뷰어.
Go + WebView2 기반, 단일 바이너리, 네트워크 불필요.

---

## 전체 구현 로드맵

```
Phase 1: 기초          Phase 2: 핵심         Phase 3: 배포          Phase 4: 확장
──────────────────  ──────────────────  ──────────────────  ──────────────────
 SPEC-UI-001         SPEC-WATCH-001      SPEC-INSTALL-001    SPEC-RENDER-001
 (기본 뷰어 MVP)     (실시간 미리보기)    (MSI 설치 프로그램)   (KaTeX/Mermaid)
       │                   │                    │                    │
       │              SPEC-WIN-001         SPEC-CONFIG-001     SPEC-THEME-001
       │              (Windows 통합)       (사용자 설정)        (다크 모드)
       │                   │                    │                    │
       ▼                   ▼                    ▼                    ▼
  .md → HTML 변환     파일 변경 감지       설치/제거 자동화      수식/다이어그램
  WebView2 표시       자동 새로고침        컨텍스트 메뉴 등록    테마 전환
                      우클릭 메뉴          파일 연결 설정
                      시스템 트레이
                      단일 인스턴스
```

---

## SPEC 의존성 그래프

```
SPEC-UI-001 (MVP 뷰어)
  ├──> SPEC-WATCH-001 (실시간 미리보기) ──> SPEC에 내장 HTTP 서버 + WebSocket 추가
  ├──> SPEC-WIN-001 (Windows 통합) ──> 컨텍스트 메뉴, 트레이, 단일 인스턴스
  │         └──> SPEC-INSTALL-001 (MSI) ──> 레지스트리 자동 등록 포함
  ├──> SPEC-CONFIG-001 (설정) ──> 모든 SPEC에서 참조하는 사용자 설정 시스템
  ├──> SPEC-RENDER-001 (확장 렌더링) ──> goldmark에 KaTeX/Mermaid 추가
  └──> SPEC-THEME-001 (테마) ──> CSS 테마 시스템 + 시스템 연동
```

**실행 순서 제약:**
- SPEC-UI-001 → 모든 후속 SPEC의 전제 조건
- SPEC-WATCH-001 → SPEC-WIN-001보다 먼저 (내장 서버가 단일 인스턴스의 기반)
- SPEC-WIN-001 → SPEC-INSTALL-001보다 먼저 (MSI가 레지스트리 등록을 자동화)
- SPEC-CONFIG-001 → SPEC-THEME-001보다 먼저 (테마 설정이 config에 저장)
- SPEC-RENDER-001, SPEC-THEME-001 → 독립적, 병렬 가능

---

## 각 SPEC 목표 및 완료 기준

### Phase 1: 기초

#### SPEC-UI-001: 기본 뷰어 MVP
- **목표**: .md 파일을 명령줄에서 열어 WebView2로 렌더링된 HTML을 표시한다
- **핵심 결과물**: winmdview.exe가 .md 파일을 인자로 받아 GitHub 스타일로 렌더링
- **완료 기준**: `winmdview.exe test.md` 실행 시 WebView2 창에 렌더링된 마크다운 표시
- **상태**: ✅ SPEC 작성 완료

### Phase 2: 핵심 기능

#### SPEC-WATCH-001: 실시간 미리보기
- **목표**: 외부 편집기에서 .md 파일을 수정하면 뷰어가 자동으로 새로고침된다
- **핵심 결과물**: fsnotify 파일 감시 + 내장 HTTP 서버 + WebSocket 실시간 통신
- **완료 기준**: VS Code에서 .md 파일 저장 시 1초 이내에 뷰어가 자동 업데이트
- **선행 조건**: SPEC-UI-001 완료

#### SPEC-WIN-001: Windows 통합
- **목표**: 파일 탐색기에서 .md 파일 우클릭 → "마크다운 뷰어로 열기"로 바로 열 수 있다
- **핵심 결과물**: 우클릭 컨텍스트 메뉴, 시스템 트레이 아이콘, 단일 인스턴스 관리
- **완료 기준**: 탐색기에서 .md 파일 우클릭 시 "마크다운 뷰어로 열기" 메뉴 표시 및 동작
- **선행 조건**: SPEC-UI-001 완료, SPEC-WATCH-001 권장 (내장 서버 재활용)

### Phase 3: 배포 및 설정

#### SPEC-INSTALL-001: MSI 설치 프로그램
- **목표**: 더블클릭으로 설치하면 컨텍스트 메뉴와 파일 연결이 자동 등록된다
- **핵심 결과물**: WiX Toolset 기반 MSI 패키지, 레지스트리 자동 등록/해제
- **완료 기준**: MSI 설치 후 .md 우클릭 메뉴 자동 등록, 제어판에서 깔끔한 제거
- **선행 조건**: SPEC-WIN-001 완료

#### SPEC-CONFIG-001: 사용자 설정
- **목표**: 사용자가 테마, 폰트 크기, 창 크기 등을 설정하고 다음 실행 시 유지할 수 있다
- **핵심 결과물**: JSON 기반 설정 파일, 설정 UI (WebView2 내 설정 페이지)
- **완료 기준**: 설정 변경 후 재시작해도 설정이 유지됨
- **선행 조건**: SPEC-UI-001 완료

### Phase 4: 확장 기능

#### SPEC-RENDER-001: 확장 렌더링 (KaTeX + Mermaid)
- **목표**: 마크다운 내 수학 수식과 다이어그램이 렌더링된다
- **핵심 결과물**: KaTeX.js 임베딩 (수식), Mermaid.js 임베딩 (다이어그램)
- **완료 기준**: `$$E=mc^2$$` 수식과 ` ```mermaid` 다이어그램이 시각적으로 렌더링
- **선행 조건**: SPEC-UI-001 완료

#### SPEC-THEME-001: 다크 모드 테마
- **목표**: 시스템 테마에 연동하거나 수동으로 라이트/다크 테마를 전환할 수 있다
- **핵심 결과물**: 다크/라이트 CSS, prefers-color-scheme 연동, 수동 전환 토글
- **완료 기준**: Windows 다크 모드 시 자동으로 다크 테마 적용, 수동 전환 가능
- **선행 조건**: SPEC-UI-001 완료, SPEC-CONFIG-001 권장

---

## 진행 추적

| SPEC ID | Phase | 목표 | 상태 | 비고 |
|---------|-------|------|------|------|
| SPEC-UI-001 | 1 | 기본 뷰어 MVP | 📋 SPEC 완료 | 구현 대기 |
| SPEC-WATCH-001 | 2 | 실시간 미리보기 | 📋 SPEC 완료 | 구현 대기 |
| SPEC-WIN-001 | 2 | Windows 통합 | 📋 SPEC 완료 | 구현 대기 |
| SPEC-INSTALL-001 | 3 | MSI 설치 프로그램 | 📋 SPEC 완료 | 구현 대기 |
| SPEC-CONFIG-001 | 3 | 사용자 설정 | 📋 SPEC 완료 | 구현 대기 |
| SPEC-RENDER-001 | 4 | KaTeX/Mermaid | 📋 SPEC 완료 | 구현 대기 |
| SPEC-THEME-001 | 4 | 다크 모드 | 📋 SPEC 완료 | 구현 대기 |

---

## 작업 흐름

각 SPEC은 다음 3단계로 진행됩니다:

```
/moai plan  →  SPEC 문서 생성 (요구사항, 구현 계획, 수용 기준)
/moai run   →  TDD 방식 구현 (RED → GREEN → REFACTOR)
/moai sync  →  문서 동기화 및 PR 생성
```

### 권장 실행 순서

```
1. /moai run SPEC-UI-001        ← 지금 여기
2. /moai run SPEC-WATCH-001
3. /moai run SPEC-WIN-001
4. /moai run SPEC-CONFIG-001
5. /moai run SPEC-INSTALL-001   ← WIN-001 이후
6. /moai run SPEC-RENDER-001    ← 독립, 언제든 가능
7. /moai run SPEC-THEME-001     ← CONFIG-001 이후 권장
```

---

Version: 1.0.0
