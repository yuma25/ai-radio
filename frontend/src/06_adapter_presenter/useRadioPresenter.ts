// 【第6層: Adapter Input / Presenter】画面の状態管理とUI通訳（カスタムフック）

import { useState, useEffect, useRef, useCallback } from 'react';
import { Dialogue, Comment } from '../01_domain_model/radio';
import { RadioUseCase } from '../03_usecase/radioUseCase';

interface PlaybackQueueItem {
  dialogue: Dialogue;
  audio?: HTMLAudioElement;
  hasDisplayed?: boolean;
}

export function useRadioPresenter(useCase: RadioUseCase) {
  const [isBroadcasting, setIsBroadcasting] = useState(false);
  const [programTitle, setProgramTitle] = useState('');
  const [currentTheme, setCurrentTheme] = useState('');
  const [isStarting, setIsStarting] = useState(false);

  const [dialogues, setDialogues] = useState<Dialogue[]>([]);
  const [comments, setComments] = useState<Comment[]>([]);
  const [userName, setUserName] = useState('');
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [needsAudioPermission, setNeedsAudioPermission] = useState(false);

  // 音声・テキスト同期再生キュー管理
  const playbackQueueRef = useRef<PlaybackQueueItem[]>([]);
  const isAudioPlayingRef = useRef<boolean>(false);
  const currentAudioRef = useRef<HTMLAudioElement | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);

  // ブラウザの自動再生ポリシーをアンロックする関数
  const unlockAudioContext = useCallback(() => {
    try {
      if (!audioContextRef.current) {
        const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
        if (AudioCtx) {
          audioContextRef.current = new AudioCtx();
        }
      }
      if (audioContextRef.current && audioContextRef.current.state === 'suspended') {
        audioContextRef.current.resume();
      }
      setNeedsAudioPermission(false);
    } catch (e) {
      console.warn('AudioContext unlock failed:', e);
    }
  }, []);

  const playNextAudio = useCallback(() => {
    if (playbackQueueRef.current.length === 0) {
      isAudioPlayingRef.current = false;
      setIsPlaying(false);
      currentAudioRef.current = null;
      return;
    }

    isAudioPlayingRef.current = true;
    setIsPlaying(true);

    const currentItem = playbackQueueRef.current[0];

    // 音声データがない場合のフォールバック（文字のみ表示し、一定時間待って次へ）
    if (!currentItem.audio) {
      playbackQueueRef.current.shift();
      setDialogues((prev) => [...prev, currentItem.dialogue]);
      setTimeout(() => {
        playNextAudio();
      }, 2000);
      return;
    }

    const audio = currentItem.audio;
    currentAudioRef.current = audio;

    audio.onended = () => {
      playbackQueueRef.current.shift(); // 正常終了時にキューから削除
      currentAudioRef.current = null;
      playNextAudio();
    };

    audio.onerror = (e) => {
      console.error('Audio playback error:', e);
      playbackQueueRef.current.shift(); // エラー時はスキップして次へ
      currentAudioRef.current = null;
      playNextAudio();
    };

    audio
      .play()
      .then(() => {
        setNeedsAudioPermission(false);
        // ★ 音声再生が開始された瞬間に画面へテキストを表示（完全同期）
        if (!currentItem.hasDisplayed) {
          currentItem.hasDisplayed = true;
          setDialogues((prev) => [...prev, currentItem.dialogue]);
        }
      })
      .catch((err) => {
        console.warn('Autoplay blocked by browser policy:', err);
        // 自動再生が拒否された場合はキューを維持し、ユーザーに許可ボタンを表示
        isAudioPlayingRef.current = false;
        setIsPlaying(false);
        setNeedsAudioPermission(true);
      });
  }, []);

  // ユーザーが明示的に音声を有効化した時のハンドラ
  const enableAudio = useCallback(() => {
    unlockAudioContext();
    setNeedsAudioPermission(false);
    if (!isAudioPlayingRef.current && playbackQueueRef.current.length > 0) {
      playNextAudio();
    }
  }, [unlockAudioContext, playNextAudio]);

  // 起動時のステータス確認
  useEffect(() => {
    useCase
      .checkStatus()
      .then((status) => {
        if (status.isRunning) {
          setIsBroadcasting(true);
          setCurrentTheme(status.theme);
        }
      })
      .catch(() => {});
  }, [useCase]);

  // リアルタイムSSEストリーム接続
  useEffect(() => {
    const unsubscribe = useCase.startListening(
      (newDialogue) => {
        // ★ 画面への即時追加は行わず、音声とセットで再生キューへ追加
        let audio: HTMLAudioElement | undefined;
        if (newDialogue.audioBase64) {
          audio = new Audio(`data:audio/wav;base64,${newDialogue.audioBase64}`);
        }

        playbackQueueRef.current.push({
          dialogue: newDialogue,
          audio,
        });

        if (!isAudioPlayingRef.current && !needsAudioPermission) {
          playNextAudio();
        }
      },
      (newComment) => {
        setComments((prev) => [newComment, ...prev]);
      }
    );

    return () => {
      unsubscribe();
      // アンマウント時のオーディオ停止
      if (currentAudioRef.current) {
        currentAudioRef.current.pause();
        currentAudioRef.current = null;
      }
    };
  }, [useCase, playNextAudio, needsAudioPermission]);

  // 番組開始ハンドラ
  const handleStartBroadcast = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!programTitle.trim()) return;

    unlockAudioContext(); // クリック時にオーディオをアンロック
    setIsStarting(true);
    setErrorMsg(null);

    // 既存の再生キュー・音声を停止リセット
    if (currentAudioRef.current) {
      currentAudioRef.current.pause();
      currentAudioRef.current = null;
    }
    playbackQueueRef.current = [];
    isAudioPlayingRef.current = false;
    setIsPlaying(false);

    try {
      await useCase.startBroadcast(programTitle);
      setCurrentTheme(programTitle);
      setIsBroadcasting(true);
      setDialogues([]);
    } catch (err: any) {
      setErrorMsg(err.message || '番組の開始に失敗しました');
    } finally {
      setIsStarting(false);
    }
  };

  // お便り送信ハンドラ
  const handlePostComment = async (e: React.FormEvent) => {
    e.preventDefault();
    unlockAudioContext();
    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      await useCase.sendComment(userName, content);
      setContent('');
    } catch (err: any) {
      setErrorMsg(err.message || 'お便りの送信に失敗しました');
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
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
    setProgramTitle,
    setUserName,
    setContent,
    handleStartBroadcast,
    handlePostComment,
    enableAudio,
  };
}
