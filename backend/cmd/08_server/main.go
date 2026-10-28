package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yuma25/AI-Radio-Room/backend/internal/07_di"
)

func main() {
	// 1. 環境変数の読み込み（ルートまたは各階層の .env を自動探索）
	loadEnvFile(".env")
	loadEnvFile("../.env")
	loadEnvFile("../../.env")

	port := getEnv("PORT", "8080")
	voicevoxURL := getEnv("VOICEVOX_URL", "http://localhost:50021")
	geminiAPIKey := getEnv("GEMINI_API_KEY", "")
	geminiModel := getEnv("GEMINI_MODEL", "gemini-2.0-flash")

	log.Printf("[Main] Starting Radio Room Backend...")
	log.Printf("[Main] Config: PORT=%s, VOICEVOX_URL=%s, MODEL=%s", port, voicevoxURL, geminiModel)
	if geminiAPIKey == "" {
		log.Printf("[Main] ERROR: GEMINI_API_KEY is empty! Please check your .env file.")
	} else {
		log.Printf("[Main] SUCCESS: GEMINI_API_KEY loaded successfully.")
	}

	// 2. 【第7層: DI Container】配線工場で依存関係を組み立て
	cfg := di.Config{
		Port:         port,
		VoicevoxURL:  voicevoxURL,
		GeminiAPIKey: geminiAPIKey,
		GeminiModel:  geminiModel,
	}
	container := di.NewContainer(cfg)

	// 3. アプリケーション全体のシグナル監視
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 4. 【第6層: Infrastructure】HTTPサーバーの起動
	go func() {
		log.Printf("[Main] HTTP Server is listening at %s", container.Server.Addr())
		if err := container.Server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Main] Server start error: %v", err)
		}
	}()

	// 5. グレースフルシャットダウンの待機
	sig := <-sigChan
	log.Printf("[Main] Received signal %v. Initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := container.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Main] Server shutdown error: %v", err)
	}

	log.Printf("[Main] Server stopped gracefully.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	log.Printf("[Main] Loaded environment file from: %s", path)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}
