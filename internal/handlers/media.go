package handlers

import "net/http"

func Media() http.Handler {
	return http.StripPrefix("/assets/", http.FileServer(http.Dir("web/dist")))
}
