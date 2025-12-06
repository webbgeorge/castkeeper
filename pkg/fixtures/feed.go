package fixtures

import (
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

var TestDataHTTPClient *http.Client = &http.Client{
	Transport: &testDataTransport{},
}

var AuthenticatedFeedCreds = podcasts.PodcastCredentials{
	Username: "fixtureUser",
	Password: "fixturePass",
}

type testDataTransport struct{}

func (t *testDataTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Host != "testdata" {
		panic("unexpected testdata file path")
	}

	// path to force and error response
	if r.URL.Path == "/error" {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	if strings.HasPrefix(r.URL.Path, "/authenticated") {
		u, p, _ := r.BasicAuth()
		if u != AuthenticatedFeedCreds.Username ||
			p != AuthenticatedFeedCreds.Password {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
	}

	testDataRoot, err := os.OpenRoot(path.Join(fixtureDir(), "testdata"))
	if err != nil {
		panic(err)
	}

	filePath := strings.TrimLeft(r.URL.Path, "/")
	f, err := testDataRoot.Open(filePath)
	if err != nil {
		panic(err)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}

func fixtureDir() string {
	_, thisFilePath, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(thisFilePath))
}
