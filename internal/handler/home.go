package handler

import (
	"net/http"

	"github.com/ttani03/gotha-boilerplate/web/templates/pages"
)

// HomeHandler handles the home page.
type HomeHandler struct{}

// NewHomeHandler creates a new HomeHandler.
func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

// Index renders the home page.
func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	render(w, r, pages.Home())
}
