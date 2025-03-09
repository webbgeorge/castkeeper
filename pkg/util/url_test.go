package util_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/util"
)

func TestValidateExtURL(t *testing.T) {
	testCases := map[string]struct {
		url         string
		expectedErr error
	}{
		"valid URL http": {
			url:         "http://www.example.com/feed.xml",
			expectedErr: nil,
		},
		"valid URL https": {
			url:         "https://www.example.com",
			expectedErr: nil,
		},
		"valid URL non-standard port": {
			url:         "http://www.example.com:8080/feed.xml",
			expectedErr: nil,
		},
		"unable to parse": {
			url: ":Not a URL",
			expectedErr: &url.Error{
				Op:  "parse",
				URL: ":Not a URL",
				Err: errors.New("missing protocol scheme"),
			},
		},
		"invalid scheme": {
			url:         "ftp://www.example.com/feed.xml",
			expectedErr: errors.New("invalid scheme in URL"),
		},
		"no explicit scheme": {
			url:         "www.example.com/feed.xml",
			expectedErr: errors.New("invalid scheme in URL"),
		},
		"missing host": {
			url:         "http:///feed.xml",
			expectedErr: errors.New("URL host must not be empty"),
		},
		"localhost standard port": {
			url:         "http://localhost",
			expectedErr: errors.New("URL host must not be localhost"),
		},
		"localhost non-standard port": {
			url:         "https://localhost:8443/feed.xml",
			expectedErr: errors.New("URL host must not be localhost"),
		},
		"ipv6 address": {
			url:         "http://[2001:db8:3333:4444:5555:6666:7777:8888]/feed.xml",
			expectedErr: errors.New("URL host cannot be IP address"),
		},
		"ipv6 address local loopback": {
			url:         "http://[::1]/feed.xml",
			expectedErr: errors.New("URL host cannot be IP address"),
		},
		"ipv4 address standard port": {
			url:         "http://10.0.0.1/feed.xml",
			expectedErr: errors.New("URL host cannot be IP address"),
		},
		"ipv4 address non-standard port": {
			url:         "http://127.0.0.1:8000/feed.xml",
			expectedErr: errors.New("URL host cannot be IP address"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := util.ValidateExtURL(tc.url)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSanitiseGUID(t *testing.T) {
	g1 := util.SanitiseGUID("123-456-789")
	assert.Equal(t, "123-456-789", g1)

	g2 := util.SanitiseGUID("123:abc:&^*")
	assert.Equal(t, "123-abc-", g2)
}
