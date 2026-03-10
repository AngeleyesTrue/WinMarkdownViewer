---
id: SPEC-ICON-001
title: Application Icon & File Association Icon - Implementation Plan
version: 1.0.0
status: Planned
created: 2026-03-10
updated: 2026-03-10
author: Claud Archive
priority: High
---

# SPEC-ICON-001: 구현 계획

## 1. 마일스톤

### Primary Goal: 아이콘 파일 교체 및 EXE 리소스 임베딩

**범위:** 핵심 아이콘 인프라 구축

**작업 목록:**

1. **멀티 사이즈 ICO 파일 준비** [SPEC-ICON-001-R1]
   - 사용자가 제공한 아이콘 파일 사용 (6개 크기: 16x16, 32x32, 48x48, 64x64, 128x128, 256x256 포함, 101.4KB)
   - `assets/icon.ico` 교체 완료 (사용자 제공)

2. **goversioninfo 설정** [SPEC-ICON-001-R2]
   - `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest`
   - `cmd/winmdview/versioninfo.json` 매니페스트 파일 생성
   - `IconPath`를 `../../assets/icon.ico`로 설정
   - `FileVersion`, `ProductVersion`, `FileDescription` 등 메타데이터 설정

3. **빌드 파이프라인 수정** [SPEC-ICON-001-R6]
   - `build.ps1`의 `Invoke-Build` 함수에 goversioninfo 실행 단계 추가
   - `build.ps1`의 `Invoke-Dev` 함수에 동일 단계 추가
   - `Invoke-Clean`에 `cmd/winmdview/resource.syso` 전체 경로 삭제 추가
   - `.gitignore`에 `*.syso` 추가

4. **빌드 검증**
   - `build.ps1 build` 실행 후 exe에 아이콘 포함 확인
   - 탐색기에서 exe 아이콘 표시 확인
   - 기존 테스트 통과 확인 (`go test ./...`)

### Secondary Goal: 인스톨러 및 파일 연결 아이콘

**범위:** WiX 인스톨러 아이콘 및 .md 파일 연결 아이콘

**작업 목록:**

5. **WiX Registry.wxs 수정** [SPEC-ICON-001-R4]
   - `WinMarkdownViewer.md` ProgID에 `DefaultIcon` 레지스트리 항목 추가
   - `DefaultIcon` 값: `[INSTALLFOLDER]winmdview.exe,0`
   - 새 컴포넌트를 `FileAssociationComponents` 그룹에 추가

6. **인스톨러 빌드 검증** [SPEC-ICON-001-R3]
   - `build-msi.ps1`로 MSI 빌드
   - exe에 리소스 아이콘이 포함되었으므로 WiX `<Icon>` 참조 정상 동작 확인
   - MSI 설치 후 프로그램 추가/제거 아이콘 확인
   - 시작 메뉴 바로가기 아이콘 확인
   - 컨텍스트 메뉴 아이콘 확인

7. **파일 연결 아이콘 검증** [SPEC-ICON-001-R4]
   - MSI 설치 시 FileAssociation 기능 활성화
   - 탐색기에서 .md 파일에 WinMarkdownViewer 아이콘 표시 확인
   - 아이콘 캐시 갱신 후에도 정상 표시 확인

### Final Goal: 테스트 및 정리

**범위:** 품질 보증 및 코드 정리

**작업 목록:**

8. **기존 테스트 업데이트** [SPEC-ICON-001-R5]
   - `assets/embed_test.go`의 ICO 유효성 테스트가 멀티 사이즈 ICO에서도 통과하는지 확인
   - 필요 시 ICO 이미지 개수(Count) 검증 테스트 추가 (6개 이미지 확인)
   - 시스템 트레이 아이콘 정상 표시 확인 (수동 테스트)

9. **문서 업데이트**
   - `tech.md`에 goversioninfo 관련 항목 업데이트 (현재 "rsrc 또는 goversioninfo" 언급 존재)
   - `build.ps1` 도움말에 goversioninfo 의존성 안내 추가

## 2. 기술 접근 방식

### 2.1 goversioninfo 선택 근거

| 도구 | CGO 필요 | 기능 | 선택 |
|------|----------|------|------|
| goversioninfo | 불필요 | 아이콘 + 버전 정보 + 매니페스트 | 선택 |
| rsrc | 불필요 | 아이콘만 (버전 정보 없음) | 미선택 |
| windres (GCC) | 필요 | 전체 리소스 컴파일 | 미선택 (CGO 금지) |

goversioninfo는 CGO 없이 아이콘과 버전 정보를 동시에 임베딩할 수 있어 프로젝트 요구사항에 가장 적합하다.

### 2.2 .syso 파일 위치

Go 컴파일러는 빌드 대상 패키지 디렉터리(이 경우 `cmd/winmdview/`)에 있는 `.syso` 파일을 자동으로 링크한다. 별도의 빌드 플래그나 설정이 필요 없다.

### 2.3 아이콘 디자인 방향

- 마크다운의 "M" 또는 "#" 기호를 활용한 심플한 디자인
- 다크/라이트 배경 모두에서 가시성 확보
- 16x16에서도 식별 가능한 단순한 형태
- 기존 마크다운 에디터들과 차별화된 색상 선택

### 2.4 ICO 파일 생성 방법

1. PNG 원본 256x256으로 디자인
2. ImageMagick 또는 온라인 도구로 멀티 사이즈 ICO 생성
3. 또는 GIMP/Photoshop으로 직접 ICO 편집

## 3. 아키텍처 설계 방향

### 3.1 변경 영향 분석

| 파일 | 변경 유형 | 설명 |
|------|-----------|------|
| `assets/icon.ico` | 교체 | 멀티 사이즈 ICO로 교체 |
| `assets/embed.go` | 변경 없음 | 기존 go:embed 유지 |
| `assets/embed_test.go` | 수정 | 멀티 사이즈 ICO 검증 추가 |
| `cmd/winmdview/versioninfo.json` | 신규 | goversioninfo 매니페스트 |
| `build.ps1` | 수정 | goversioninfo 실행 단계 추가 |
| `installer/wix/Registry.wxs` | 수정 | DefaultIcon 레지스트리 추가 |
| `.gitignore` | 수정 | *.syso 추가 |

### 3.2 의존성

- `goversioninfo`: Go 도구로 `go install`로 설치
- 런타임 의존성 추가 없음
- 기존 API 변경 없음

## 4. 리스크 및 대응

| 리스크 | 영향 | 대응 방안 |
|--------|------|-----------|
| goversioninfo가 Go 1.26+에서 호환 안 됨 | 빌드 실패 | rsrc 도구로 폴백 (아이콘만 임베딩) |
| 멀티 사이즈 ICO 파일이 너무 클 경우 | 바이너리 크기 증가 | 256x256은 PNG 압축 사용, 전체 크기 200KB 이하 유지 |
| 탐색기 아이콘 캐시 문제 | 아이콘 미표시 | ie4uinit.exe -show 또는 재부팅으로 캐시 갱신 안내 |
| WiX가 새 리소스 아이콘을 인식 못 함 | MSI 아이콘 누락 | SourceFile을 직접 ICO 파일로 변경하는 폴백 |

## 5. 전문가 상담 권장

### 권장 상담

이 SPEC는 순수한 Windows 데스크톱 빌드 파이프라인 변경이므로 별도의 전문가 에이전트 상담 없이 구현 가능하다.

필요 시:
- **expert-backend**: goversioninfo 설정 최적화 및 빌드 스크립트 통합
- **expert-devops**: CI/CD (GitHub Actions)에서 goversioninfo 통합

## 6. 후속 과제

- CI/CD (GitHub Actions) `release.yml`에 `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest` 단계 추가
- `docs/tech.md`의 systray 라이브러리 이름 수정 (`github.com/energye/systray`)
