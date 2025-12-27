package converter

import (
	"net/http"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	l := layout{
		Title:       "GoMP3",
		Description: "GoMP3 is a simple youtube to mp3 converter built with Go.",
		Yield:       indexEl(),
	}.New()

	if err := l.Render(w); err != nil {
		server.Errorf(w, http.StatusInternalServerError, "error rendering layour %w", err)
		return
	}
}
