package e2e

import (
	"testing"
)

const baseURL = "http://localhost:8081"

func TestLogin(t *testing.T) {
	page := browser.MustIncognito().MustPage(baseURL)
	defer page.MustClose()

	page.MustElementR("input", "Username").MustInput("test")
	page.MustElementR("input", "Password").MustInput("123456")
	page.MustElementR("button", "Login").MustClick()

	page.MustScreenshot("my.png")
}
