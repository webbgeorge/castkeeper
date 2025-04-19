package fixures

import (
	"net/http"
	"os"
	"path"

	"github.com/stretchr/testify/assert"
)

var TestItunesHTTPClient *http.Client = &http.Client{
	Transport: &testItunesTransport{},
}

type testItunesTransport struct{}

func (t *testItunesTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	query := r.URL.Query().Get("term")

	switch query {
	case "error":
		return nil, assert.AnError
	case "500":
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	testDataRoot, err := os.OpenRoot(path.Join(fixtureDir(), "testdata", "itunes"))
	if err != nil {
		panic(err)
	}

	f, err := testDataRoot.Open(query + ".json")
	if err != nil {
		panic(err)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}
