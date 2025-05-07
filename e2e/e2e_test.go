package e2e

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
)

const debugModeEnabled = false

func TestE2E_SQLite_LocalObjects(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	browser, cleanup := setupE2ETests(
		configProfileSqlite,
		testing.Verbose(),
		debugModeEnabled,
	)
	t.Cleanup(cleanup)

	runE2ETests(t, browser, "http://localhost:8082")
}

// NOTE: this test requires that docker compose is running
func TestE2E_PostgreSQL_S3Objects(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	browser, cleanup := setupE2ETests(
		configProfilePostgres,
		testing.Verbose(),
		debugModeEnabled,
	)
	t.Cleanup(cleanup)

	runE2ETests(t, browser, "http://localhost:8083")
}

func runE2ETests(t *testing.T, browser *rod.Browser, baseURL string) {
	page := browser.
		MustIncognito().
		MustPage(baseURL).
		MustWindowFullscreen()
	t.Cleanup(page.MustClose)

	t.Run("login_unknown", func(t *testing.T) {
		assert.Equal(t, "Login", page.MustElement("h1").MustText())

		page.MustElementR("input", "Username").MustInput("not-user")
		page.MustElementR("input", "Password").MustInput("not-password")
		page.MustElementR("button", "Login").MustClick()

		page.MustWaitDOMStable()

		assert.Equal(t, "Unknown username or password", page.MustElement(".alert").MustText())
	})

	t.Run("login_success", func(t *testing.T) {
		page.MustElementR("input", "Username").MustInput("e2euser")
		page.MustElementR("input", "Password").MustInput("e2epass")
		page.MustElementR("button", "Login").MustClick()

		page.MustWaitDOMStable()

		assert.Equal(t, "Your Podcasts", page.MustElement("h1").MustText())
	})

	t.Run("search_podcasts", func(t *testing.T) {
		page.MustElementR("a", "Add a podcast").MustClick()

		page.MustWaitDOMStable()

		assert.Equal(t, "Add Podcast", page.MustElement("h1").MustText())

		page.MustElementR("input", "Search").MustInput("elis and john")
		page.MustElementR("button", "Search").MustClick()

		page.MustWaitDOMStable()

		page.MustElementR("h2", "How Do You Cope\\? …with Elis and John")
	})

	t.Run("add_podcast_from_search_results", func(t *testing.T) {
		page.
			MustElementR("h2", "How Do You Cope\\? …with Elis and John").
			MustParent().
			MustElementR("button", "Add Podcast").
			MustClick()

		page.MustWaitDOMStable()

		assert.Equal(t, "Podcast added", page.MustElement(".alert").MustText())
	})

	t.Run("add_podcast_from_url", func(t *testing.T) {
		// TODO
		t.Skip("Not implemented")
	})

	t.Run("list_podcasts", func(t *testing.T) {
		// TODO
		t.Skip("Not implemented")
	})

	t.Run("view_podcast", func(t *testing.T) {
		// TODO
		t.Skip("Not implemented")
	})
}
