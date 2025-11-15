package handlers

import (
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	aiclient "github.com/JDinABox/sirch/internal/aiClient"
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
			slog.Info("Status Code", "INFO", strconv.Itoa(res.StatusCode()))
			slog.Info("err", "INFO", err)
			if err != nil || res.StatusCode() >= 400 {
				dataChan <- templates.SlotContents{
					Name:     "result",
					Contents: search.R("Something went wrong"),
				}
				return
			}
			var j map[string]string
			json.Unmarshal(res.Bytes(), &j)
			/*var jr []search.SearchResult
			if _, ok := j[""] {

			}*/
			dataChan <- templates.SlotContents{
				Name:     "result",
				Contents: search.R(fmt.Sprint(j)),
			}
		}()
		go func() {
			defer wg.Done()
			data, err := aiclient.Run(fmt.Sprintf("[%s]", queryWSpaces))
			if err != nil {
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
