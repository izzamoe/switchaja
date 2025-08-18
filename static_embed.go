package switchiot

import "embed"

//go:embed web/static/*
var EmbeddedStatic embed.FS
