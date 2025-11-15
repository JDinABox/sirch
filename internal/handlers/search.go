package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	aiclient "github.com/JDinABox/sirch/internal/aiClient"
	"github.com/JDinABox/sirch/internal/searxng"
	"github.com/JDinABox/sirch/internal/templates"
	"github.com/JDinABox/sirch/internal/templates/search"
	"github.com/a-h/templ"
	"resty.dev/v3"
)

func Search(aiClient *aiclient.Client, searchClient *searxng.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		queryWSpaces := strings.ReplaceAll(query, "+", " ")

		dataChan := make(chan templates.SlotContents)
		var wg sync.WaitGroup

		client := resty.New()
		defer client.Close()

		wg.Go(func() {
			sr, err := searchClient.Search(r.Context(), query)
			if err != nil {
				slog.Error("unable to get searxng response", "ERROR", err)
				dataChan <- templates.SlotContents{
					Name:     "result",
					Contents: search.R("Something went wrong"),
				}
				return
			}

			if len(sr.Results) == 0 {
				dataChan <- templates.SlotContents{
					Name:     "result",
					Contents: search.R("No results"),
				}
				return
			}

			dataChan <- templates.SlotContents{
				Name:     "result",
				Contents: search.Results(sr),
			}
		})
		wg.Go(func() {
			data, err := aiClient.Run(r.Context(), fmt.Sprintf("[%s]", queryWSpaces))
			if err != nil {
				slog.Error("unable to get ai recommendations", "ERROR", err)
				dataChan <- templates.SlotContents{
					Name:     "recommendations",
					Contents: search.R("Something went wrong"),
				}
				return
			}

			dataChan <- templates.SlotContents{
				Name:     "recommendations",
				Contents: search.Recommendations(strings.Split(data, "\n")),
			}
		})

		go func() {
			wg.Wait()
			close(dataChan)
		}()

		c := templates.Layout(search.Head(), search.Body(queryWSpaces, dataChan))

		w.Header().Set("Content-Type", "text/html")
		templ.Handler(c, templ.WithStreaming()).ServeHTTP(w, r)
	}
}
