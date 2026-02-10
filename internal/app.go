// Package internal integrates the app and loads general settings.
package internal

import (
	"cmp"
	"net/http"
	"os"

	"github.com/MateoCaicedoW/gomp3/internal/converter"
	"github.com/MateoCaicedoW/gomp3/internal/system/assets"

	"go.leapkit.dev/core/server"
)

var (

	// Server configuration variables loaded from
	host          = cmp.Or(os.Getenv("HOST"), "0.0.0.0")
	port          = cmp.Or(os.Getenv("PORT"), "3000")
	sessionSecret = cmp.Or(os.Getenv("SESSION_SECRET"), "d720c059-9664-4980-8169-1158e167ae57")
	sessionName   = cmp.Or(os.Getenv("SESSION_NAME"), "leapkit_session")
)

// New creates the http handler using the Leapkit server package
// and returns it with the address it is listening on.
func New() (http.Handler, string) {
	// Creating a new server instance with the host and port
	// variables read from the environment or default values.
	r := server.New(
		server.WithHost(host),
		server.WithPort(port),
		server.WithSession(sessionSecret, sessionName),

		// Mounting the assets folder in the /assets URL path
		server.WithAssets(assets.Files, "/internal/system/assets"),
	)

	// Defining the routes in the application.
	r.HandleFunc("GET /{$}", converter.Index)
	r.HandleFunc("POST /convert", converter.Convert)

	r.Folder(assets.Manager.HandlerPattern(), assets.Manager)
	return r.Handler(), r.Addr()
}
