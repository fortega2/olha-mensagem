package handlers

import "net/http"

func (h *Handler) RootPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/templates/index.html")
}
