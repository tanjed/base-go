// Package migrations embeds the goose SQL migrations: add NNNN_name.sql files here (expand, then
// contract: a migration must be safe while older replicas still serve). None yet.
package migrations

import "embed"

//go:embed *
var FS embed.FS
