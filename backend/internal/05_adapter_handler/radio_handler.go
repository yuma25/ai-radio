package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yuma25/AI-Radio-Room/backend/internal/03_usecase"
)

// RadioHandler はラジオ番組の開始・状態制御に関するHTTPハンドラ
type RadioHandler struct {
	radioUseCase usecase.RadioUseCase
}

func NewRadioHandler(radioUseCase usecase.RadioUseCase) *RadioHandler {
	return &RadioHandler{
		radioUseCase: radioUseCase,
	}
}

type startRadioRequest struct {
	Title string `json:"title"`
}

type radioStatusResponse struct {
	IsRunning bool   `json:"is_running"`
	Theme     string `json:"theme"`
}

// HandleStart は POST /api/radio/start を処理する
func (h *RadioHandler) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req startRadioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.radioUseCase.StartProgram(r.Context(), req.Title); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"theme":   req.Title,
	})
}

// HandleStatus は GET /api/radio/status を処理する
func (h *RadioHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	isRunning, theme := h.radioUseCase.GetProgramStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(radioStatusResponse{
		IsRunning: isRunning,
		Theme:     theme,
	})
}
