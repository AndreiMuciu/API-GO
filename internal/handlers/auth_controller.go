package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"API-GO/internal/logger"
	"API-GO/internal/models"
	"API-GO/internal/services"
	"API-GO/internal/utils"
)

type AuthHandler struct {
    Svc *services.AuthService
    CookieName    string
    SecureCookies bool
}

func NewAuthHandler(svc *services.AuthService, cookieName string, secure bool) *AuthHandler {
    return &AuthHandler{Svc: svc, CookieName: cookieName, SecureCookies: secure}
}

func (h *AuthHandler) Signup() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in models.AuthSignUpRequest
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            utils.WriteBadRequest(w, "invalid request body", err.Error())
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        user, token, exp, err := h.Svc.SignUp(ctx, in)
        if err != nil {
            logger.Warnf("signup_failed", logger.Fields{"error": err.Error(), "email": in.Email})
            switch err.Error() {
            case "email already exists", "phone number already exists":
                utils.WriteConflict(w, err.Error())
            case "passwords do not match", "invalid email format", "invalid phone format", "name, email and password are required":
                utils.WriteBadRequest(w, err.Error())
            default:
                utils.WriteInternalServerError(w, "failed to sign up", err.Error())
            }
            return
        }
        setAuthCookie(w, h.CookieName, token, exp, h.SecureCookies)
        logger.Infof("signup_success", logger.Fields{"user_id": user.ID, "email": user.Email})
        utils.WriteCreated(w, "signed up successfully", user)
    }
}

func (h *AuthHandler) Login() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var in models.AuthLoginRequest
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            utils.WriteBadRequest(w, "invalid request body", err.Error())
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        user, token, exp, err := h.Svc.Login(ctx, in)
        if err != nil {
            logger.Warnf("login_failed", logger.Fields{"error": err.Error(), "email": in.Email})
            utils.WriteUnauthorized(w, "invalid credentials")
            return
        }
        setAuthCookie(w, h.CookieName, token, exp, h.SecureCookies)
        logger.Infof("login_success", logger.Fields{"user_id": user.ID, "email": user.Email})
        utils.WriteSuccess(w, "logged in successfully", user)
    }
}

func (h *AuthHandler) Logout() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Clear cookie
        http.SetCookie(w, &http.Cookie{
            Name:     h.CookieName,
            Value:    "",
            Path:     "/",
            Expires:  time.Unix(0, 0),
            MaxAge:   -1,
            HttpOnly: true,
            Secure:   h.SecureCookies,
            SameSite: http.SameSiteLaxMode,
        })
        logger.Infof("logout", logger.Fields{})
        utils.WriteNoContent(w)
    }
}

func setAuthCookie(w http.ResponseWriter, name, token string, exp time.Time, secure bool) {
    http.SetCookie(w, &http.Cookie{
        Name:     name,
        Value:    token,
        Path:     "/",
        Expires:  exp,
        HttpOnly: true,
        Secure:   secure,
        SameSite: http.SameSiteLaxMode,
    })
}
