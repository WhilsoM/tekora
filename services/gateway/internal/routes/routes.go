package routes

import "github.com/go-chi/chi/v5"

func InitUsersHandlers(r chi.Router) {
	r.Post("/users", nil)
}
