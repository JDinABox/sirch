package sirch

import (
	"github.com/JDinABox/sirch/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Config struct is assumed to exist and is used for other application configurations.
// ZITADEL parameters are passed as arguments to NewApp.

func NewApp(c *Config) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Get("/", handlers.Home())
	r.Route("/search", func(r chi.Router) {
		r.Get("/", handlers.Search())
	})
	return r, nil
}
