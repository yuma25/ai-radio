package gateway

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

// SSEBroadcaster はSSEクライアントへのイベント配信Gateway
type SSEBroadcaster interface {
	repository.StreamBroadcaster
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type clientChan chan []byte

type sseBroadcaster struct {
	mu      sync.RWMutex
	clients map[clientChan]struct{}
}

// NewSSEBroadcaster はSSEBroadcasterを生成する
func NewSSEBroadcaster() SSEBroadcaster {
	return &sseBroadcaster{
		clients: make(map[clientChan]struct{}),
	}
}

// イベント送信用DTO
type sseEvent struct {
	Type      string      `json:"type"` // "dialogue" or "comment"
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type dialogueEventData struct {
	Speaker     string `json:"speaker"`
	Text        string `json:"text"`
	AudioBase64 string `json:"audioBase64,omitempty"`
}

type commentEventData struct {
	UserName string `json:"userName"`
	Content  string `json:"content"`
}

func (b *sseBroadcaster) BroadcastDialogue(dialogue *model.Dialogue, audio []byte) {
	var audioB64 string
	if len(audio) > 0 {
		audioB64 = base64.StdEncoding.EncodeToString(audio)
	}

	payload := sseEvent{
		Type: "dialogue",
		Data: dialogueEventData{
			Speaker:     string(dialogue.Speaker),
			Text:        dialogue.Text,
			AudioBase64: audioB64,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	b.sendToAll(payload)
}

func (b *sseBroadcaster) BroadcastComment(comment *model.Comment) {
	payload := sseEvent{
		Type: "comment",
		Data: commentEventData{
			UserName: comment.UserName,
			Content:  comment.Content,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	b.sendToAll(payload)
}

func (b *sseBroadcaster) sendToAll(event sseEvent) {
	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	msg := fmt.Sprintf("data: %s\n\n", string(bytes))

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- []byte(msg):
		default:
			// クライアントバッファ詰まり時はスキップ
		}
	}
}

func (b *sseBroadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(clientChan, 16)

	b.mu.Lock()
	b.clients[client] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, client)
		close(client)
		b.mu.Unlock()
	}()

	// 接続確立の初期イベント送信
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-client:
			if !ok {
				return
			}
			_, _ = w.Write(msg)
			flusher.Flush()
		}
	}
}
