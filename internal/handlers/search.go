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

type sd struct {
	Title string
	URL   string
	MD    string
}

func Search(aiClient *aiclient.Client, searchClient *searxng.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		queryWSpaces := strings.ReplaceAll(query, "+", " ")

		dataChan := make(chan templ.Component)
		var wg sync.WaitGroup

		client := resty.New()
		defer client.Close()

		wg.Go(func() {
			sr, err := searchClient.Search(r.Context(), query)
			if err != nil {
				slog.Error("unable to get searxng response", "ERROR", err)
				dataChan <- search.R("result", "Something went wrong")
				return
			}

			if len(sr.Results) == 0 {
				dataChan <- search.R("result", "No results")
				return
			}

			dataChan <- search.Results(sr)

			// wg.Go(func() {
			// 	l := len(sr.Results)
			// 	// Limit to 2 results
			// 	if l > 2 {
			// 		l = 2
			// 	}
			// 	searxng.OrderForContext(&sr.Results)
			// 	var wgIn sync.WaitGroup
			// 	siteData := make(chan message.AnswerSummaryData, l)
			// 	for i := range l {
			// 		wgIn.Go(func() {
			// 			md, err := webclient.Get(sr.Results[i].URL)
			// 			if err != nil {
			// 				slog.Warn("unable to get page", "WARN", err)
			// 			}

			// 			siteData <- message.AnswerSummaryData{
			// 				Title: sr.Results[i].Title,
			// 				URL:   sr.Results[i].URL,
			// 				MD:    md,
			// 			}
			// 		})
			// 	}
			// 	go func() {
			// 		wgIn.Wait()
			// 		close(siteData)
			// 	}()
			// 	dc := make([]string, 6)
			// 	var data []message.AnswerSummaryData
			// 	for d := range siteData {
			// 		data = append(data, d)
			// 	}

			// 	q := message.AnswerSummaryPrompt(&data, queryWSpaces)
			// 	tokens := len(q) / 4
			// 	dc = append(dc, "Tokens: "+strconv.Itoa(tokens)+"\n\n")
			// 	/*var wg2 sync.WaitGroup
			// 	mo := map[string]string{
			// 		"google/gemini-2.5-flash-lite-preview-09-2025": "Gemini 2.5 Flash Lite Preview 09/25",
			// 		"qwen/qwen3-30b-a3b-instruct-2507":             "Qwen3 30b a3b instruct 2507",
			// 		"openai/gpt-5-nano":                            "GPT 5 Nano",
			// 	}
			// 	for k, v := range mo {
			// 		wg2.Go(func() {
			// 			s := time.Now()
			// 			o, _ := aiClient.Run(r.Context(), k, message.SystemAnswerSummary, q)
			// 			e := time.Since(s).Seconds()
			// 			dc = append(dc, fmt.Sprintf("%s: Time %f Cost: %15.14f\n%s", v, e, o.Cost, o.Content))
			// 		})
			// 	}
			// 	wg2.Wait()*/
			// 	dc = append(dc, q)
			// 	dataChan <- templates.SlotContents{
			// 		Name:     "top2",
			// 		Contents: search.Top2(dc),
			// 	}
			// })
		})
		wg.Go(func() {
			data, err := aiClient.RunQueryExpand(r.Context(), fmt.Sprintf("[%s]", queryWSpaces))
			if err != nil {
				slog.Error("unable to get ai recommendations", "ERROR", err)
				dataChan <- search.R("recommendations", "Something went wrong")
				return
			}

			dataChan <- search.Recommendations(strings.Split(data.Content, "\n"))
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
