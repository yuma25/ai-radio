// 【第3層: UseCase】画面の文脈にとらわれない業務シナリオ

import { Comment, Dialogue, validateComment } from '../01_domain_model/radio';
import { RadioRepositoryPort } from '../02_domain_repository/radioRepositoryPort';

export interface RadioUseCase {
  startBroadcast(title: string): Promise<void>;
  checkStatus(): Promise<{ isRunning: boolean; theme: string }>;
  sendComment(userName: string, content: string): Promise<void>;
  startListening(onDialogue: (d: Dialogue) => void, onComment: (c: Comment) => void): () => void;
}

export class RadioUseCaseImpl implements RadioUseCase {
  constructor(private readonly repository: RadioRepositoryPort) {}

  async startBroadcast(title: string): Promise<void> {
    const trimmed = title.trim();
    if (!trimmed) {
      throw new Error('番組タイトルを入力してください');
    }
    await this.repository.startRadio(trimmed);
  }

  async checkStatus(): Promise<{ isRunning: boolean; theme: string }> {
    return this.repository.getStatus();
  }

  async sendComment(userName: string, content: string): Promise<void> {
    const validation = validateComment(userName, content);
    if (!validation.valid) {
      throw new Error(validation.error || '入力内容をご確認ください');
    }

    const comment: Comment = {
      userName: userName.trim() || '名無しリスナー',
      content: content.trim(),
    };

    await this.repository.postComment(comment);
  }

  startListening(onDialogue: (d: Dialogue) => void, onComment: (c: Comment) => void): () => void {
    return this.repository.subscribeStream(onDialogue, onComment);
  }
}
