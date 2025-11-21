package sirch

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"database/sql"

	sqlcdb "github.com/JDinABox/sirch/db"
	aiclient "github.com/JDinABox/sirch/internal/aiClient"
	"github.com/JDinABox/sirch/internal/db"
	"github.com/JDinABox/sirch/internal/searxng"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func Start(options ...Option) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := NewConfig(options...)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	aiClient := aiclient.NewClient("https://openrouter.ai/api/v1", conf.OpenAIKey)

	if err = applyMigrations(conf.DBPath); err != nil {
		return err
	}
	dbConn, err := sql.Open("sqlite3", "file:"+conf.DBPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	queries := db.New(dbConn)

	wg.Go(func() {
		t := time.Tick(5 * time.Minute)
		for {
			select {
			case <-t:
				ctxLimit, cancel := context.WithTimeout(ctx, time.Second*30)
				defer cancel()
				err := queries.DeleteOld(ctxLimit, time.Now())
				if err != nil {
					slog.Error("err running DB Cleanup", "Error", err)
				}
			case <-ctx.Done():
				slog.Info("Closing DB Cleanup")
				return
			}
		}
	})

	searchClient := searxng.NewClient(conf.SearxngHost, queries)
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

	ctxTmt, cancelTmt := context.WithTimeout(ctx, time.Second*15)
	defer cancelTmt()

	slog.Info("shutting server down...")
	if err = server.Shutdown(ctxTmt); err != nil {
		slog.Error("shutting down server", "ERROR", err)
	}

	cancel()

	wg.Wait()

	return nil
}

func applyMigrations(dbConn string) error {
	u, _ := url.Parse("sqlite:" + dbConn)
	db := dbmate.New(u)
	db.FS = sqlcdb.FS
	db.MigrationsDir = []string{"./migrations"}

	migrations, err := db.FindMigrations()
	if err != nil {
		return err
	}
	for _, m := range migrations {
		slog.Info("Migration:", "Version", m.Version, "Path", m.FilePath, "Applied", m.Applied)
	}

	slog.Info("Applying migrations...")
	err = db.CreateAndMigrate()
	if err != nil {
		return err
	}
	return nil
}
