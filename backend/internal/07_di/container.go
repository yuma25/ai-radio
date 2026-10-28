package di

import (
	"github.com/yuma25/AI-Radio-Room/backend/internal/04_adapter_gateway"
	"github.com/yuma25/AI-Radio-Room/backend/internal/05_adapter_handler"
	"github.com/yuma25/AI-Radio-Room/backend/internal/06_infrastructure"
	"github.com/yuma25/AI-Radio-Room/backend/internal/03_usecase"
)

// Config はDIコンテナの組み立てに必要な外部設定値
type Config struct {
	Port         string
	VoicevoxURL  string
	GeminiAPIKey string
	GeminiModel  string
}

// Container は全レイヤの依存関係を組み立てて保持するDIコンテナ
type Container struct {
	cfg          Config
	Server       *infrastructure.Server
	RadioUseCase usecase.RadioUseCase
}

// NewContainer は内側から外側へ依存オブジェクトを注入・配線してコンテナを生成する
func NewContainer(cfg Config) *Container {
	// 【第4層: Adapter Output / Gateway】外部通信・リポジトリ実装の生成
	commentRepo := gateway.NewMemoryCommentRepository()
	voiceService := gateway.NewVoicevoxGateway(cfg.VoicevoxURL)
	aiService := gateway.NewGeminiGateway(cfg.GeminiAPIKey, cfg.GeminiModel)
	sseBroadcaster := gateway.NewSSEBroadcaster()

	// 【第3層: UseCase】業務シナリオの生成
	commentUseCase := usecase.NewCommentUseCase(commentRepo, sseBroadcaster)
	radioUseCase := usecase.NewRadioUseCase(commentRepo, aiService, voiceService, sseBroadcaster)

	// 【第5層: Adapter Input / Handler】HTTP受付窓口の生成
	radioHandler := handler.NewRadioHandler(radioUseCase)
	commentHandler := handler.NewCommentHandler(commentUseCase)
	healthHandler := handler.NewHealthHandler()

	// 【第6層: Infrastructure】ルーター・サーバーの生成
	router := infrastructure.NewRouter(infrastructure.RouterConfig{
		RadioHandler:   radioHandler,
		CommentHandler: commentHandler,
		HealthHandler:  healthHandler,
		SSEBroadcaster: sseBroadcaster,
	})

	server := infrastructure.NewServer(cfg.Port, router)

	return &Container{
		cfg:          cfg,
		Server:       server,
		RadioUseCase: radioUseCase,
	}
}
