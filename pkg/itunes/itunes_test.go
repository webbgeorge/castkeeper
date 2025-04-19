package itunes_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
)

func TestItunesAPISearch_SuccessWithResults(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "testPods")
	assert.Nil(t, err)
	assert.Len(t, results, 2)
}

func TestItunesAPISearch_SuccessNoResults(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "noTestPods")
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestItunesAPISearch_NoQuery(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "")
	assert.Equal(t, "expected query to be between 1 and 250 chars, got 0", err.Error())
	assert.Nil(t, results)
}

func TestItunesAPISearch_QueryTooLong(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), fixtures.StrOfLen(251))
	assert.Equal(t, "expected query to be between 1 and 250 chars, got 251", err.Error())
	assert.Nil(t, results)
}

func TestItunesAPISearch_ReqError(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "error")
	assert.Equal(t, "Get \"https://itunes.apple.com/search?entity=podcast&media=podcast&term=error\": assert.AnError general error for testing", err.Error())
	assert.Nil(t, results)
}

func TestItunesAPISearch_Non200(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "500")
	assert.Equal(t, "itunes returned status '500'", err.Error())
	assert.Nil(t, results)
}

func TestItunesAPISearch_InvalidResBody(t *testing.T) {
	it := itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	results, err := it.Search(context.Background(), "invalidBody")
	assert.Equal(t, "invalid character 'N' looking for beginning of value", err.Error())
	assert.Nil(t, results)
}

func TestSearchResultArtworkURL(t *testing.T) {
	searchResult := itunes.SearchResult{
		ArtworkURL600: "img-600.png",
		ArtworkURL100: "img-100.png",
		ArtworkURL60:  "img-60.png",
	}

	assert.Equal(t, "img-600.png", searchResult.ArtworkURL())

	searchResult.ArtworkURL600 = ""
	assert.Equal(t, "img-100.png", searchResult.ArtworkURL())

	searchResult.ArtworkURL100 = ""
	assert.Equal(t, "img-60.png", searchResult.ArtworkURL())

	searchResult.ArtworkURL60 = ""
	assert.Equal(t, "", searchResult.ArtworkURL())
}
