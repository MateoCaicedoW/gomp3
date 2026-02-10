package converter

import (
	"fmt"
	"net/http"

	"github.com/MateoCaicedoW/gomp3/internal/system/services/mp3"
	"go.leapkit.dev/core/server"
)

func Convert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		server.Errorf(w, http.StatusBadRequest, "failed to parse form: %w", err)
		return
	}

	videoURL := r.FormValue("youtube-url")
	if videoURL == "" {
		server.Errorf(w, http.StatusBadRequest, "youtube-url is required")
		return
	}

	// Get video info first for the filename
	svc := mp3.New()
	info, err := svc.GetVideoInfo(videoURL)
	if err != nil {
		server.Errorf(w, http.StatusInternalServerError, "failed to get video info: %w", err)
		return
	}

	sanitizedTitle := mp3.SanitizeFilename(info.Title)

	// Set headers before starting conversion
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.mp3\"", sanitizedTitle))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Cache-Control", "no-cache")

	// Stream directly to response writer using the service
	if err := svc.ConvertToWriter(r.Context(), videoURL, w, nil); err != nil {
		server.Errorf(w, http.StatusInternalServerError, "conversion failed: %w", err)
		return
	}
}
