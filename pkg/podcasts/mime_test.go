package podcasts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/gopodcast"
)

func TestDetectMIMEType(t *testing.T) {
	testCases := map[string]struct {
		enclosure        gopodcast.Enclosure
		expectedErr      bool
		expectedMIMEType string
	}{
		"usesTypeByDefault": {
			enclosure: gopodcast.Enclosure{
				Type: "audio/mpeg",
				URL:  "http://example.com/podcast1.mov",
			},
			expectedErr:      false,
			expectedMIMEType: "audio/mpeg",
		},
		"fallsBackToURLExtWhenTypeNotPresent": {
			enclosure: gopodcast.Enclosure{
				Type: "",
				URL:  "http://example.com/podcast1.mov",
			},
			expectedErr:      false,
			expectedMIMEType: "video/quicktime",
		},
		"fallsBackToURLExtWhenTypeNotValid": {
			enclosure: gopodcast.Enclosure{
				Type: "not/valid",
				URL:  "http://example.com/podcast2.m4a",
			},
			expectedErr:      false,
			expectedMIMEType: "audio/x-m4a",
		},
		"errorWhenNeitherPresent": {
			enclosure: gopodcast.Enclosure{
				Type: "",
				URL:  "",
			},
			expectedErr:      true,
			expectedMIMEType: "",
		},
		"errorWhenNeitherValid": {
			enclosure: gopodcast.Enclosure{
				Type: "text/plain",
				URL:  "http://example.com/podcast1.doc",
			},
			expectedErr:      true,
			expectedMIMEType: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mimeType, err := podcasts.DetectMIMEType(tc.enclosure)
			assert.Equal(t, tc.expectedErr, err != nil)
			assert.Equal(t, tc.expectedMIMEType, mimeType)
		})
	}
}

func TestMIMETypeExtension(t *testing.T) {
	testCases := map[string]struct {
		mimeType          string
		expectedErr       bool
		expectedExtension string
	}{
		"mp3": {
			mimeType:          "audio/mpeg",
			expectedErr:       false,
			expectedExtension: "mp3",
		},
		"m4a": {
			mimeType:          "audio/x-m4a",
			expectedErr:       false,
			expectedExtension: "m4a",
		},
		"mp4": {
			mimeType:          "video/mp4",
			expectedErr:       false,
			expectedExtension: "mp4",
		},
		"mov": {
			mimeType:          "video/quicktime",
			expectedErr:       false,
			expectedExtension: "mov",
		},
		"invalid": {
			mimeType:          "text/plain",
			expectedErr:       true,
			expectedExtension: "",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			extension, err := podcasts.MIMETypeExtension(tc.mimeType)
			assert.Equal(t, tc.expectedErr, err != nil)
			assert.Equal(t, tc.expectedExtension, extension)
		})
	}
}
