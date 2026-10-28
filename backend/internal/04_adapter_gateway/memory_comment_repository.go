package gateway

import (
	"context"
	"sync"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

type memoryCommentRepository struct {
	mu           sync.RWMutex
	allComments  []*model.Comment
	unprocessed  []*model.Comment
}

// NewMemoryCommentRepository はCommentRepositoryのインメモリ実装を生成する
func NewMemoryCommentRepository() repository.CommentRepository {
	return &memoryCommentRepository{
		allComments: make([]*model.Comment, 0),
		unprocessed: make([]*model.Comment, 0),
	}
}

func (r *memoryCommentRepository) Save(ctx context.Context, comment *model.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.allComments = append(r.allComments, comment)
	r.unprocessed = append(r.unprocessed, comment)
	return nil
}

func (r *memoryCommentRepository) FindAll(ctx context.Context) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copied := make([]*model.Comment, len(r.allComments))
	copy(copied, r.allComments)
	return copied, nil
}

func (r *memoryCommentRepository) PopNextUnprocessed(ctx context.Context) (*model.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.unprocessed) == 0 {
		return nil, nil
	}

	next := r.unprocessed[0]
	r.unprocessed = r.unprocessed[1:]
	return next, nil
}
