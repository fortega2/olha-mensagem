package server

import (
	"io/fs"
	"net/http"

	"github.com/fortega2/real-time-chat/internal/frontend"
)

func (s *Server) serveStaticFiles() http.Handler {
	distFS, err := fs.Sub(frontend.StaticFiles, "olha-mensagem-app/build")
	if err != nil {
		s.logger.Fatal("Failed to create static assets subdirectory", "error", err)
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := distFS.Open(r.URL.Path[1:])
		if err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
