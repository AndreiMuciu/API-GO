package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"API-GO/internal/models"
	"API-GO/internal/repository"
	"API-GO/internal/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BooksHandler struct {
    Repo repository.BookRepository
}

func NewBooksHandler(repo repository.BookRepository) *BooksHandler {
    return &BooksHandler{Repo: repo}
}

func (h *BooksHandler) GetAll() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        // allowed fields for filtering and sorting
        allowed := map[string]string{
            "title": "string",
            "author": "string",
            "genre": "string",
            "yearPublished": "int",
        }
        allowedSort := map[string]bool{ "title": true, "author": true, "genre": true, "yearPublished": true }
        q := utils.ParseListQuery(r, allowed, allowedSort, "title", 20, 100)
        items, total, err := h.Repo.ListWithQuery(ctx, q)
        if err != nil { utils.WriteInternalServerError(w, "failed to fetch books", err.Error()); return }
        resp := map[string]interface{}{
            "items": items,
            "page":  q.Page,
            "limit": q.Limit,
            "total": total,
        }
        utils.WriteSuccess(w, "books retrieved successfully", resp)
    }
}

func (h *BooksHandler) GetOne() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        oid, err := primitive.ObjectIDFromHex(idParam)
        if err != nil { utils.WriteBadRequest(w, "invalid book ID format"); return }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        item, err := h.Repo.GetByID(ctx, oid)
        if err != nil { utils.WriteNotFound(w, "book not found"); return }
        utils.WriteSuccess(w, "book retrieved successfully", item)
    }
}

func (h *BooksHandler) Create() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in models.Book
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil { utils.WriteBadRequest(w, "invalid request body", err.Error()); return }
        if in.Title == "" || in.Author == "" { utils.WriteBadRequest(w, "title and author are required"); return }
        if in.YearPublished < 0 { utils.WriteBadRequest(w, "yearPublished must be positive"); return }
        in.ID = primitive.NewObjectID()
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        if err := h.Repo.Create(ctx, &in); err != nil { utils.WriteInternalServerError(w, "failed to create book", err.Error()); return }
        utils.WriteCreated(w, "book created successfully", in)
    }
}

func (h *BooksHandler) Update() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        oid, err := primitive.ObjectIDFromHex(idParam)
        if err != nil { utils.WriteBadRequest(w, "invalid book ID format"); return }
        var payload map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { utils.WriteBadRequest(w, "invalid request body", err.Error()); return }
        // basic validation
        if v, ok := payload["yearPublished"]; ok {
            switch vv := v.(type) {
            case float64:
                if vv < 0 { utils.WriteBadRequest(w, "yearPublished must be positive"); return }
                payload["yearPublished"] = int(vv)
            }
        }
        delete(payload, "id")
        if len(payload) == 0 { utils.WriteBadRequest(w, "no valid fields to update"); return }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        ok, err := h.Repo.UpdateFields(ctx, oid, payload)
        if err != nil { utils.WriteInternalServerError(w, "failed to update book", err.Error()); return }
        if !ok { utils.WriteNotFound(w, "book not found"); return }
        utils.WriteSuccess(w, "book updated successfully", nil)
    }
}

func (h *BooksHandler) Delete() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        oid, err := primitive.ObjectIDFromHex(idParam)
        if err != nil { utils.WriteBadRequest(w, "invalid book ID format"); return }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        ok, err := h.Repo.DeleteByID(ctx, oid)
        if err != nil { utils.WriteInternalServerError(w, "failed to delete book", err.Error()); return }
        if !ok { utils.WriteNotFound(w, "book not found"); return }
        utils.WriteSuccess(w, "book deleted successfully", nil)
    }
}
