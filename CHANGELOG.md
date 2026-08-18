# 변경 이력

이 프로젝트의 모든 주요 변경 사항을 기록합니다.

형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따르며,
버전 관리는 [Semantic Versioning](https://semver.org/lang/ko/)을 따릅니다.

## [1.0.0] - 2026-08-18

### 추가

- **Markdown 렌더링**: GitHub Flavored Markdown(테이블, 취소선, 자동 링크, 태스크 리스트) 지원, 코드 블록 구문 강조, GitHub 스타일 CSS
- **실시간 미리보기**: 파일 저장 시 자동으로 변경 내용 반영, 스크롤 위치 유지로 끊김 없는 편집 경험
- **다크 모드**: 라이트/다크/시스템 세 가지 테마, 토글 버튼 및 `Ctrl+Shift+D` 단축키
- **사용자 설정**: 테마·폰트 크기·창 크기를 JSON 설정 파일로 저장
- **수학 수식**: KaTeX 기반 인라인(`$...$`) 및 블록(`$$...$$`) 수식 렌더링
- **다이어그램**: Mermaid 기반 flowchart, sequence, class, state, gantt, pie 다이어그램 렌더링
- **Windows 통합**: .md 파일 우클릭 컨텍스트 메뉴, 시스템 트레이 최소화, 단일 인스턴스 실행(Named Pipe로 새 파일 전달)
- **멀티 윈도우**: 여러 .md 파일을 각각 독립된 윈도우에서 동시에 열기(최대 10개)
- **애플리케이션 아이콘**: 멀티사이즈 아이콘(16~256px) — exe 파일, 타이틀바, 파일 연결, 인스톨러 전 영역 적용
- **오류 처리**: 파일 없음, 읽기 권한 오류, WebView2 Runtime 미설치, 빈 파일 안내
- **보안**: CSP 헤더로 XSS 방지, localhost 전용 서버 바인딩
- **배포**: MSI 인스톨러(WiX), PowerShell 원라이너 포터블 설치(`irm ... | iex`)
- 순수 Go로 작성 — CGO 불필요
