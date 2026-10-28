// 【第7層: Presentation / View】JSXマークアップとUI描画
// 役割: AI感を排した本物のラジオスタジオ風の受動的UI

import React from 'react';
import { Dialogue, Comment } from '../01_domain_model/radio';
import { Mic, Send, Radio, Volume2, Play, RefreshCw, Mail } from 'lucide-react';

interface RadioRoomViewProps {
  isBroadcasting: boolean;
  programTitle: string;
  currentTheme: string;
  isStarting: boolean;
  dialogues: Dialogue[];
  comments: Comment[];
  userName: string;
  content: string;
  isSubmitting: boolean;
  errorMsg: string | null;
  isPlaying: boolean;
  needsAudioPermission: boolean;
  onProgramTitleChange: (val: string) => void;
  onUserNameChange: (val: string) => void;
  onContentChange: (val: string) => void;
  onStartBroadcast: (e: React.FormEvent) => void;
  onSubmitComment: (e: React.FormEvent) => void;
  onEnableAudio: () => void;
}

export const RadioRoomView: React.FC<RadioRoomViewProps> = ({
  isBroadcasting,
  programTitle,
  currentTheme,
  isStarting,
  dialogues,
  comments,
  userName,
  content,
  isSubmitting,
  errorMsg,
  isPlaying,
  needsAudioPermission,
  onProgramTitleChange,
  onUserNameChange,
  onContentChange,
  onStartBroadcast,
  onSubmitComment,
  onEnableAudio,
}) => {
  return (
    <div className="min-h-screen bg-[#1c1917] text-[#e7e5e4] flex flex-col font-sans selection:bg-amber-800 selection:text-white">
      {/* スタジオヘッダー */}
      <header className="border-b border-[#292524] bg-[#141211] sticky top-0 z-20 px-6 py-4 flex items-center justify-between shadow-md">
        <div className="flex items-center gap-3">
          <div className="p-2.5 bg-[#292524] text-amber-500 rounded-lg border border-[#44403c]">
            <Radio className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-bold tracking-tight text-stone-100">
                FM RADIO ROOM
              </h1>
              <span className="text-[11px] font-mono uppercase tracking-wider px-2 py-0.5 rounded bg-stone-800 text-stone-400 border border-stone-700">
                STUDIO 1
              </span>
            </div>
            <p className="text-xs text-stone-400">パーソナリティ: ずんだもん & 四国めたん</p>
          </div>
        </div>

        {/* ON AIR サイン（本物のスタジオランプ風） */}
        <div className="flex items-center gap-3">
          <div
            className={`flex items-center gap-2 px-3.5 py-1.5 rounded border text-xs font-bold tracking-widest uppercase transition-all ${
              isBroadcasting
                ? 'bg-red-950/80 border-red-600 text-red-500 shadow-[0_0_15px_rgba(239,68,68,0.35)]'
                : 'bg-stone-900 border-stone-800 text-stone-600'
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${
                isBroadcasting ? 'bg-red-500 animate-pulse' : 'bg-stone-700'
              }`}
            />
            ON AIR
          </div>

          {isPlaying && (
            <span className="hidden sm:flex items-center gap-1.5 px-3 py-1 bg-amber-950/40 border border-amber-800/40 text-amber-400 rounded text-xs">
              <Volume2 className="w-3.5 h-3.5" /> 再生中
            </span>
          )}
        </div>
      </header>

      {/* 音声自動再生ブロック時の許可バー */}
      {needsAudioPermission && (
        <div className="bg-amber-500 text-stone-950 px-6 py-2.5 flex items-center justify-between text-xs font-bold shadow-md animate-pulse">
          <div className="flex items-center gap-2">
            <Volume2 className="w-4 h-4" />
            <span>ブラウザの自動再生ポリシーにより音声が保留されています。音声を聴くにはボタンを押してください。</span>
          </div>
          <button
            onClick={onEnableAudio}
            className="px-3 py-1 bg-stone-950 text-amber-400 hover:bg-stone-900 rounded font-bold text-xs transition"
          >
            音声を有効にする
          </button>
        </div>
      )}

      {/* メインエリア */}
      <main className="flex-1 max-w-6xl w-full mx-auto p-4 md:p-6 flex flex-col gap-6">
        {/* 番組未開始時のタイトル入力セットアップカード */}
        {!isBroadcasting ? (
          <div className="my-auto max-w-xl w-full mx-auto bg-[#262626] border border-[#404040] rounded-xl p-8 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-3 bg-amber-500/10 text-amber-500 rounded-lg">
                <Mic className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-lg font-bold text-stone-100">生放送のタイトルを決める</h2>
                <p className="text-xs text-stone-400">
                  ずんだもんとめたんが話すテーマを入力して、番組をスタートしてください。
                </p>
              </div>
            </div>

            <form onSubmit={onStartBroadcast} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-stone-300 mb-1.5">
                  番組タイトル・トークテーマ
                </label>
                <input
                  type="text"
                  placeholder="例: 深夜のプログラミング雑談 / 最近買ってよかったもの / 失敗談SP"
                  value={programTitle}
                  onChange={(e) => onProgramTitleChange(e.target.value)}
                  className="w-full bg-[#171717] border border-[#404040] rounded-lg px-4 py-3 text-sm text-stone-100 placeholder-stone-500 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
                  autoFocus
                />
              </div>

              {errorMsg && (
                <div className="text-xs text-rose-400 bg-rose-950/30 p-2.5 rounded border border-rose-900/50">
                  {errorMsg}
                </div>
              )}

              <button
                type="submit"
                disabled={isStarting || !programTitle.trim()}
                className="w-full py-3 bg-amber-600 hover:bg-amber-500 disabled:opacity-50 disabled:cursor-not-allowed text-stone-950 font-bold text-sm rounded-lg transition shadow-md flex items-center justify-center gap-2"
              >
                {isStarting ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    ブース接続中...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 fill-current" />
                    生放送を開始する（ON AIR）
                  </>
                )}
              </button>
            </form>
          </div>
        ) : (
          /* 放送中のスタジオ画面 */
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1">
            {/* 左2列: スタジオトークログ */}
            <div className="lg:col-span-2 flex flex-col bg-[#262626] border border-[#3f3f46]/40 rounded-xl p-5 shadow-lg">
              {/* 番組名サインボード */}
              <div className="border-b border-[#3f3f46]/40 pb-4 mb-4 flex items-center justify-between">
                <div>
                  <div className="text-[11px] font-semibold text-amber-500 uppercase tracking-wider">
                    TODAY'S THEME
                  </div>
                  <h2 className="text-base sm:text-lg font-bold text-stone-100 mt-0.5">
                    『{currentTheme}』
                  </h2>
                </div>

                <button
                  onClick={() => onProgramTitleChange('')}
                  title="別のテーマで再スタート"
                  className="text-xs text-stone-400 hover:text-stone-200 px-3 py-1 rounded bg-[#1f1f1f] border border-stone-700/60 transition flex items-center gap-1.5"
                >
                  <RefreshCw className="w-3.5 h-3.5" /> テーマ変更
                </button>
              </div>

              {/* トークログ（台本・スクリプト風） */}
              <div className="flex-1 overflow-y-auto space-y-4 pr-1 max-h-[560px]">
                {dialogues.length === 0 ? (
                  <div className="h-64 flex flex-col items-center justify-center text-stone-500 text-sm">
                    <Mic className="w-8 h-8 mb-2 opacity-40 animate-pulse text-amber-500" />
                    パーソナリティがオープニングトークを準備しています...
                  </div>
                ) : (
                  dialogues.map((d, index) => {
                    const isZundamon = d.speaker === 'zundamon';
                    return (
                      <div key={index} className="flex items-start gap-3.5">
                        {/* DJ名札 */}
                        <div
                          className={`w-24 shrink-0 pt-1 text-right text-xs font-bold ${
                            isZundamon ? 'text-emerald-400' : 'text-rose-400'
                          }`}
                        >
                          {isZundamon ? 'ずんだもん' : '四国めたん'}
                        </div>

                        {/* セリフカード */}
                        <div
                          className={`flex-1 rounded-lg px-4 py-3 text-sm leading-relaxed border ${
                            isZundamon
                              ? 'bg-[#1e2621] border-emerald-900/40 text-stone-200'
                              : 'bg-[#291f24] border-rose-900/40 text-stone-200'
                          }`}
                        >
                          {d.text}
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            {/* 右1列: お便りポスト & リスナーメッセージ */}
            <div className="flex flex-col gap-6">
              {/* お便り投稿フォーム */}
              <div className="bg-[#262626] border border-[#3f3f46]/40 rounded-xl p-5 shadow-lg">
                <div className="flex items-center gap-2 mb-3">
                  <Mail className="w-4 h-4 text-amber-500" />
                  <h3 className="text-sm font-bold text-stone-200">お便りポスト（メッセージ）</h3>
                </div>

                <form onSubmit={onSubmitComment} className="space-y-3">
                  <div>
                    <input
                      type="text"
                      placeholder="ラジオネーム（省略時は名無し）"
                      value={userName}
                      onChange={(e) => onUserNameChange(e.target.value)}
                      className="w-full bg-[#171717] border border-[#404040] rounded-lg px-3 py-2 text-xs text-stone-200 placeholder-stone-500 focus:outline-none focus:border-amber-500"
                    />
                  </div>
                  <div>
                    <textarea
                      rows={3}
                      placeholder="パーソナリティへの質問・感想・話題の提供（200文字以内）"
                      value={content}
                      onChange={(e) => onContentChange(e.target.value)}
                      className="w-full bg-[#171717] border border-[#404040] rounded-lg px-3 py-2 text-xs text-stone-200 placeholder-stone-500 focus:outline-none focus:border-amber-500 resize-none"
                    />
                  </div>

                  {errorMsg && <div className="text-xs text-rose-400">{errorMsg}</div>}

                  <button
                    type="submit"
                    disabled={isSubmitting || !content.trim()}
                    className="w-full py-2 bg-stone-700 hover:bg-stone-600 disabled:opacity-40 disabled:cursor-not-allowed text-stone-100 text-xs font-semibold rounded-lg transition flex items-center justify-center gap-1.5"
                  >
                    <Send className="w-3.5 h-3.5" />
                    {isSubmitting ? '投函中...' : 'お便りを番組に届ける'}
                  </button>
                </form>
              </div>

              {/* 届いたお便り一覧 */}
              <div className="flex-1 bg-[#262626] border border-[#3f3f46]/40 rounded-xl p-5 shadow-lg flex flex-col">
                <div className="text-xs font-semibold text-stone-400 mb-3 flex items-center justify-between">
                  <span>番組に届いたお便り</span>
                  <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-stone-800 text-stone-400">
                    {comments.length} 通
                  </span>
                </div>

                <div className="flex-1 overflow-y-auto space-y-2.5 max-h-[280px]">
                  {comments.length === 0 ? (
                    <div className="text-center text-xs text-stone-600 py-8">
                      まだお便りは届いていません
                    </div>
                  ) : (
                    comments.map((c, i) => (
                      <div
                        key={i}
                        className="p-3 bg-[#1c1917] rounded-lg border border-stone-800 text-xs"
                      >
                        <div className="font-semibold text-amber-500/90 mb-1">
                          ラジオネーム: {c.userName} さん
                        </div>
                        <div className="text-stone-300 leading-normal">{c.content}</div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};
