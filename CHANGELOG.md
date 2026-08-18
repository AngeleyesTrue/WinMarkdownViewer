# 변경 이력

이 프로젝트의 모든 주요 변경 사항을 기록합니다.

형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따르며,
버전 관리는 [Semantic Versioning](https://semver.org/lang/ko/)을 따릅니다.

## [1.0.4] - 2026-03-10

### 추가 (애플리케이션 아이콘 개선)

- 멀티사이즈 아이콘 적용 (16, 32, 48, 64, 128, 256px)
- goversioninfo로 exe에 Windows 리소스 아이콘 임베딩
- WebView2 윈도우 타이틀바에 커스텀 아이콘 설정 (WM_SETICON)
- .md 파일 연결 시 앱 아이콘 표시 (DefaultIcon 레지스트리)
- 빌드 파이프라인에 goversioninfo 단계 통합 (build.ps1, build-msi.ps1)
- 멀티사이즈 ICO 검증 테스트 추가

## [미출시]

### 추가 (Markdown Viewer MVP)

- Go 모듈 초기화 (`github.com/AngeleyesTrue/WinMarkdownViewer`)
- goldmark 기반 Markdown 렌더링 엔진 (`internal/markdown/renderer.go`)
  - GFM 확장 지원 (테이블, 취소선, 자동 링크, 태스크 리스트)
  - chroma 기반 코드 블록 구문 강조
- WebView2 뷰어 윈도우 관리 (`internal/viewer/viewer.go`)
  - go-webview2 순수 Go 바인딩 사용 (CGO 불필요)
  - 윈도우 타이틀에 파일명 표시
  - 기본 크기 1024x768
- CLI 파싱 및 파일 검증 파이프라인 (`internal/app/app.go`)
- 오류 처리
  - 파일 없음 (exit code 1)
  - 파일 읽기 권한 오류
  - WebView2 Runtime 미설치 안내
  - 빈 Markdown 파일 안내 메시지
- HTML 템플릿 및 GitHub 스타일 CSS (`web/`)
  - CSP(Content Security Policy) 헤더로 XSS 방지
  - go:embed로 모든 정적 리소스 임베딩 (외부 네트워크 요청 없음)
- 테스트 37개 작성 및 통과
