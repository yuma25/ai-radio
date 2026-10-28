package infrastructure

import (
	"net/http"

	"github.com/yuma25/AI-Radio-Room/backend/internal/04_adapter_gateway"
	"github.com/yuma25/AI-Radio-Room/backend/internal/05_adapter_handler"
)

// RouterConfig はルーティング設定に必要なハンドラ・Gateway群
type RouterConfig struct {
	RadioHandler   *handler.RadioHandler
	CommentHandler *handler.CommentHandler
	HealthHandler  *handler.HealthHandler
	SSEBroadcaster gateway.SSEBroadcaster
}

// NewRouter はHTTPリクエストのルーティングとCORSミドルウェアを設定したHandlerを生成する
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// ヘルスチェック
	mux.HandleFunc("/health", cfg.HealthHandler.HandleHealth)

	// ラジオ番組制御 API
	mux.HandleFunc("/api/radio/start", cfg.RadioHandler.HandleStart)
	mux.HandleFunc("/api/radio/status", cfg.RadioHandler.HandleStatus)

	// コメント API
	mux.HandleFunc("/api/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			cfg.CommentHandler.HandlePostComment(w, r)
		case http.MethodGet:
			cfg.CommentHandler.HandleGetComments(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// SSE ストリーミング配信
	mux.HandleFunc("/api/radio/stream", cfg.SSEBroadcaster.ServeHTTP)

	// CORS ミドルウェア適用
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
