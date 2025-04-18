package fixures

import (
	"net/http"
	"os"
	"path"
	"runtime"
)

var TestDataHTTPClient *http.Client = &http.Client{
	Transport: &testDataTransport{},
}

type testDataTransport struct{}

func (t *testDataTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	filePath := r.URL.Host + r.URL.Path

	if filePath[:9] != "testdata/" {
		panic("unexpected testdata file path")
	}

	filePath = path.Join(fixtureDir(), filePath)

	testDataRoot, err := os.OpenRoot(path.Join(fixtureDir(), "testdata"))
	if err != nil {
		panic(err)
	}

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
