package sirch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Addr      string
	OpenAIKey string
}
type Option func(*Config) error

func WithAddr(addr string) Option {
	return func(c *Config) error {
		c.Addr = addr
		return nil
	}
}

func WithOpenAIKey(key string) Option {
	return func(c *Config) error {
		c.OpenAIKey = key
		return nil
	}
}

func NewConfig(options ...Option) (*Config, error) {
	conf := &Config{
		Addr: ":8080",
	}
	for _, o := range options {
		if err := o(conf); err != nil {
			return nil, err
		}
	}

	if conf.OpenAIKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	return conf, nil
}

func Start(options ...Option) error {
	conf, err := NewConfig(options...)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	router, err := NewApp(conf)
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
