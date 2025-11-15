package searxng

import (
	"encoding/json/v2"
	"fmt"
	"strconv"

	"resty.dev/v3"
)

type Client struct {
	url         string
	restyClient *resty.Client
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
		restyClient: resty.New().
			SetHeader("Accept", "application/json, text/html").
			SetHeader("Accept-Language", "*").SetBaseURL(url),
	}
}

func (c *Client) Search(query string, page ...int) (*SearchResponse, error) {
	queryParams := map[string]string{
		"q":      query,
		"format": "json",
	}
	if len(page) != 0 {
		queryParams["pageno"] = strconv.Itoa(page[0])
	}

	res, err := c.restyClient.R().
		SetQueryParams(queryParams).
		Get("/search")
	if err != nil || res.StatusCode() >= 400 {
		return nil, fmt.Errorf("unable to fetch searxng response status: %d error: %w", res.StatusCode(), err)
	}

	rawJSON := res.Bytes()
	var sr *SearchResponse
	if err = json.Unmarshal(rawJSON, &sr); err != nil {
		return nil, fmt.Errorf("unable to unmarshal searxng response error: %w\nraw JSON:\n%s\n", err, rawJSON)
	}
	SortResults(&sr.Results)
	return sr, nil
}
