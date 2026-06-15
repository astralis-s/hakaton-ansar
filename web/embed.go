// Package web embeds the compiled single-page frontend and serves it (with the
// vendored assets) from the Go binary, so the whole product ships as one
// container. API routes (/api/*, /swagger/*, /health) are registered before this
// handler, so it only ever serves the SPA shell and static assets.
package web

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed index.html assets
var content embed.FS

var fileServer = http.FileServer(http.FS(content))

// Handler serves embedded static files; any unknown path returns index.html so
// client-side navigation works.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" {
			if f, err := content.Open(p); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		serveIndex(w)
	})
}

func serveIndex(w http.ResponseWriter) {
	data, err := content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "frontend not built", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}
