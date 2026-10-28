package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

type voicevoxGateway struct {
	baseURL    string
	httpClient *http.Client
}

// NewVoicevoxGateway はVoiceServiceのアダプタ実装を生成する
func NewVoicevoxGateway(baseURL string) repository.VoiceService {
	return &voicevoxGateway{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *voicevoxGateway) SynthesizeVoice(ctx context.Context, speaker model.Speaker, text string) ([]byte, error) {
	styleID := 3 // デフォルト: ずんだもん (ノーマル: 3)
	if speaker == model.SpeakerMetan {
		styleID = 2 // 四国めたん (ノーマル: 2)
	}

	// 1. 音声合成クエリ作成 (/audio_query)
	queryURL := fmt.Sprintf("%s/audio_query?speaker=%s&text=%s",
		g.baseURL,
		strconv.Itoa(styleID),
		url.QueryEscape(text),
	)

	reqQuery, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create audio_query request failed: %w", err)
	}

	resQuery, err := g.httpClient.Do(reqQuery)
	if err != nil {
		return nil, fmt.Errorf("audio_query request failed: %w", err)
	}
	defer resQuery.Body.Close()

	if resQuery.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resQuery.Body)
		return nil, fmt.Errorf("audio_query error: status=%d body=%s", resQuery.StatusCode, string(body))
	}

	queryJSON, err := io.ReadAll(resQuery.Body)
	if err != nil {
		return nil, fmt.Errorf("read audio_query response failed: %w", err)
	}

	// 2. 音声合成 (/synthesis)
	synthesisURL := fmt.Sprintf("%s/synthesis?speaker=%s", g.baseURL, strconv.Itoa(styleID))
	reqSynth, err := http.NewRequestWithContext(ctx, http.MethodPost, synthesisURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("create synthesis request failed: %w", err)
	}
	reqSynth.Header.Set("Content-Type", "application/json")

	resSynth, err := g.httpClient.Do(reqSynth)
	if err != nil {
		return nil, fmt.Errorf("synthesis request failed: %w", err)
	}
	defer resSynth.Body.Close()

	if resSynth.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resSynth.Body)
		return nil, fmt.Errorf("synthesis error: status=%d body=%s", resSynth.StatusCode, string(body))
	}

	audioBytes, err := io.ReadAll(resSynth.Body)
	if err != nil {
		return nil, fmt.Errorf("read synthesis audio failed: %w", err)
	}

	return audioBytes, nil
}
