package repository

import (
	"context"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
)

// CommentRepository はリスナーお便りの保存・取得契約
type CommentRepository interface {
	Save(ctx context.Context, comment *model.Comment) error
	FindAll(ctx context.Context) ([]*model.Comment, error)
	PopNextUnprocessed(ctx context.Context) (*model.Comment, error)
}

// StreamBroadcaster はリスナーへのリアルタイムストリーミング配信契約
type StreamBroadcaster interface {
	BroadcastDialogue(dialogue *model.Dialogue, audio []byte)
	BroadcastComment(comment *model.Comment)
}

// AIService は対話AIによるラジオトーク原稿生成契約
type AIService interface {
	GenerateRadioTalk(ctx context.Context, theme string, recentDialogues []*model.Dialogue, pickupComment *model.Comment) ([]*model.Dialogue, error)
}

// VoiceService は音声合成エンジンによる音声バイナリ生成契約
type VoiceService interface {
	SynthesizeVoice(ctx context.Context, speaker model.Speaker, text string) ([]byte, error)
}
