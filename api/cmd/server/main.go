package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csv-import-kit/api/internal/handlers"
	"csv-import-kit/api/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	// ローカルだけ .env を読む（存在しなくてもOK）
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// --- DB接続（pgxpool） ---
	st, err := store.New()
	if err != nil {
		logger.Error("failed to init DB", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	// --- Handlers / Router ---
	r := chi.NewRouter()
	r.Use(handlers.CORSMiddleware())

	// Health endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	// アップロード→プレビュー
	r.Post("/api/imports", handlers.HandleUploadPreview())

	// マッピング適用（サーバ側）
	r.Post("/api/mappings/apply", handlers.ApplyMapping)

	// テンプレート保存/一覧
	tpl := handlers.NewTemplateHandler(st)
	r.Post("/api/templates", tpl.CreateTemplate)
	r.Get("/api/templates", tpl.ListTemplates)
	r.Get("/api/templates/{id}", tpl.GetTemplateByID)
	r.Delete("/api/templates/{id}", tpl.DeleteTemplate)

	// --- HTTP Server（タイムアウト強化 & Graceful Shutdown） ---
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 起動
	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting api server", "port", port)
		errCh <- srv.ListenAndServe()
	}()

	// シグナル待ち（Ctrl+C / SIGTERM）
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("server stopped gracefully")
}
