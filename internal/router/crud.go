package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

// CRUDHandlers defines the minimal contract for RESTful CRUD controllers.
type CRUDHandlers interface {
    GetAll() http.HandlerFunc
    GetOne() http.HandlerFunc
    Create() http.HandlerFunc
    Update() http.HandlerFunc
    Delete() http.HandlerFunc
}

// MountCRUD wires standard CRUD routes under the given base path.
func MountCRUD(r *mux.Router, base string, h CRUDHandlers) {
    r.HandleFunc(base, h.GetAll()).Methods("GET")
    r.HandleFunc(base+"/{id}", h.GetOne()).Methods("GET")
    r.HandleFunc(base, h.Create()).Methods("POST")
    r.HandleFunc(base+"/{id}", h.Update()).Methods("PUT")
    r.HandleFunc(base+"/{id}", h.Delete()).Methods("DELETE")
}
