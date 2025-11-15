package handlers

import (
	"net/http"

	"github.com/JDinABox/sirch/internal/templates"
	"github.com/JDinABox/sirch/internal/templates/home"
)

func Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.Layout(home.Head(), home.Body()).Render(r.Context(), w)
	}
}
