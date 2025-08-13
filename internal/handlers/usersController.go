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

// CreateUser inserează un nou user
func (h *UsersHandler) CreateUser() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in models.UserInput
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            utils.WriteBadRequest(w, "invalid request body", err.Error())
            return
        }
        if in.Name == "" || in.Email == "" || in.Password == "" {
            utils.WriteBadRequest(w, "name, email and password are required")
            return
        }
        if !utils.IsValidEmail(in.Email) {
            utils.WriteBadRequest(w, "invalid email format")
            return
        }
        if in.Phone != "" && !utils.IsValidPhone(in.Phone) {
            utils.WriteBadRequest(w, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
            return
        }
        if len(in.Password) < 8 {
            utils.WriteBadRequest(w, "password must be at least 8 characters")
            return
        }
        if in.Password != in.PasswordConfirm {
            utils.WriteBadRequest(w, "passwords do not match")
            return
        }

        // Verificări în paralel pentru email/telefon
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        var wg sync.WaitGroup
        wg.Add(2)
        var emailExists, phoneExists bool
        var emailErr, phoneErr error
        go func() {
            defer wg.Done()
            emailExists, emailErr = h.Repo.EmailExists(ctx, in.Email)
        }()
        go func() {
            defer wg.Done()
            if in.Phone != "" {
                phoneExists, phoneErr = h.Repo.PhoneExists(ctx, in.Phone)
            }
        }()
        wg.Wait()
        if emailErr != nil || phoneErr != nil {
            if emailErr != nil {
                utils.WriteInternalServerError(w, "failed to check email uniqueness", emailErr.Error())
            } else {
                utils.WriteInternalServerError(w, "failed to check phone uniqueness", phoneErr.Error())
            }
            return
        }
        if emailExists {
            utils.WriteConflict(w, "email already exists")
            return
        }
        if phoneExists {
            utils.WriteConflict(w, "phone number already exists")
            return
        }

        hashed, err := utils.HashPassword(in.Password)
        if err != nil {
            utils.WriteInternalServerError(w, "failed to hash password", err.Error())
            return
        }
        user := models.User{
            ID:       primitive.NewObjectID(),
            Name:     in.Name,
            Email:    in.Email,
            Password: hashed,
            Phone:    in.Phone,
        }
        if err := h.Repo.Create(ctx, &user); err != nil {
            utils.WriteInternalServerError(w, "failed to create user", err.Error())
            return
        }
        user.Password = ""
        utils.WriteCreated(w, "user created successfully", user)
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
        // Validare și verificare email
        if update.Email != "" {
            if !utils.IsValidEmail(update.Email) {
                utils.WriteBadRequest(w, "invalid email format")
                return
            }
            exists, err := h.Repo.EmailExists(r.Context(), update.Email, objID)
            if err != nil {
                utils.WriteInternalServerError(w, "failed to check email uniqueness", err.Error())
                return
            }
            if exists {
                utils.WriteConflict(w, "email already exists")
                return
            }
            fields["email"] = update.Email
        }
        // Validare și verificare telefon
        if update.Phone != "" {
            if !utils.IsValidPhone(update.Phone) {
                utils.WriteBadRequest(w, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
                return
            }
            exists, err := h.Repo.PhoneExists(r.Context(), update.Phone, objID)
            if err != nil {
                utils.WriteInternalServerError(w, "failed to check phone uniqueness", err.Error())
                return
            }
            if exists {
                utils.WriteConflict(w, "phone number already exists")
                return
            }
            fields["phone"] = update.Phone
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