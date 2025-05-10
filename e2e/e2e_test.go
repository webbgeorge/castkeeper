package e2e

import (
	"testing"
	"time"

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
		page.MustReload() // just to make it easier to assert for success alert
		page.MustWaitDOMStable()

		page.MustElementR("button", "Add Feed URL").MustClick()
		page.MustWaitDOMStable()

		page.
			MustElementR("input", "Feed URL").
			MustInput("https://feeds.simplecast.com/Sp_45INC")
		page.MustElementR("button", "Add Podcast").MustClick()
		page.MustWaitDOMStable()

		assert.Equal(t, "Podcast added", page.MustElement(".alert").MustText())

		// close modal
		page.MustElementR("button", "✕").MustClick()
	})

	// ensure require queue jobs are processed before proceeding with test
	time.Sleep(time.Second * 2)

	t.Run("list_podcasts", func(t *testing.T) {
		page.MustElementR("a", "CastKeeper").MustClick()
		page.MustWaitDOMStable()

		page.MustElementR("h2", "How Do You Cope\\? …with Elis and John")
		page.MustElementR("h2", "Top Scallops")
	})

	t.Run("view_podcast", func(t *testing.T) {
		page.
			MustElementR("h2", "How Do You Cope\\? …with Elis and John").
			MustParent().
			MustElementR("a", "View").
			MustClick()
		page.MustWaitDOMStable()

		episodeRow := page.
			MustElementR("td", "How Do You Cope - Series 4 teaser").
			MustParent()

		assert.Equal(t, "pending", episodeRow.MustElement(".badge").MustText())
		page.MustElementR("td", "3m7s")
	})
}
