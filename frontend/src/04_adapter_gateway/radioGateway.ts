// 【第4層: Adapter Output / Gateway】外部API通信とデータマッピングの専門職人

import { Comment, Dialogue } from '../01_domain_model/radio';
import { RadioRepositoryPort } from '../02_domain_repository/radioRepositoryPort';
import { HttpClient } from '../05_infrastructure/httpClient';

export class RadioGateway implements RadioRepositoryPort {
  constructor(private readonly client: HttpClient) {}

  async startRadio(title: string): Promise<void> {
    await this.client.post('/api/radio/start', { title });
  }

  async getStatus(): Promise<{ isRunning: boolean; theme: string }> {
    const res = await this.client.get<{ is_running: boolean; theme: string }>('/api/radio/status');
    return {
      isRunning: res.is_running,
      theme: res.theme,
    };
  }

  async postComment(comment: Comment): Promise<void> {
    await this.client.post('/api/comments', {
      user_name: comment.userName,
      content: comment.content,
    });
  }

  subscribeStream(onDialogue: (d: Dialogue) => void, onComment: (c: Comment) => void): () => void {
    const es = this.client.createEventSource('/api/radio/stream');

    es.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === 'dialogue' && payload.data) {
          const d: Dialogue = {
            speaker: payload.data.speaker,
            text: payload.data.text,
            audioBase64: payload.data.audioBase64,
            timestamp: payload.timestamp,
          };
          onDialogue(d);
        } else if (payload.type === 'comment' && payload.data) {
          const c: Comment = {
            userName: payload.data.userName,
            content: payload.data.content,
          };
          onComment(c);
        }
      } catch (err) {
        console.error('SSE JSON parse error:', err);
      }
    };

    return () => {
      es.close();
    };
  }
}
