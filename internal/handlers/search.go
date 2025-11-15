package handlers

import (
	"encoding/json/v2"
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

func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		queryWSpaces := strings.ReplaceAll(query, "+", " ")

		dataChan := make(chan templates.SlotContents)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			client := resty.New()
			defer client.Close()

			res, err := client.R().
				SetQueryParams(map[string]string{
					"q":      query,
					"format": "json",
				}).
				SetHeader("Accept", "application/json, text/html").
				SetHeader("Accept-Language", "*").
				Get("https://searxng.crawford.zone/search")
			if err != nil || res.StatusCode() >= 400 {
				slog.Error("unable to fetch searxng response", "ERROR", fmt.Sprintf("status: %d error: %v", res.StatusCode(), err))
				dataChan <- templates.SlotContents{
					Name:     "result",
					Contents: search.R("Something went wrong"),
				}
				return
			}

			// Debug: Log the raw JSON response
			rawJSON := res.Bytes()

			var sr searxng.SearchResponse
			sr.Results = []searxng.Result{}

			err = json.Unmarshal(rawJSON, &sr)
			if err != nil {
				slog.Error("unable to decode json", "ERROR", fmt.Sprintf("error: %v json: %s", err, rawJSON))
				dataChan <- templates.SlotContents{
					Name:     "result",
					Contents: search.R("Something went wrong"),
				}
			}

			searxng.SortResults(&sr.Results)

			dataChan <- templates.SlotContents{
				Name:     "result",
				Contents: search.Results(sr),
			}
		}()
		go func() {
			defer wg.Done()
			data, err := aiclient.Run(fmt.Sprintf("[%s]", queryWSpaces))
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
		}()

		go func() {
			wg.Wait()
			close(dataChan)
		}()

		c := templates.Layout(search.Head(), search.Body(queryWSpaces, dataChan))

		w.Header().Set("Content-Type", "text/html")
		templ.Handler(c, templ.WithStreaming()).ServeHTTP(w, r)
	}
}
