package util

import (
	"errors"
	"net"
	"net/url"
	"regexp"
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

	if u.Hostname() == "localhost" {
		return errors.New("URL host must not be localhost")
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
