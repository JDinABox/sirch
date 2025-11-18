package sirch

import (
	"time"

	aiclient "github.com/JDinABox/sirch/internal/aiClient"
	"github.com/JDinABox/sirch/internal/handlers"
	"github.com/JDinABox/sirch/internal/searxng"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// Config struct is assumed to exist and is used for other application configurations.
// ZITADEL parameters are passed as arguments to NewApp.

func NewApp(conf *Config, aiClient *aiclient.Client, searchClient *searxng.Client) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Get("/", handlers.Home())
	r.Route("/search", func(r chi.Router) {
		if conf.Public {
			r.Use(httprate.LimitByRealIP(10, time.Minute))
		}
		r.Get("/", handlers.Search(aiClient, searchClient))
	})
	r.Handle("/assets/*", handlers.Media())
	return r, nil
}
