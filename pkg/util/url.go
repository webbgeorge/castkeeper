package util

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

func ValidateExtURL(uStr string) error {
	u, err := url.Parse(uStr)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid scheme in URL")
	}

	if u.Host == "" {
		return errors.New("URL host must not be empty")
	}

	// no localhost, or IPv6 (e.g. "[::1]")
	if u.Hostname() == "localhost" || strings.Contains(u.Hostname(), "[") {
		return errors.New("URL host must not be localhost or IP address")
	}

	if ip := net.ParseIP(u.Hostname()); ip != nil {
		return errors.New("URL host cannot be IP address")
	}

	return nil
}

func SanitiseGUID(guid string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9=_-]+")
	return reg.ReplaceAllString(guid, "-")
}
