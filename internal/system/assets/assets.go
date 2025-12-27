package assets

import (
	"embed"

	"go.leapkit.dev/core/assets"
)

//go:embed *
var Files embed.FS

var Manager = assets.NewManager(Files, "/public")
