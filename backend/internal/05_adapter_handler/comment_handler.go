package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yuma25/AI-Radio-Room/backend/internal/03_usecase"
)

// CommentHandler はリスナーお便りに関するHTTPハンドラ
type CommentHandler struct {
	commentUseCase usecase.CommentUseCase
}

func NewCommentHandler(commentUseCase usecase.CommentUseCase) *CommentHandler {
	return &CommentHandler{
		commentUseCase: commentUseCase,
	}
}

// リクエスト受付用DTO
type postCommentRequest struct {
	UserName string `json:"user_name"`
	Content  string `json:"content"`
}

// レスポンス用DTO
type commentResponse struct {
	UserName string `json:"user_name"`
	Content  string `json:"content"`
}

// HandlePostComment は POST /api/comments を処理する
func (h *CommentHandler) HandlePostComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req postCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input := usecase.PostCommentInput{
		UserName: req.UserName,
		Content:  req.Content,
	}

	output, err := h.commentUseCase.PostComment(r.Context(), input)
	if err != nil {
		// ビジネスルール違反・ドメインエラー
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	res := commentResponse{
		UserName: output.UserName,
		Content:  output.Content,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

// HandleGetComments は GET /api/comments を処理する
func (h *CommentHandler) HandleGetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	outputs, err := h.commentUseCase.GetComments(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resList := make([]commentResponse, len(outputs))
	for i, o := range outputs {
		resList[i] = commentResponse{
			UserName: o.UserName,
			Content:  o.Content,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resList)
}
