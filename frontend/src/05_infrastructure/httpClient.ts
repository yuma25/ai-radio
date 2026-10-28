// 【第5層: Infrastructure】HTTP通信クライアント基盤

export class HttpClient {
  constructor(private readonly baseURL: string = '') {}

  async get<T>(path: string): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const response = await fetch(url);
    if (!response.ok) {
      const errData = await response.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP error! status: ${response.status}`);
    }
    return response.json();
  }

  async post<T>(path: string, body: unknown): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errData = await response.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  createEventSource(path: string): EventSource {
    const url = `${this.baseURL}${path}`;
    return new EventSource(url);
  }
}
