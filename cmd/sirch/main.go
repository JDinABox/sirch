package main

import (
	_ "embed"
	"log/slog"
	"os"
	"strings"

	"github.com/JDinABox/sirch"
)

func main() {
	confOptions := []sirch.Option{
		sirch.WithOpenAIKey(trimGetEnv("OPENAI_API_KEY")),
	}
	if addr := trimGetEnv("ADDRESS"); addr != "" {
		confOptions = append(confOptions, sirch.WithAddr(addr))
	}
	if err := sirch.Start(confOptions...); err != nil {
		slog.Error("application exited with an error", "error", err)
		os.Exit(1)
	}
}

func trimGetEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
