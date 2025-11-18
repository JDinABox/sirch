package searxng

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/JDinABox/sirch/internal/db"
	"resty.dev/v3"
)

type Client struct {
	url         string
	restyClient *resty.Client
	queries     *db.Queries
}

func NewClient(url string, q *db.Queries) *Client {
	return &Client{
		url: url,
		restyClient: resty.New().
			SetHeader("Accept", "application/json, text/html").
			SetHeader("Accept-Language", "*").SetBaseURL(url),
		queries: q,
	}
}

func (c *Client) Search(ctx context.Context, query string, page ...int) (*SearchResponse, error) {
	pageno := "1"
	if len(page) != 0 {
		pageno = strconv.Itoa(page[0])
	}

	cacheKey := "search-" + query + "-" + pageno
	s, err := c.getFromDB(ctx, cacheKey)
	if err == nil {
		slog.Info("Cache Hit", "Key", cacheKey)
		return s, nil
	}
	if !errors.Is(err, errNotFoundInCache) && !errors.Is(err, errOldCache) {
		return nil, err
	}
	queryParams := map[string]string{
		"q":      query,
		"format": "json",
	}
	if len(page) != 0 {
		queryParams["pageno"] = pageno
	}

	res, err := c.restyClient.R().WithContext(ctx).
		SetQueryParams(queryParams).
		Get("/search")
	if err != nil || res.StatusCode() >= 400 {
		return nil, fmt.Errorf("unable to fetch searxng response status: %d error: %w", res.StatusCode(), err)
	}
	defer res.Body.Close()

	rawJSON := res.Bytes()
	var sr SearchResponse
	if err = json.Unmarshal(rawJSON, &sr); err != nil {
		return nil, fmt.Errorf("unable to unmarshal searxng response error: %w\nraw JSON:\n%s\n", err, rawJSON)
	}
	OrderResults(&sr.Results)

	return &sr, c.addToDB(ctx, cacheKey, &sr)
}

func (c *Client) Close() {
	c.restyClient.Close()
}

var errNotFoundInCache = errors.New("cache result not found")
var errOldCache = errors.New("cache result old")

func (c *Client) getFromDB(ctx context.Context, key string) (*SearchResponse, error) {
	r, err := c.queries.GetCache(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("unable to find: %s, err: %w", key, errNotFoundInCache)
		}
		return nil, err
	}
	if time.Since(r.CreatedAt) > time.Minute*5 {
		return nil, fmt.Errorf("key: %s, err: %w", key, errOldCache)
	}

	sr := &SearchResponse{}
	_, err = sr.UnmarshalMsg(r.Data)
	if err != nil {
		return nil, err
	}

	return sr, nil
}

var errNilSearchResponse = errors.New("nil SearchResponse pointer")

func (c *Client) addToDB(ctx context.Context, key string, sr *SearchResponse) error {
	if sr == nil {
		return fmt.Errorf("key: %s err: %w", errNilSearchResponse)
	}
	o, err := sr.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return c.queries.InsertCache(ctx, db.InsertCacheParams{Key: key, Data: o})
}
