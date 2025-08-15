package router

import (
	"API-GO/internal/handlers"
	"API-GO/internal/repository"

	"github.com/gorilla/mux"
)

func NewBooksRouter(repo repository.BookRepository) *mux.Router {
    r := mux.NewRouter()
    h := handlers.NewBooksHandler(repo)
    MountCRUD(r, "/books", h)
    return r
}
