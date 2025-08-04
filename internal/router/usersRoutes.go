package router

import (
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"

	"API-GO/internal/handlers"
)

// New construieşte un *mux.Router, înregistrează rutele de users şi îl returnează
func New(db *mongo.Client) *mux.Router {
    r := mux.NewRouter()

    // GET    /users
    r.HandleFunc("/users", handlers.GetAllUsers(db)).Methods("GET")
    // GET    /users/{id}
    r.HandleFunc("/users/{id}", handlers.GetUser(db)).Methods("GET")
    // POST   /users
    r.HandleFunc("/users", handlers.CreateUser(db)).Methods("POST")
    // PUT    /users/{id}
    r.HandleFunc("/users/{id}", handlers.UpdateUser(db)).Methods("PUT")
    // DELETE /users/{id}
    r.HandleFunc("/users/{id}", handlers.DeleteUser(db)).Methods("DELETE")

    return r
}