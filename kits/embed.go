// Package kits embeds the bundled design systems, one kit per directory. Each directory holds a
// manifest.json, D2 style files and assets; see README.md.
package kits

import "embed"

//go:embed plain carbon cloudscape fluent
var FS embed.FS
