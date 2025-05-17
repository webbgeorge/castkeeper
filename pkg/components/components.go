package components

// this dependency is only used in .templ files,
// so is ignored by GitHub dependabot - causing
// it to be removed in those PRs, breaking the
// build.
import _ "github.com/microcosm-cc/bluemonday"
