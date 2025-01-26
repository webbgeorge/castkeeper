package itunes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ItunesAPI struct {
	HTTPClient *http.Client
}

type SearchResult struct {
	CollectionName string `json:"collectionName"`
	FeedURL        string `json:"feedUrl"`
	ArtistName     string `json:"artistName"`
	TrackCount     int    `json:"trackCount"`
	ArtworkURL60   string `json:"artworkUrl60"`
	ArtworkURL100  string `json:"artworkUrl100"`
	ArtworkURL600  string `json:"artworkUrl600"`
}

func (r SearchResult) ArtworkURL() string {
	if r.ArtworkURL600 != "" {
		return r.ArtworkURL600
	}
	if r.ArtworkURL100 != "" {
		return r.ArtworkURL100
	}
	if r.ArtworkURL60 != "" {
		return r.ArtworkURL60
	}
	return ""
}

type searchResponse struct {
	Results []SearchResult `json:"results"`
}

func (i *ItunesAPI) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" || len(query) > 250 {
		return nil, fmt.Errorf("expected query to be between 1 and 250 chars, got %d", len(query))
	}

	q := url.Values{}
	q.Add("term", query)
	q.Add("media", "podcast")
	q.Add("entity", "podcast")
	u := fmt.Sprintf("https://itunes.apple.com/search?%s", q.Encode())

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	res, err := i.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("itunes returned status '%d'", res.Status)
	}

	var bodyData searchResponse
	err = json.NewDecoder(res.Body).Decode(&bodyData)
	if err != nil {
		return nil, err
	}

	return bodyData.Results, nil
}
