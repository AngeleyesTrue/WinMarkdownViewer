// Package server 는 WebSocket 기반의 실시간 HTML 콘텐츠 제공 서버를 구현한다.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"sync"

	"github.com/AngeleyesTrue/WinMarkdownViewer/web"
	"github.com/gorilla/websocket"
)

// maxPortRetries 는 포트 바인딩 재시도 횟수이다.
const maxPortRetries = 3

// wsMessage 는 WebSocket을 통해 클라이언트로 전송되는 메시지 구조체이다.
type wsMessage struct {
	Type string `json:"type"`
	HTML string `json:"html"`
}

// templateData 는 viewer.html 템플릿에 전달할 데이터이다.
type templateData struct {
	Title    string
	CSS      template.CSS
	FontSize int
	Content  template.HTML
}

// viewerTmpl 은 모듈 초기화 시 파싱되는 HTML 뷰어 템플릿이다.
var viewerTmpl = template.Must(template.New("viewer").Parse(string(web.ViewerHTML)))

// Server 는 HTTP와 WebSocket을 제공하는 서버이다.
type Server struct {
	mu       sync.RWMutex
	title    string
	content  string
	fontSize int
	clients  map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
	server   *http.Server
	listener net.Listener
}

// NewServer 는 새로운 Server 인스턴스를 생성한다.
func NewServer() (*Server, error) {
	s := &Server{
		fontSize: 16,
		clients:  make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			// 로컬호스트에서만 사용하므로 모든 오리진을 허용한다
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return s, nil
}

// Start 는 서버를 127.0.0.1의 랜덤 포트에서 시작하고 포트 번호를 반환한다.
// 포트 바인딩에 실패하면 최대 3회 재시도한다.
func (s *Server) Start() (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/ws", s.handleWebSocket)

	// 정적 파일 서빙: /static/ 경로로 임베디드 JS, CSS, 폰트 파일을 제공한다
	// @MX:NOTE: [AUTO] embed.FS의 루트(".")에서 Sub를 생성하면 에러가 발생하지 않는다
	staticFS, _ := fs.Sub(web.ExtensionAssets, ".")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	s.server = &http.Server{
		Handler: mux,
	}

	var lastErr error
	for i := 0; i < maxPortRetries; i++ {
		// 127.0.0.1:0 으로 바인딩하여 랜덤 포트를 할당받는다
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			lastErr = err
			continue
		}

		s.listener = listener
		port := listener.Addr().(*net.TCPAddr).Port

		go func() {
			if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
				// 서버 에러를 무시한다 (Shutdown에 의한 종료)
			}
		}()

		return port, nil
	}

	return 0, fmt.Errorf("포트 바인딩 실패 (%d회 재시도): %w", maxPortRetries, lastErr)
}

// Shutdown 은 서버를 정상적으로 종료한다.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	// 모든 WebSocket 클라이언트 연결을 닫는다
	for conn := range s.clients {
		conn.Close()
		delete(s.clients, conn)
	}
	s.mu.Unlock()

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// SetTitle 은 페이지 제목을 설정한다.
func (s *Server) SetTitle(title string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.title = title
}

// SetFontSize 는 본문 폰트 크기를 설정한다.
func (s *Server) SetFontSize(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fontSize = size
}

// SetContent 는 현재 HTML 콘텐츠를 설정한다.
func (s *Server) SetContent(html string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.content = html
}

// Broadcast 는 모든 연결된 WebSocket 클라이언트에 HTML 콘텐츠를 전송한다.
func (s *Server) Broadcast(html string) {
	s.mu.Lock()
	s.content = html
	s.mu.Unlock()

	msg := wsMessage{
		Type: "update",
		HTML: html,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			continue
		}
	}
}

// handleRoot 는 GET / 요청을 처리하여 viewer.html 템플릿을 렌더링한다.
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	data := templateData{
		Title:    s.title,
		CSS:      template.CSS(web.GitHubMarkdownCSS),
		FontSize: s.fontSize,
		Content:  template.HTML(s.content),
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer
	if err := viewerTmpl.Execute(&buf, data); err != nil {
		http.Error(w, "템플릿 렌더링 실패", http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

// handleWebSocket 은 WebSocket 업그레이드 요청을 처리한다.
// 연결 즉시 현재 콘텐츠를 전송하고, 클라이언트를 허브에 등록한다.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// 클라이언트를 등록한다
	s.mu.Lock()
	s.clients[conn] = struct{}{}
	content := s.content
	s.mu.Unlock()

	// 현재 콘텐츠를 즉시 전송한다
	msg := wsMessage{
		Type: "update",
		HTML: content,
	}
	if msgBytes, err := json.Marshal(msg); err == nil {
		conn.WriteMessage(websocket.TextMessage, msgBytes)
	}

	// 클라이언트의 메시지를 읽는 고루틴 (연결 유지 및 종료 감지)
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			conn.Close()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}
