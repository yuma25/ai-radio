package model

import (
	"errors"
	"strings"
)

var (
	ErrEmptyCommentContent = errors.New("comment content cannot be empty")
	ErrCommentTooLong      = errors.New("comment content exceeds maximum length of 200 characters")
	ErrInvalidSpeaker      = errors.New("invalid speaker")
	ErrEmptyDialogueText   = errors.New("dialogue text cannot be empty")
)

// Speaker はラジオ内の発言者（話者）を表す値オブジェクト
type Speaker string

const (
	SpeakerZundamon Speaker = "zundamon"
	SpeakerMetan    Speaker = "metan"
)

func (s Speaker) IsValid() bool {
	return s == SpeakerZundamon || s == SpeakerMetan
}

// Comment はリスナーから番組へ届いたお便り（純粋な業務データ）
type Comment struct {
	UserName string
	Content  string
}

// NewComment はお便りの不変条件（ルール）を検証して生成する
func NewComment(userName, content string) (*Comment, error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, ErrEmptyCommentContent
	}
	if len([]rune(trimmedContent)) > 200 {
		return nil, ErrCommentTooLong
	}

	trimmedUser := strings.TrimSpace(userName)
	if trimmedUser == "" {
		trimmedUser = "名無しリスナー"
	}

	return &Comment{
		UserName: trimmedUser,
		Content:  trimmedContent,
	}, nil
}

// Dialogue はパーソナリティの発話1コマ（純粋な業務データ）
type Dialogue struct {
	Speaker Speaker
	Text    string
}

// NewDialogue は発話の不変条件（ルール）を検証して生成する
func NewDialogue(speaker Speaker, text string) (*Dialogue, error) {
	if !speaker.IsValid() {
		return nil, ErrInvalidSpeaker
	}
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil, ErrEmptyDialogueText
	}

	return &Dialogue{
		Speaker: speaker,
		Text:    trimmedText,
	}, nil
}
