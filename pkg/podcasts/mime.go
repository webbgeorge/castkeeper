package podcasts

import (
	"fmt"
	"strings"

	"github.com/webbgeorge/gopodcast"
)

var mimeToExt = map[string]string{
	"audio/mpeg":      "mp3",
	"audio/x-m4a":     "m4a",
	"video/mp4":       "mp4",
	"video/quicktime": "mov",
}

var extToMIME = map[string]string{
	"mp3": "audio/mpeg",
	"m4a": "audio/x-m4a",
	"mp4": "video/mp4",
	"mov": "video/quicktime",
}

func DetectMIMEType(enclosure gopodcast.Enclosure) (string, error) {
	// use enclosure type by default
	if _, ok := mimeToExt[enclosure.Type]; ok {
		return enclosure.Type, nil
	}

	// fallback to file extension
	strParts := strings.Split(enclosure.URL, ".")
	extension := strParts[len(strParts)-1]
	if mimeType, ok := extToMIME[extension]; ok {
		return mimeType, nil
	}

	return "", fmt.Errorf(
		"unable to detect MIME type, enclosure type: '%s', url: '%s'",
		enclosure.Type,
		enclosure.URL,
	)
}

func MIMETypeExtension(mimeType string) (string, error) {
	extension, ok := mimeToExt[mimeType]
	if !ok {
		return "", fmt.Errorf("unsupported MIME type '%s'", mimeType)
	}
	return extension, nil
}
