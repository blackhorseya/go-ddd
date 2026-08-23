// Package identity embeds the Identity bounded context's SQL schema migrations.
//
// //go:embed is rooted at its own package directory and cannot reach outside it,
// so each bounded context keeps this thin package next to its .sql files and
// exposes them to the persistence layer that runs them.
package identity

import "embed"

// FS holds the embedded .sql migration files of this directory.
//
//go:embed *.sql
var FS embed.FS
