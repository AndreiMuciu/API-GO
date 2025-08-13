package router

import (
	"API-GO/internal/handlers"
	"API-GO/internal/repository"

	"github.com/gorilla/mux"
)

// NewUsersRouter construieşte routerul de users folosind repository
func NewUsersRouter(repo repository.UserRepository) *mux.Router {
    r := mux.NewRouter()
    h := handlers.NewUsersHandler(repo)

    r.HandleFunc("/users", h.GetAllUsers()).Methods("GET")
    r.HandleFunc("/users/{id}", h.GetUser()).Methods("GET")
    r.HandleFunc("/users", h.CreateUser()).Methods("POST")
    r.HandleFunc("/users/{id}", h.UpdateUser()).Methods("PUT")
    r.HandleFunc("/users/{id}", h.DeleteUser()).Methods("DELETE")

    return r
}