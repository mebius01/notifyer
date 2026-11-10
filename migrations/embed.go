// Package migrations embeds SQL migrations into the binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
