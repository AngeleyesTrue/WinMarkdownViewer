---
id: SPEC-ICON-001
title: Application Icon & File Association Icon - Review
version: 1.0.0
status: Reviewed
created: 2026-03-10
updated: 2026-03-10
reviewer: Antigravity
result: Conditional Approval (수정 필요 항목 존재)
---

# SPEC-ICON-001: 설계 문서 리뷰

## 리뷰 요약

| 항목 | 판정 |
|------|------|
| spec.md | ⚠️ 수정 필요 (Minor Issues 3건, Suggestion 2건) |
| plan.md | ⚠️ 수정 필요 (Minor Issues 2건, Suggestion 1건) |
| acceptance.md | ✅ 양호 (Suggestion 1건) |
| 전체 종합 | ⚠️ Conditional Approval |

---

## 1. spec.md 리뷰

### ✅ 잘된 점

- **현재 상태 분석이 정확함**: `assets/icon.ico` (16x16 단일), `.syso` 파일 없음, `DefaultIcon` 미등록 등 현재 코드베이스와 정확히 일치
- **요구사항 분류 체계 우수**: UBIQUITOUS/EVENT/STATE/UNWANTED 패턴을 적절히 사용
- **아이콘 파이프라인 시각화 (섹션 4.6)**: 전체 흐름을 한 눈에 파악 가능
- **제약 조건이 명확함**: CGO 금지, 기존 인터페이스 보존, GUID 변경 금지 등

### ⚠️ 수정 필요

#### Issue 1: `tech.md`의 systray 라이브러리 불일치 [Medium]

**현상:**
- `spec.md` 섹션 1.2에서 시스템 트레이 라이브러리를 `github.com/energye/systray`로 기술
- `docs/tech.md`에서는 `github.com/getlantern/systray`로 기술
- 실제 코드(`internal/tray/tray.go`)에서는 `github.com/energye/systray`를 import

**권장 조치:**
- `docs/tech.md`의 systray 라이브러리를 `github.com/energye/systray`로 업데이트 (이것은 SPEC 범위 밖이지만 plan.md 섹션 9 "문서 업데이트"에 포함 권장)
- spec.md의 기술은 현재 코드와 일치하므로 spec.md는 수정 불필요

#### Issue 2: Package.wxs `<Icon>` 참조 방식 주의 [Medium]

**현상:**
- `spec.md` 섹션 4.5에서 "`<Icon Id="AppIcon.exe" SourceFile="$(var.ExeFileName)" />`이 정상 동작한다. 변경 불필요."라고 기술
- 현재 `Package.wxs`의 `<Icon>` 태그에서 `SourceFile="$(var.ExeFileName)"`이 **상대 경로**로 되어 있음

**잠재 문제:**
- WiX 빌드 시 exe 파일의 실제 위치는 `dist/winmdview.exe`이지만, `$(var.ExeFileName)`은 `winmdview.exe`로만 정의되어 있음 (`Variables.wxi` 참조)
- `build-msi.ps1`에서 WiX 빌드 시 exe 파일 위치를 어떻게 resolve하는지에 따라 동작이 달라질 수 있음
- spec에서 "변경 불필요"라고 단정하기보다, **"exe에 리소스가 임베딩된 후 기존 경로가 유효한지 빌드 검증 필요"** 라는 노트를 추가 권장

**권장 조치:**
- 섹션 4.5를 "변경 불필요 (빌드 시 exe 경로가 올바르게 resolve되는지 검증 필요)"로 보완

#### Issue 3: `versioninfo.json`에 대한 구체적 예시 부재 [Low]

**현상:**
- 섹션 4.2에서 주요 설정 항목만 나열했으나, 실제 `versioninfo.json`의 전체 구조/예시가 없음
- goversioninfo의 JSON 스키마는 필수 필드가 여러 개 있어, 구현 시 혼란이 있을 수 있음

**권장 조치:**
- `versioninfo.json`의 최소 샘플 JSON을 포함하여 구현자가 바로 사용할 수 있도록 보완

```json
{
  "FixedFileInfo": {
    "FileVersion": { "Major": 1, "Minor": 0, "Patch": 0, "Build": 0 },
    "ProductVersion": { "Major": 1, "Minor": 0, "Patch": 0, "Build": 0 },
    "FileType": "VFT_APP"
  },
  "StringFileInfo": {
    "FileDescription": "WinMarkdownViewer",
    "ProductName": "WinMarkdownViewer",
    "CompanyName": "WinMarkdownViewer",
    "LegalCopyright": "",
    "InternalName": "winmdview",
    "OriginalFilename": "winmdview.exe"
  },
  "IconPath": "../../assets/icon.ico"
}
```

### 💡 제안 사항

#### Suggestion 1: 컨텍스트 메뉴 아이콘의 동작 방식 명확화

- R3에서 "컨텍스트 메뉴에 앱 아이콘이 표시되어야 한다"고 했는데, 현재 `Registry.wxs` 컨텍스트 메뉴 등록에 이미 `<RegistryValue Name="Icon" Type="string" Value="[INSTALLFOLDER]$(var.ExeFileName)" />`이 존재함
- 이 Icon 값은 **exe 파일의 첫 번째 리소스 아이콘**을 사용하므로, exe에 아이콘이 임베딩되면 자동으로 해결됨
- 이 점을 spec에 명시하면 추적성이 향상됨 (현재 Registry.wxs에 이미 존재하는 Icon 참조가 R3를 자연스럽게 충족한다는 점)

#### Suggestion 2: goversioninfo Go 모듈 tool 디렉티브 구체화

- C5에서 "goversioninfo는 Go 모듈의 `tool` 디렉티브 또는 `go install`로 관리"라고 했으나, Go 1.24+의 tool 디렉티브와 `go install`의 두 가지 접근 방식 중 **어떤 것을 사용할지 명확하게 결정**하면 구현 혼선 방지

---

## 2. plan.md 리뷰

### ✅ 잘된 점

- **마일스톤 구분이 논리적**: Primary(인프라) → Secondary(인스톨러/파일연결) → Final(테스트/정리) 단계가 합리적
- **goversioninfo 선택 근거 표**: rsrc, windres와의 비교가 명쾌
- **리스크 분석표**: 현실적인 리스크와 대응 방안 제시 (아이콘 캐시, ICO 크기 등)
- **변경 영향 분석**: 변경/미변경 파일이 명확히 구분됨

### ⚠️ 수정 필요

#### Issue 1: `Invoke-Clean` 기존 코드와의 정합성 [Medium]

**현상:**
- plan.md 작업 3에서 "`Invoke-Clean`에 `resource.syso` 삭제 추가"라고 했으나, 현재 `build.ps1`의 `Invoke-Clean` 함수는 정적 경로 배열(`$targets`)을 사용
- `cmd/winmdview/resource.syso`는 프로젝트 루트가 아닌 하위 디렉터리에 생성되므로, 기존 패턴인 `$targets`에 전체 상대 경로 `cmd/winmdview/resource.syso`를 추가해야 함

**권장 조치:**
- 구현 시 `$targets` 배열에 `"cmd/winmdview/resource.syso"`를 추가하도록 명시

#### Issue 2: 아이콘 디자인 작업이 계획에만 언급되고 담당이 불명확 [Low]

**현상:**
- 작업 1 "멀티 사이즈 ICO 파일 준비"에서 "마크다운 뷰어 컨셉의 아이콘 디자인"이 포함되어 있으나, 이것이 수동 작업(디자이너)인지, AI 도구 생성인지, 기존 오픈소스 아이콘 활용인지 불명확
- 아이콘 디자인은 기술적 구현과 별개의 작업이므로 담당/방법을 명확히 해야 함

**권장 조치:**
- 아이콘 디자인 작업의 수행 방법(외부 도구로 제작, AI 생성, 오픈소스 활용 등)과 담당을 명시

### 💡 제안 사항

#### Suggestion 1: CI/CD 연동 고려

- plan.md 섹션 5 전문가 상담에서 "`expert-devops`: CI/CD (GitHub Actions)에서 goversioninfo 통합"을 언급했으나, 현재 계획에는 CI/CD 변경이 포함되어 있지 않음
- GitHub Actions workflow에 `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest` 단계 추가가 필요할 수 있으므로, 별도 작업으로 추가하거나 "후속 과제"로 명시 권장

---

## 3. acceptance.md 리뷰

### ✅ 잘된 점

- **Gherkin 형식의 시나리오가 명확**: Given/When/Then 구조로 검증 조건이 구체적
- **모든 요구사항(R1~R6)을 빠짐없이 커버**: 각 요구사항에 대한 인수 기준이 존재
- **품질 게이트(Definition of Done)와 검증 방법 표**: 자동/수동 검증 구분이 명확
- **네거티브 시나리오 포함**: Scenario 4.3 (파일 연결 비활성화 시) 등 엣지 케이스 고려

### 💡 제안 사항

#### Suggestion 1: exe 버전 정보 검증의 자동화 가능성

- Scenario 2.3에서 "파일 속성 > 자세히 탭"을 수동 확인하도록 했으나, PowerShell의 `[System.Diagnostics.FileVersionInfo]::GetVersionInfo()` 로 자동 검증 가능
- 빌드 후 자동 검증 스크립트를 추가하면 품질 게이트를 자동화할 수 있음

```powershell
# 자동 검증 예시
$info = [System.Diagnostics.FileVersionInfo]::GetVersionInfo("dist/winmdview.exe")
if ($info.FileDescription -ne "WinMarkdownViewer") { throw "FileDescription mismatch" }
if ($info.ProductName -ne "WinMarkdownViewer") { throw "ProductName mismatch" }
```

---

## 4. 추가 발견 사항 (코드베이스 교차 검증)

### 4.1 `.gitignore` 현황

현재 `.gitignore`에 `*.syso`가 **이미 포함되어 있지 않음**. 그러나 `*.exe`는 포함되어 있음.
→ spec/plan에서 말한 대로 `.gitignore`에 `*.syso` 추가가 **반드시 필요**. ✅ 정확한 기술.

### 4.2 `Scope="perMachine"` 주의

`Package.wxs`에서 `Scope="perMachine"`으로 설정되어 있으나, `Registry.wxs`에서는 모든 레지스트리가 `Root="HKCU"`에 등록됨.
- **DefaultIcon도 HKCU에 등록**하므로 일관성 있음 ✅
- 단, `perMachine` 설치인데 `HKCU` 레지스트리만 사용하면 **설치한 사용자만 아이콘이 적용**됨
- 이는 기존 설계의 의도된 동작이므로 SPEC-ICON-001의 범위가 아님 (향후 고려)

### 4.3 tray.go의 NewTray 시그니처 정확성

- spec.md A6: "기존 `assets.IconData` 기반 시스템 트레이 로직은 그대로 유지"
- 실제 코드: `NewTray(iconData []byte)` → `systray.SetIcon(t.iconData)`
- **정확히 일치** ✅. ICO 파일을 교체해도 embed 경로와 API가 동일하므로 호환성 보장됨

---

## 5. 최종 권고

### 필수 수정 (반영 필요)

| # | 문서 | 내용 | 심각도 |
|---|------|------|--------|
| 1 | spec.md 4.5 | Package.wxs 변경 불필요 → exe 경로 검증 필요 노트 추가 | Medium |
| 2 | spec.md 4.2 | `versioninfo.json` 최소 샘플 JSON 추가 | Low |
| 3 | plan.md 3 | `Invoke-Clean`에 `cmd/winmdview/resource.syso` 전체 경로 명시 | Medium |

### 권장 수정 (Optional)

| # | 문서 | 내용 |
|---|------|------|
| 4 | spec.md 3.3 | R3 컨텍스트 메뉴 아이콘이 기존 Registry.wxs의 Icon 참조로 충족됨을 명시 |
| 5 | spec.md C5 | goversioninfo 관리 방식(tool 디렉티브 vs go install) 택일 |
| 6 | plan.md 1 | 아이콘 디자인 담당/방법 명시 |
| 7 | plan.md | CI/CD(GitHub Actions) goversioninfo 설치 단계를 후속 과제로 명시 |
| 8 | acceptance.md | exe 버전 정보 자동 검증 스크립트 추가 |
| 9 | 프로젝트 | `docs/tech.md`의 systray 라이브러리 이름 수정 (`energye/systray`) |
