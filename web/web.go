package web

import "embed"

// This file embed static assets for the web server.
// go embed requires to be above the level of the embedded assets.

//go:embed static
var StaticAssets embed.FS
