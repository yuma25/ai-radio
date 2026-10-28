// 【第8層: Application Setup / DI Container】
// 役割: 各レイヤのインスタンスを生成・結合する配線工場

import { HttpClient } from '../05_infrastructure/httpClient';
import { RadioGateway } from '../04_adapter_gateway/radioGateway';
import { RadioUseCase, RadioUseCaseImpl } from '../03_usecase/radioUseCase';

export interface FrontendContainer {
  radioUseCase: RadioUseCase;
}

export function createContainer(): FrontendContainer {
  // 第5層: Infrastructure
  const httpClient = new HttpClient();

  // 第4層: Adapter Output / Gateway
  const radioGateway = new RadioGateway(httpClient);

  // 第3層: UseCase
  const radioUseCase = new RadioUseCaseImpl(radioGateway);

  return {
    radioUseCase,
  };
}

export const container = createContainer();
