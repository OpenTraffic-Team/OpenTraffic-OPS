package static

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// Handler returns an http.Handler that serves the frontend static files.
//
// Production: uses go:embed (the dist directory must exist at compile time).
// Development: if RTM_STATIC_DIR env var is set, serves from disk instead,
//              so you don't need to copy dist into backend/pkg/static/dist
//              every time the frontend changes.
func Handler() http.Handler {
	var rootFS http.FileSystem

	if devDir := os.Getenv("RTM_STATIC_DIR"); devDir != "" {
		rootFS = http.Dir(devDir)
	} else {
		distFS, err := fs.Sub(dist, "dist")
		if err != nil {
			panic(err)
		}
		rootFS = http.FS(distFS)
	}

	fileServer := http.FileServer(rootFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Guard: API and health endpoints should never be served by the static handler.
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to open the requested file.
		f, err := rootFS.Open(path)
		if err != nil {
			// File does not exist -> fallback to index.html for SPA routes.
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		// If the path points to a directory, also fallback to index.html.
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		// Real file: serve it as-is.
		fileServer.ServeHTTP(w, r)
	})
}
