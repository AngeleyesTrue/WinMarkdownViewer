package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestNewServer_정상생성 Server를 정상적으로 생성할 수 있는지 검증한다.
func TestNewServer_정상생성(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())
}

// TestStart_랜덤포트 Start가 랜덤 포트를 할당하고 포트 번호를 반환하는지 검증한다.
func TestStart_랜덤포트(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	if port <= 0 || port > 65535 {
		t.Errorf("유효하지 않은 포트 번호: %d", port)
	}
}

// TestStart_로컬호스트바인딩 서버가 127.0.0.1에만 바인딩되는지 검증한다.
func TestStart_로컬호스트바인딩(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	// 127.0.0.1로 연결 가능해야 한다
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err != nil {
		t.Fatalf("127.0.0.1:%d 연결 실패: %v", port, err)
	}
	conn.Close()
}

// TestShutdown_정상종료 Shutdown이 정상적으로 서버를 종료하는지 검증한다.
func TestShutdown_정상종료(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}

	_, err = s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() 오류: %v", err)
	}
}

// TestSetContent_HTML설정 SetContent로 HTML 콘텐츠를 설정하고 GET /로 제공하는지 검증한다.
func TestSetContent_HTML설정(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	testHTML := "<h1>테스트 콘텐츠</h1>"
	s.SetContent(testHTML)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err != nil {
		t.Fatalf("HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("응답 본문 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if !strings.Contains(string(body), testHTML) {
		t.Errorf("응답에 설정된 HTML이 포함되지 않음: %s", string(body))
	}
}

// TestWebSocket_핸드셰이크 WebSocket 연결이 성공적으로 수립되는지 검증한다.
func TestWebSocket_핸드셰이크(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket 연결 실패: %v", err)
	}
	defer conn.Close()
}

// TestWebSocket_초기콘텐츠전송 WebSocket 연결 시 현재 콘텐츠를 즉시 전송하는지 검증한다.
func TestWebSocket_초기콘텐츠전송(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	initialHTML := "<p>초기 콘텐츠</p>"
	s.SetContent(initialHTML)

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket 연결 실패: %v", err)
	}
	defer conn.Close()

	// 연결 즉시 초기 콘텐츠를 수신해야 한다
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("WebSocket 메시지 수신 실패: %v", err)
	}

	var msg wsMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		t.Fatalf("메시지 JSON 파싱 실패: %v", err)
	}

	if msg.Type != "update" {
		t.Errorf("메시지 타입 = %q, want %q", msg.Type, "update")
	}

	if msg.HTML != initialHTML {
		t.Errorf("초기 콘텐츠 = %q, want %q", msg.HTML, initialHTML)
	}
}

// TestBroadcast_모든클라이언트전송 Broadcast가 모든 연결된 WebSocket 클라이언트에 메시지를 전송하는지 검증한다.
func TestBroadcast_모든클라이언트전송(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	// 초기 콘텐츠를 설정한다
	s.SetContent("<p>초기</p>")

	const clientCount = 3
	conns := make([]*websocket.Conn, clientCount)

	// 여러 클라이언트를 연결한다
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)
	for i := 0; i < clientCount; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("클라이언트 %d WebSocket 연결 실패: %v", i, err)
		}
		defer conn.Close()
		conns[i] = conn
	}

	// 각 클라이언트의 초기 콘텐츠 메시지를 소비한다
	for i, conn := range conns {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("클라이언트 %d 초기 메시지 수신 실패: %v", i, err)
		}
	}

	// 브로드캐스트를 실행한다
	broadcastHTML := "<h1>브로드캐스트 메시지</h1>"
	s.Broadcast(broadcastHTML)

	// 모든 클라이언트가 브로드캐스트 메시지를 수신했는지 검증한다
	var wg sync.WaitGroup
	errors := make([]error, clientCount)

	for i, conn := range conns {
		wg.Add(1)
		go func(idx int, c *websocket.Conn) {
			defer wg.Done()
			c.SetReadDeadline(time.Now().Add(3 * time.Second))
			_, msgBytes, err := c.ReadMessage()
			if err != nil {
				errors[idx] = fmt.Errorf("클라이언트 %d 메시지 수신 실패: %v", idx, err)
				return
			}

			var msg wsMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				errors[idx] = fmt.Errorf("클라이언트 %d JSON 파싱 실패: %v", idx, err)
				return
			}

			if msg.Type != "update" || msg.HTML != broadcastHTML {
				errors[idx] = fmt.Errorf("클라이언트 %d: type=%q html=%q, want type=update html=%q",
					idx, msg.Type, msg.HTML, broadcastHTML)
			}
		}(i, conn)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

// TestStart_포트재시도 포트 바인딩 실패 시 재시도하는지 검증한다.
func TestStart_포트재시도(t *testing.T) {
	t.Parallel()

	// 서버가 정상적으로 포트를 찾아 시작할 수 있는지만 검증한다
	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	if port <= 0 {
		t.Errorf("유효한 포트가 할당되어야 한다, got: %d", port)
	}
}

// TestGetRoot_HTML페이지 GET / 엔드포인트가 viewer.html 템플릿을 렌더링하는지 검증한다.
func TestGetRoot_HTML페이지(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	s.SetTitle("테스트 타이틀")
	s.SetFontSize(18)
	s.SetContent("<p>테스트</p>")

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err != nil {
		t.Fatalf("HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("응답 본문 읽기 실패: %v", err)
	}

	bodyStr := string(body)

	// viewer.html 템플릿이 렌더링되어야 한다
	checks := []string{
		"<!DOCTYPE html>",
		"테스트 타이틀",
		"18px",
		"<p>테스트</p>",
		"markdown-body",
		"WebSocket",
		"connect()",
	}

	for _, want := range checks {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("응답에 %q 가 포함되지 않음", want)
		}
	}
}

// TestSetTitle_제목설정 SetTitle로 페이지 제목을 설정하는지 검증한다.
func TestSetTitle_제목설정(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}

	s.SetTitle("My Title")

	s.mu.RLock()
	title := s.title
	s.mu.RUnlock()

	if title != "My Title" {
		t.Errorf("title = %q, want %q", title, "My Title")
	}
}

// TestSetFontSize_폰트크기설정 SetFontSize로 폰트 크기를 설정하는지 검증한다.
func TestSetFontSize_폰트크기설정(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}

	s.SetFontSize(20)

	s.mu.RLock()
	fontSize := s.fontSize
	s.mu.RUnlock()

	if fontSize != 20 {
		t.Errorf("fontSize = %d, want %d", fontSize, 20)
	}
}

// TestSetTheme_테마설정 SetTheme으로 테마를 설정하는지 검증한다.
func TestSetTheme_테마설정(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}

	tests := []struct {
		name  string
		theme string
	}{
		{"라이트 테마", "light"},
		{"다크 테마", "dark"},
		{"시스템 테마", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetTheme(tt.theme)

			s.mu.RLock()
			theme := s.theme
			s.mu.RUnlock()

			if theme != tt.theme {
				t.Errorf("theme = %q, want %q", theme, tt.theme)
			}
		})
	}
}

// TestSetTheme_기본값 테마 미설정 시 handleRoot에서 기본값(system)이 사용되는지 검증한다.
func TestSetTheme_기본값(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	// 테마를 설정하지 않고 요청
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err != nil {
		t.Fatalf("HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("응답 본문 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// HTML 문서가 정상적으로 렌더링되어야 한다
	if !strings.Contains(string(body), "<!DOCTYPE html>") {
		t.Error("기본 테마에서 HTML 문서가 정상적으로 렌더링되어야 한다")
	}
}

// TestStaticFileServing_JS정적파일 /static/ 경로로 JS 파일이 서빙되는지 검증한다.
func TestStaticFileServing_JS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	// /static/js/ 경로에 접근 가능해야 한다
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/js/render-extensions.js", port))
	if err != nil {
		t.Fatalf("정적 파일 HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("정적 JS 파일 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("정적 파일 응답 읽기 실패: %v", err)
	}

	if len(body) == 0 {
		t.Error("정적 JS 파일 응답이 비어있음")
	}
}

// TestStaticFileServing_CSS정적파일 /static/ 경로로 CSS 파일이 서빙되는지 검증한다.
func TestStaticFileServing_CSS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/css/github-markdown.css", port))
	if err != nil {
		t.Fatalf("정적 CSS 파일 HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("정적 CSS 파일 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestStaticFileServing_KaTeXJS정적파일 KaTeX JS 파일이 정적으로 서빙되는지 검증한다.
func TestStaticFileServing_KaTeXJS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/js/katex.min.js", port))
	if err != nil {
		t.Fatalf("KaTeX JS HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("KaTeX JS 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestStaticFileServing_MermaidJS정적파일 Mermaid JS 파일이 정적으로 서빙되는지 검증한다.
func TestStaticFileServing_MermaidJS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/js/mermaid.min.js", port))
	if err != nil {
		t.Fatalf("Mermaid JS HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Mermaid JS 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestStaticFileServing_KaTeXCSS정적파일 KaTeX CSS 파일이 정적으로 서빙되는지 검증한다.
func TestStaticFileServing_KaTeXCSS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/css/katex.min.css", port))
	if err != nil {
		t.Fatalf("KaTeX CSS HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("KaTeX CSS 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("KaTeX CSS 응답 읽기 실패: %v", err)
	}

	// CSS 내용에 폰트 경로가 포함되어 있는지 확인
	if !strings.Contains(string(body), "../fonts/") {
		t.Error("KaTeX CSS에 수정된 폰트 경로(../fonts/)가 포함되지 않음")
	}
}

// TestStaticFileServing_KaTeXFont정적파일 KaTeX 폰트 파일이 정적으로 서빙되는지 검증한다.
func TestStaticFileServing_KaTeXFont정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	// 대표 폰트 파일 1개 확인
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/fonts/KaTeX_Main-Regular.woff2", port))
	if err != nil {
		t.Fatalf("KaTeX 폰트 HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("KaTeX 폰트 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestStaticFileServing_SyntaxLightCSS정적파일 구문 강조 라이트 테마 CSS가 서빙되는지 검증한다.
func TestStaticFileServing_SyntaxLightCSS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/css/syntax-light.css", port))
	if err != nil {
		t.Fatalf("syntax-light.css HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("syntax-light.css 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("syntax-light.css 응답 읽기 실패: %v", err)
	}

	// 라이트 테마 CSS에 chroma 클래스가 포함되어야 한다
	if !strings.Contains(string(body), ".chroma") {
		t.Error("syntax-light.css에 .chroma 클래스가 포함되지 않음")
	}
}

// TestStaticFileServing_SyntaxDarkCSS정적파일 구문 강조 다크 테마 CSS가 서빙되는지 검증한다.
func TestStaticFileServing_SyntaxDarkCSS정적파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/css/syntax-dark.css", port))
	if err != nil {
		t.Fatalf("syntax-dark.css HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("syntax-dark.css 상태 코드 = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("syntax-dark.css 응답 읽기 실패: %v", err)
	}

	// 다크 테마 CSS에 chroma 클래스가 포함되어야 한다
	if !strings.Contains(string(body), ".chroma") {
		t.Error("syntax-dark.css에 .chroma 클래스가 포함되지 않음")
	}

	// 다크 테마에는 prefers-color-scheme 미디어 쿼리가 포함되어야 한다
	if !strings.Contains(string(body), "prefers-color-scheme: dark") {
		t.Error("syntax-dark.css에 prefers-color-scheme 미디어 쿼리가 포함되지 않음")
	}
}

// TestStaticFileServing_존재하지않는파일 존재하지 않는 정적 파일에 대해 404를 반환하는지 검증한다.
func TestStaticFileServing_존재하지않는파일(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/static/nonexistent.file", port))
	if err != nil {
		t.Fatalf("HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("존재하지 않는 파일 상태 코드 = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// TestGetRoot_확장렌더링통합 GET / 응답에 확장 렌더링 관련 태그가 포함되는지 검증한다.
func TestGetRoot_확장렌더링통합(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}
	defer s.Shutdown(context.Background())

	port, err := s.Start()
	if err != nil {
		t.Fatalf("Start() 오류: %v", err)
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err != nil {
		t.Fatalf("HTTP GET 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("응답 본문 읽기 실패: %v", err)
	}

	bodyStr := string(body)

	// 확장 렌더링 관련 태그 확인
	checks := []string{
		"katex.min.css",
		"katex.min.js",
		"mermaid.min.js",
		"render-extensions.js",
		"renderExtensions",
	}

	for _, want := range checks {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("응답에 %q 가 포함되지 않음", want)
		}
	}
}

// TestShutdown_서버미시작 서버가 시작되지 않은 상태에서 Shutdown을 호출해도 안전한지 검증한다.
func TestShutdown_서버미시작(t *testing.T) {
	t.Parallel()

	s, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() 오류: %v", err)
	}

	// 서버를 시작하지 않고 Shutdown 호출
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("미시작 서버 Shutdown 오류: %v", err)
	}
}
