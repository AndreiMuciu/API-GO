package router

import (
	"API-GO/internal/handlers"
	"API-GO/internal/services"

	"github.com/gorilla/mux"
)

func NewAuthRouter(svc *services.AuthService, cookieName string, secure bool) *mux.Router {
    r := mux.NewRouter()
    h := handlers.NewAuthHandler(svc, cookieName, secure)

    r.HandleFunc("/auth/signup", h.Signup()).Methods("POST")
    r.HandleFunc("/auth/login", h.Login()).Methods("POST")
    r.HandleFunc("/auth/logout", h.Logout()).Methods("POST")

    return r
}
