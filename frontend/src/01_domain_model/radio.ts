// 【第1層: Domain Model】システムの存在意義である中核データと普遍ルール
// 制約: React Hooks, JSX, DOM API, HTTP通信を絶対に持ち込まず純粋なTypeScriptで表現

export type Speaker = 'zundamon' | 'metan';

export interface Comment {
  userName: string;
  content: string;
}

export interface Dialogue {
  speaker: Speaker;
  text: string;
  audioBase64?: string;
  timestamp?: number;
}

// ドメイン不変条件（ビジネスルール）の検証
export function validateComment(userName: string, content: string): { valid: boolean; error?: string } {
  if (userName.trim().length > 30) {
    return { valid: false, error: 'ラジオネームは30文字以内で入力してください' };
  }
  if (!content.trim()) {
    return { valid: false, error: 'お便りの本文を入力してください' };
  }
  if (content.trim().length > 200) {
    return { valid: false, error: 'お便りは200文字以内で入力してください' };
  }
  return { valid: true };
}
