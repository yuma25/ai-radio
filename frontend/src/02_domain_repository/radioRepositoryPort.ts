// 【第2層: Domain Repository Port】外部通信・データ取得の契約書

import { Comment, Dialogue } from '../01_domain_model/radio';

export interface RadioRepositoryPort {
  startRadio(title: string): Promise<void>;
  getStatus(): Promise<{ isRunning: boolean; theme: string }>;
  postComment(comment: Comment): Promise<void>;
  subscribeStream(onDialogue: (dialogue: Dialogue) => void, onComment: (comment: Comment) => void): () => void;
}
