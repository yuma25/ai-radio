package usecase

import (
	"context"

	"github.com/yuma25/AI-Radio-Room/backend/internal/01_domain_model"
	"github.com/yuma25/AI-Radio-Room/backend/internal/02_domain_repository"
)

// PostCommentInput はお便り投稿シナリオの入力DTO
type PostCommentInput struct {
	UserName string
	Content  string
}

// CommentDTO はお便りシナリオの出力DTO
type CommentDTO struct {
	UserName string
	Content  string
}

// CommentUseCase はお便りに関する業務シナリオ契約
type CommentUseCase interface {
	PostComment(ctx context.Context, input PostCommentInput) (*CommentDTO, error)
	GetComments(ctx context.Context) ([]*CommentDTO, error)
}

type commentUseCase struct {
	commentRepo repository.CommentRepository
	broadcaster repository.StreamBroadcaster
}

// NewCommentUseCase はCommentUseCaseを生成する
func NewCommentUseCase(
	commentRepo repository.CommentRepository,
	broadcaster repository.StreamBroadcaster,
) CommentUseCase {
	return &commentUseCase{
		commentRepo: commentRepo,
		broadcaster: broadcaster,
	}
}

// PostComment はお便り投稿シナリオを実行する
func (uc *commentUseCase) PostComment(ctx context.Context, input PostCommentInput) (*CommentDTO, error) {
	// 1. 第1層ドメインモデルを生成し、不変条件を検証
	comment, err := model.NewComment(input.UserName, input.Content)
	if err != nil {
		return nil, err
	}

	// 2. 第2層リポジトリ経由で保存
	if err := uc.commentRepo.Save(ctx, comment); err != nil {
		return nil, err
	}

	// 3. リアルタイム配信基盤へお便り受信を通知
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastComment(comment)
	}

	// 4. 出力DTOへ詰め替えて返却
	return &CommentDTO{
		UserName: comment.UserName,
		Content:  comment.Content,
	}, nil
}

// GetComments はお便り一覧取得シナリオを実行する
func (uc *commentUseCase) GetComments(ctx context.Context) ([]*CommentDTO, error) {
	comments, err := uc.commentRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]*CommentDTO, len(comments))
	for i, c := range comments {
		outputs[i] = &CommentDTO{
			UserName: c.UserName,
			Content:  c.Content,
		}
	}
	return outputs, nil
}
