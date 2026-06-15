// Package migrations embeds the goose SQL migration files so they ship inside
// the single binary and run on boot (ARCHITECTURE.md: "миграции goose эмбедятся
// в бинарь и прогоняются на старте").
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
