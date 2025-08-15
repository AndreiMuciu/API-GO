package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"API-GO/internal/models"
	"API-GO/internal/repository"
	"API-GO/internal/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsersHandler lucrează prin repository pentru consistență și testabilitate
type UsersHandler struct {
    Repo repository.UserRepository
}

func NewUsersHandler(repo repository.UserRepository) *UsersHandler {
    return &UsersHandler{Repo: repo}
}

// GetAllUsers returnează toți userii
func (h *UsersHandler) GetAllUsers() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        users, err := h.Repo.List(ctx)
        if err != nil {
            utils.WriteInternalServerError(w, "failed to fetch users", err.Error())
            return
        }
        for i := range users {
            users[i].Password = ""
        }
        utils.WriteSuccess(w, "users retrieved successfully", users)
    }
}

// GetUser returnează un user după ID
func (h *UsersHandler) GetUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteBadRequest(w, "invalid user ID format")
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        user, err := h.Repo.GetByID(ctx, objID)
        if err != nil {
            utils.WriteNotFound(w, "user not found")
            return
        }
        user.Password = ""
        utils.WriteSuccess(w, "user retrieved successfully", user)
    }
}

// CreateUser a fost dezactivat: crearea de utilizatori se face doar prin /auth/signup
func (h *UsersHandler) CreateUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    utils.WriteMethodNotAllowed(w, "creating users via this endpoint is disabled; use /auth/signup")
    }
}

// UpdateUser modifică un user existent
func (h *UsersHandler) UpdateUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteBadRequest(w, "invalid user ID format")
            return
        }
        var update models.UserInput
        if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
            utils.WriteBadRequest(w, "invalid request body", err.Error())
            return
        }
        fields := map[string]interface{}{}
        // Validare + verificări în paralel pentru email și telefon
        type chk struct{ exists bool; err error }
        var emailChk, phoneChk chk
        var wgV sync.WaitGroup
        // Email
        if update.Email != "" {
            if !utils.IsValidEmail(update.Email) {
                utils.WriteBadRequest(w, "invalid email format")
                return
            }
            wgV.Add(1)
            go func(email string) {
                defer wgV.Done()
                ex, er := h.Repo.EmailExists(r.Context(), email, objID)
                emailChk = chk{exists: ex, err: er}
            }(update.Email)
            fields["email"] = update.Email
        }
        // Phone
        if update.Phone != "" {
            if !utils.IsValidPhone(update.Phone) {
                utils.WriteBadRequest(w, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
                return
            }
            wgV.Add(1)
            go func(phone string) {
                defer wgV.Done()
                ex, er := h.Repo.PhoneExists(r.Context(), phone, objID)
                phoneChk = chk{exists: ex, err: er}
            }(update.Phone)
            fields["phone"] = update.Phone
        }
        wgV.Wait()
        if emailChk.err != nil {
            utils.WriteInternalServerError(w, "failed to check email uniqueness", emailChk.err.Error())
            return
        }
        if phoneChk.err != nil {
            utils.WriteInternalServerError(w, "failed to check phone uniqueness", phoneChk.err.Error())
            return
        }
        if emailChk.exists {
            utils.WriteConflict(w, "email already exists")
            return
        }
        if phoneChk.exists {
            utils.WriteConflict(w, "phone number already exists")
            return
        }
        // Validare nume
        if update.Name != "" {
            if len(update.Name) < 2 {
                utils.WriteBadRequest(w, "name must be at least 2 characters")
                return
            }
            if len(update.Name) > 50 {
                utils.WriteBadRequest(w, "name cannot exceed 50 characters")
                return
            }
            fields["name"] = update.Name
        }
        if len(fields) == 0 {
            utils.WriteBadRequest(w, "no valid fields to update")
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        ok, err := h.Repo.UpdateFields(ctx, objID, fields)
        if err != nil {
            utils.WriteInternalServerError(w, "failed to update user", err.Error())
            return
        }
        if !ok {
            utils.WriteNotFound(w, "user not found")
            return
        }
        utils.WriteSuccess(w, "user updated successfully", nil)
    }
}

// DeleteUser șterge un user după ID
func (h *UsersHandler) DeleteUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteBadRequest(w, "invalid user ID format")
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        ok, err := h.Repo.DeleteByID(ctx, objID)
        if err != nil {
            utils.WriteInternalServerError(w, "failed to delete user", err.Error())
            return
        }
        if !ok {
            utils.WriteNotFound(w, "user not found")
            return
        }
        utils.WriteSuccess(w, "user deleted successfully", nil)
    }
}