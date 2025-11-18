package sirch

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	aiclient "github.com/JDinABox/sirch/internal/aiClient"
	"github.com/JDinABox/sirch/internal/searxng"
)

func Start(options ...Option) error {
	conf, err := NewConfig(options...)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	aiClient := aiclient.NewClient("https://openrouter.ai/api/v1", conf.OpenAIKey)
	searchClient := searxng.NewClient(conf.SearxngHost)
	defer searchClient.Close()

	router, err := NewApp(conf, aiClient, searchClient)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	serverErrChan := make(chan error, 1)

	server := &http.Server{
		Addr:    conf.Addr,
		Handler: router,
	}

	go func() {
		slog.Info("server listening", "addr", conf.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		slog.Error("unexpected server error", "error", err)
	case sig := <-quitChan:
		slog.Info("received signal", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	slog.Info("shutting server down...")

	return server.Shutdown(ctx)
}
