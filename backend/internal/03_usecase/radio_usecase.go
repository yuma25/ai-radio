package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

var ErrEmptyTheme = errors.New("program title cannot be empty")

// RadioUseCase はラジオ番組の自律進行シナリオを統括するインターフェース
type RadioUseCase interface {
	StartProgram(ctx context.Context, title string) error
	GetProgramStatus() (bool, string)
	GetRecentDialogues() []*model.Dialogue
}

type radioUseCase struct {
	commentRepo  repository.CommentRepository
	aiService    repository.AIService
	voiceService repository.VoiceService
	broadcaster  repository.StreamBroadcaster

	mu              sync.RWMutex
	currentTheme    string
	isRunning       bool
	cancelLoop      context.CancelFunc
	recentDialogues []*model.Dialogue
}

// NewRadioUseCase はRadioUseCaseを生成する
func NewRadioUseCase(
	commentRepo repository.CommentRepository,
	aiService repository.AIService,
	voiceService repository.VoiceService,
	broadcaster repository.StreamBroadcaster,
) RadioUseCase {
	return &radioUseCase{
		commentRepo:     commentRepo,
		aiService:       aiService,
		voiceService:    voiceService,
		broadcaster:     broadcaster,
		recentDialogues: make([]*model.Dialogue, 0),
	}
}

// StartProgram は指定されたタイトル・テーマで生放送を開始するシナリオ
func (uc *radioUseCase) StartProgram(ctx context.Context, title string) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ErrEmptyTheme
	}

	uc.mu.Lock()
	if uc.cancelLoop != nil {
		uc.cancelLoop() // 既に動いている番組があればリセット
	}

	loopCtx, cancel := context.WithCancel(context.Background())
	uc.cancelLoop = cancel
	uc.currentTheme = trimmed
	uc.isRunning = true
	uc.recentDialogues = make([]*model.Dialogue, 0)
	uc.mu.Unlock()

	// バックグラウンドで自律生放送ループを開始
	go uc.runBroadcastLoop(loopCtx, trimmed)

	return nil
}

func (uc *radioUseCase) GetProgramStatus() (bool, string) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.isRunning, uc.currentTheme
}

func (uc *radioUseCase) runBroadcastLoop(ctx context.Context, theme string) {
	// オープニングトーク即時実行
	_ = uc.proceedNextTalk(ctx, theme)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = uc.proceedNextTalk(ctx, theme)
		}
	}
}

func (uc *radioUseCase) proceedNextTalk(ctx context.Context, theme string) error {
	log.Printf("[RadioUseCase] proceedNextTalk called with theme: %s", theme)

	pickupComment, err := uc.commentRepo.PopNextUnprocessed(ctx)
	if err != nil {
		log.Printf("[RadioUseCase] PopNextUnprocessed error: %v", err)
		return err
	}

	uc.mu.RLock()
	dialoguesHistory := make([]*model.Dialogue, len(uc.recentDialogues))
	copy(dialoguesHistory, uc.recentDialogues)
	uc.mu.RUnlock()

	log.Printf("[RadioUseCase] Generating talk script with AI...")
	newDialogues, err := uc.aiService.GenerateRadioTalk(ctx, theme, dialoguesHistory, pickupComment)
	if err != nil {
		log.Printf("[RadioUseCase] GenerateRadioTalk failed: %v", err)
		return err
	}
	log.Printf("[RadioUseCase] Generated %d dialogues. Starting voice synthesis...", len(newDialogues))

	for i, d := range newDialogues {
		log.Printf("[RadioUseCase] Synthesizing voice %d/%d for %s: %s", i+1, len(newDialogues), d.Speaker, d.Text)
		audioBytes, err := uc.voiceService.SynthesizeVoice(ctx, d.Speaker, d.Text)
		if err != nil {
			log.Printf("[RadioUseCase] Voice synthesis error for %s: %v", d.Speaker, err)
			audioBytes = nil
		} else {
			log.Printf("[RadioUseCase] Voice synthesized (%d bytes)", len(audioBytes))
		}

		if uc.broadcaster != nil {
			log.Printf("[RadioUseCase] Broadcasting dialogue to SSE clients...")
			uc.broadcaster.BroadcastDialogue(d, audioBytes)
		}

		uc.mu.Lock()
		uc.recentDialogues = append(uc.recentDialogues, d)
		if len(uc.recentDialogues) > 20 {
			uc.recentDialogues = uc.recentDialogues[len(uc.recentDialogues)-20:]
		}
		uc.mu.Unlock()
	}

	log.Printf("[RadioUseCase] Completed talk block.")
	return nil
}

func (uc *radioUseCase) GetRecentDialogues() []*model.Dialogue {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	copied := make([]*model.Dialogue, len(uc.recentDialogues))
	copy(copied, uc.recentDialogues)
	return copied
}
