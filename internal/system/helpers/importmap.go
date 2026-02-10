package helpers

import (
	"log/slog"

	"github.com/MateoCaicedoW/gomp3/internal/system/assets"
)

func Importmap() string {
	importMap, err := assets.Manager.ImportMap()
	if err != nil {
		slog.Error("error getting import map", "error", err)
		return ""
	}
	return string(importMap)
}
