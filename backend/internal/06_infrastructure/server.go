package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server はHTTPサーバーの物理起動とグレースフルシャットダウンを管理する
type Server struct {
	httpServer *http.Server
}

// NewServer はServerインスタンスを初期化する
func NewServer(port string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 0, // SSE ストリーミング維持のため無制限
			IdleTimeout:  120 * time.Second,
		},
	}
}

// Start はサーバーを非同期起動する
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown は安全にサーバーを停止する
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return fmt.Sprintf("http://localhost%s", s.httpServer.Addr)
}
