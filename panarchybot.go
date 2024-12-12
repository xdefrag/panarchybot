package panarchybot

import "embed"

//go:embed migrations/*.sql
var EmbedMigrations embed.FS
