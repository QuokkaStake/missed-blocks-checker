package postgres

import "embed"

//go:embed **.sql
var EmbedFS embed.FS
