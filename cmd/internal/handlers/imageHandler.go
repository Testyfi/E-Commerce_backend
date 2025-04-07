package handlers

import (
	"net/http"
	"path"

	"github.com/go-chi/chi"
)

func ServeImage(w http.ResponseWriter, r *http.Request) {
	image := chi.URLParam(r, "image")
	imagePath := path.Join("./assets", image)

	http.ServeFile(w, r, imagePath)
}
