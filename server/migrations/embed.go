// Package migrations embeds the SQL migration files so the server can apply
// them at startup without depending on the working directory.
package migrations

import "embed"

// FS holds the ordered *.sql migration files.
//
//go:embed *.sql
var FS embed.FS
