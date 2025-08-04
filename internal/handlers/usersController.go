package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"API-GO/internal/database"
	"API-GO/internal/models"
	"API-GO/internal/utils"
)

// GetAllUsers returnează toți userii din colecție
func GetAllUsers(client *mongo.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        coll := database.UserCollection(client)
        cursor, err := coll.Find(ctx, bson.M{})
        if err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to fetch users", err.Error())
            return
        }
        defer cursor.Close(ctx)

        var users []models.User
        if err := cursor.All(ctx, &users); err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to decode users", err.Error())
            return
        }

        // Ascunde parolele din toate răspunsurile
        for i := range users {
            users[i].Password = ""
        }

        utils.WriteSuccess(w, "users retrieved successfully", users)
    }
}

// GetUser returnează un user după ID
func GetUser(client *mongo.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteError(w, http.StatusBadRequest, "invalid user ID format")
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        var user models.User
        coll := database.UserCollection(client)
        if err := coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
            if err == mongo.ErrNoDocuments {
                utils.WriteError(w, http.StatusNotFound, "user not found")
            } else {
                utils.WriteError(w, http.StatusInternalServerError, "failed to fetch user", err.Error())
            }
            return
        }

        // Ascunde parola
        user.Password = ""
        utils.WriteSuccess(w, "user retrieved successfully", user)
    }
}

// CreateUser inserează un nou user cu parolă hash-uită
func CreateUser(client *mongo.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in models.UserInput
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            utils.WriteError(w, http.StatusBadRequest, "invalid request body", err.Error())
            return
        }
        
        // Validare de bază
        if in.Name == "" || in.Email == "" || in.Password == "" {
            utils.WriteError(w, http.StatusBadRequest, "name, email and password are required")
            return
        }
        
        // Validare format email
        if !utils.IsValidEmail(in.Email) {
            utils.WriteError(w, http.StatusBadRequest, "invalid email format")
            return
        }
        
        // Validare format telefon
        if in.Phone != "" && !utils.IsValidPhone(in.Phone) {
            utils.WriteError(w, http.StatusBadRequest, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
            return
        }
        
        // Validare parolă
        if len(in.Password) < 8 {
            utils.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
            return
        }
        
        if in.Password != in.PasswordConfirm {
            utils.WriteError(w, http.StatusBadRequest, "passwords do not match")
            return
        }

        // Verifică unicitatea emailului
        emailExists, err := database.CheckEmailExists(client, in.Email)
        if err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to check email uniqueness", err.Error())
            return
        }
        if emailExists {
            utils.WriteError(w, http.StatusConflict, "email already exists")
            return
        }

        // Verifică unicitatea telefonului (dacă este furnizat)
        if in.Phone != "" {
            phoneExists, err := database.CheckPhoneExists(client, in.Phone)
            if err != nil {
                utils.WriteError(w, http.StatusInternalServerError, "failed to check phone uniqueness", err.Error())
                return
            }
            if phoneExists {
                utils.WriteError(w, http.StatusConflict, "phone number already exists")
                return
            }
        }

        // hash password
        h, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
        if err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to hash password")
            return
        }

        user := models.User{
            ID:       primitive.NewObjectID(),
            Name:     in.Name,
            Email:    in.Email,
            Password: string(h),
            Phone:    in.Phone,
        }

        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        coll := database.UserCollection(client)
        if _, err := coll.InsertOne(ctx, user); err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to create user", err.Error())
            return
        }

        // Ascunde parola din răspuns
        user.Password = ""
        utils.WriteCreated(w, "user created successfully", user)
    }
}

// UpdateUser modifică un user existent
func UpdateUser(client *mongo.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteError(w, http.StatusBadRequest, "invalid user ID format")
            return
        }

        var update models.UserInput
        if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
            utils.WriteError(w, http.StatusBadRequest, "invalid request body", err.Error())
            return
        }

        updateDoc := bson.M{}
        
        // Validare și verificare email
        if update.Email != "" {
            if !utils.IsValidEmail(update.Email) {
                utils.WriteError(w, http.StatusBadRequest, "invalid email format")
                return
            }
            
            emailExists, err := database.CheckEmailExists(client, update.Email, idParam)
            if err != nil {
                utils.WriteError(w, http.StatusInternalServerError, "failed to check email uniqueness", err.Error())
                return
            }
            if emailExists {
                utils.WriteError(w, http.StatusConflict, "email already exists")
                return
            }
            updateDoc["email"] = update.Email
        }

        // Validare și verificare telefon
        if update.Phone != "" {
            if !utils.IsValidPhone(update.Phone) {
                utils.WriteError(w, http.StatusBadRequest, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
                return
            }
            
            phoneExists, err := database.CheckPhoneExists(client, update.Phone, idParam)
            if err != nil {
                utils.WriteError(w, http.StatusInternalServerError, "failed to check phone uniqueness", err.Error())
                return
            }
            if phoneExists {
                utils.WriteError(w, http.StatusConflict, "phone number already exists")
                return
            }
            updateDoc["phone"] = update.Phone
        }

        // Validare nume
        if update.Name != "" {
            if len(update.Name) < 2 {
                utils.WriteError(w, http.StatusBadRequest, "name must be at least 2 characters")
                return
            }
            if len(update.Name) > 50 {
                utils.WriteError(w, http.StatusBadRequest, "name cannot exceed 50 characters")
                return
            }
            updateDoc["name"] = update.Name
        }

        if len(updateDoc) == 0 {
            utils.WriteError(w, http.StatusBadRequest, "no valid fields to update")
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        coll := database.UserCollection(client)
        res, err := coll.UpdateOne(
            ctx,
            bson.M{"_id": objID},
            bson.M{"$set": updateDoc},
        )
        if err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to update user", err.Error())
            return
        }
        if res.MatchedCount == 0 {
            utils.WriteError(w, http.StatusNotFound, "user not found")
            return
        }

        utils.WriteSuccess(w, "user updated successfully", nil)
    }
}

// DeleteUser șterge un user după ID
func DeleteUser(client *mongo.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idParam := mux.Vars(r)["id"]
        objID, err := primitive.ObjectIDFromHex(idParam)
        if err != nil {
            utils.WriteError(w, http.StatusBadRequest, "invalid user ID format")
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        coll := database.UserCollection(client)
        res, err := coll.DeleteOne(ctx, bson.M{"_id": objID})
        if err != nil {
            utils.WriteError(w, http.StatusInternalServerError, "failed to delete user", err.Error())
            return
        }
        if res.DeletedCount == 0 {
            utils.WriteError(w, http.StatusNotFound, "user not found")
            return
        }

        utils.WriteSuccess(w, "user deleted successfully", nil)
    }
}