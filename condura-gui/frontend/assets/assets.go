// Package assets holds the embedded static assets for the Wails desktop GUI.
//
// The Wails shell (condura-gui/shell) imports this package and passes
// its embedded FS to the asset server. The package lives next to dist/
// so the embed pattern can be a simple relative path — Go's //go:embed
// forbids ".." in patterns, so the embed must live in a directory that
// is a parent of the files it embeds.
package assets

import "embed"

//go:embed all:dist
var FS embed.FS
