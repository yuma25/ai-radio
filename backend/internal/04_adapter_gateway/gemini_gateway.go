package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

type geminiGateway struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiGateway はAIServiceのアダプタ実装を生成する
func NewGeminiGateway(apiKey, modelName string) repository.AIService {
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}
	return &geminiGateway{
		apiKey:     apiKey,
		model:      modelName,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// --- Gemini API リクエスト / レスポンス内部定義 (Adapter層に隔離) ---

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type schemaProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type schemaItems struct {
	Type       string                    `json:"type"`
	Properties map[string]schemaProperty `json:"properties"`
	Required   []string                  `json:"required"`
}

type responseSchemaDef struct {
	Type  string      `json:"type"`
	Items schemaItems `json:"items"`
}

type generationConfig struct {
	ResponseMimeType string             `json:"responseMimeType,omitempty"`
	ResponseSchema   *responseSchemaDef `json:"responseSchema,omitempty"`
	Temperature      float64            `json:"temperature,omitempty"`
	TopP             float64            `json:"topP,omitempty"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent    `json:"system_instruction,omitempty"`
	Contents          []geminiContent   `json:"contents"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

type dialogueJSON struct {
	Speaker string `json:"speaker"` // "zundamon" または "metan"
	Text    string `json:"text"`
}

// GenerateRadioTalk はGemini APIを呼び出してラジオトーク台本を生成し、ドメインモデルへマッピングする
func (g *geminiGateway) GenerateRadioTalk(
	ctx context.Context,
	theme string,
	recentDialogues []*model.Dialogue,
	pickupComment *model.Comment,
) ([]*model.Dialogue, error) {
	// 1. システム規範（ペルソナ・役割分担・ルール）の構築
	systemInstruction := buildSystemInstruction()

	// 2. 進行コンテキスト（基本テーマ、直前の会話履歴、お便り）の構築
	userPrompt := buildUserPrompt(theme, recentDialogues, pickupComment)

	// 3. 型安全な構造化出力スキーマの定義 (Structured Outputs)
	structuredSchema := &responseSchemaDef{
		Type: "ARRAY",
		Items: schemaItems{
			Type: "OBJECT",
			Properties: map[string]schemaProperty{
				"speaker": {
					Type:        "STRING",
					Description: "発話者（zundamon または metan）",
					Enum:        []string{"zundamon", "metan"},
				},
				"text": {
					Type:        "STRING",
					Description: "パーソナリティの発話セリフ（20〜60文字程度で読みやすい日本語）",
				},
			},
			Required: []string{"speaker", "text"},
		},
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: systemInstruction},
			},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: userPrompt},
				},
			},
		},
		GenerationConfig: &generationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   structuredSchema,
			Temperature:      0.75, // 自然な雑談とユーモアのための創造性
			TopP:             0.95, // 破綻単語の抑制
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request failed: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model,
		g.apiKey,
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("create gemini request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpRes, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpRes.Body)
		return nil, fmt.Errorf("gemini api error: status=%d body=%s", httpRes.StatusCode, string(body))
	}

	var geminiRes geminiResponse
	if err := json.NewDecoder(httpRes.Body).Decode(&geminiRes); err != nil {
		return nil, fmt.Errorf("decode gemini response failed: %w", err)
	}

	if len(geminiRes.Candidates) == 0 || len(geminiRes.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	rawJSON := geminiRes.Candidates[0].Content.Parts[0].Text

	// Markdown コードブロックのトリム処理
	cleanedJSON := strings.TrimSpace(rawJSON)
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```json")
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```")
	cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")
	cleanedJSON = strings.TrimSpace(cleanedJSON)

	var rawDialogues []dialogueJSON
	if err := json.Unmarshal([]byte(cleanedJSON), &rawDialogues); err != nil {
		return nil, fmt.Errorf("unmarshal dialogues json failed: %w, rawText: %s", err, rawJSON)
	}

	// 4. ドメインモデルへのマッピング
	dialogues := make([]*model.Dialogue, 0, len(rawDialogues))
	for _, item := range rawDialogues {
		var spk model.Speaker
		if strings.ToLower(item.Speaker) == "metan" {
			spk = model.SpeakerMetan
		} else {
			spk = model.SpeakerZundamon
		}

		d, err := model.NewDialogue(spk, item.Text)
		if err != nil {
			continue // ドメイン不変条件に反するセリフは安全に除外
		}
		dialogues = append(dialogues, d)
	}

	return dialogues, nil
}

// buildSystemInstruction は両者の明確なペルソナと掛け合いルールを定義する
func buildSystemInstruction() string {
	return `あなたは双方向AIラジオ番組「AI Radio Room」のパーソナリティである「ずんだもん」と「四国めたん」の掛け合い台本を執筆する構成作家AIです。

【ずんだもんのペルソナ仕様】
- 一人称:「ボク」
- 語尾:「〜のだ」「〜なのだ」（語尾を省略せず自然に可愛らしく使う）
- 性格: 明るく好奇心旺盛、お調子者で素直。ボケ役。お便りや新しい話題にすぐ食いついて興奮する。

【四国めたんのペルソナ仕様】
- 一人称:「わたくし」
- 語尾:「〜かしら」「〜ですわ」「〜わね」のお嬢様口調
- 性格: 落ち着いていて知的、上品で面倒見が良い。ツッコミ＆進行役。ずんだもんの脱線や暴走を優しくたしなめつつ話を前に進める。

【掛け合い＆演出の鉄則ルール】
1. 【交互発言】ずんだもんとめたんが交互に発言し、テンポの良い会話のラリー（4〜6発話）を展開してください。
2. 【お便り最優先の深掘り】リスナーのお便りがある場合、必ず「ラジオネーム」を呼び、お便りの内容に深く共感・ツッコミ・感想を交えて盛り上げてください。
3. 【発話の長さ】1発話あたり20〜60文字程度とし、音声合成（VOICEVOX）で耳から聞いて自然にスッと入る長さに収めてください。
4. 【余計な解説の排除】Markdownコードフェンスや解説文は一切出力せず、指定されたJSON配列のみを出力してください。`
}

// buildUserPrompt は現在の会話文脈とお便り情報を構築する
func buildUserPrompt(theme string, history []*model.Dialogue, comment *model.Comment) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("【番組テーマ】: %s\n\n", theme))

	if len(history) > 0 {
		sb.WriteString("【直前の会話の流れ】\n")
		// 直近最大6件の履歴を提示して文脈の連続性を保つ
		start := 0
		if len(history) > 6 {
			start = len(history) - 6
		}
		for _, h := range history[start:] {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", h.Speaker, h.Text))
		}
		sb.WriteString("\n")
	}

	if comment != nil {
		sb.WriteString("【★新着リスナーお便り（必ずこのお便りを取り上げて2人でトークしてください）】\n")
		sb.WriteString(fmt.Sprintf("- ラジオネーム: %s さん\n", comment.UserName))
		sb.WriteString(fmt.Sprintf("- お便り本文: %s\n", comment.Content))
		sb.WriteString("指示: めたんがお便りを読み上げ、ずんだもんがリアクションし、2人でお便りのテーマを広げる掛け合いを生成してください。\n")
	} else {
		sb.WriteString("指示: 直前の会話やテーマをさらに深掘りし、2人の個性が光る楽しいトークを続けてください。\n")
	}

	return sb.String()
}
