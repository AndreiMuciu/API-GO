package services

import (
	"context"
	"errors"
	"time"

	"API-GO/internal/models"
	"API-GO/internal/repository"
	"API-GO/internal/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
    Users repository.UserRepository
    JWT   *utils.JWTManager
}

func NewAuthService(users repository.UserRepository, jwt *utils.JWTManager) *AuthService {
    return &AuthService{Users: users, JWT: jwt}
}

// SignUp creează un utilizator nou
func (s *AuthService) SignUp(ctx context.Context, in models.AuthSignUpRequest) (*models.AuthResponse, string, time.Time, error) {
    if in.Name == "" || in.Email == "" || in.Password == "" {
        return nil, "", time.Time{}, errors.New("name, email and password are required")
    }
    if in.Password != in.PasswordConfirm {
        return nil, "", time.Time{}, errors.New("passwords do not match")
    }
    if !utils.IsValidEmail(in.Email) {
        return nil, "", time.Time{}, errors.New("invalid email format")
    }
    if in.Phone != "" && !utils.IsValidPhone(in.Phone) {
        return nil, "", time.Time{}, errors.New("invalid phone format")
    }

    exists, err := s.Users.EmailExists(ctx, in.Email)
    if err != nil {
        return nil, "", time.Time{}, err
    }
    if exists {
        return nil, "", time.Time{}, errors.New("email already exists")
    }
    if in.Phone != "" {
        pexists, err := s.Users.PhoneExists(ctx, in.Phone)
        if err != nil {
            return nil, "", time.Time{}, err
        }
        if pexists {
            return nil, "", time.Time{}, errors.New("phone number already exists")
        }
    }

    hashed, err := utils.HashPassword(in.Password)
    if err != nil {
        return nil, "", time.Time{}, err
    }

    u := models.User{
        ID:       primitive.NewObjectID(),
        Name:     in.Name,
        Email:    in.Email,
        Password: hashed,
        Phone:    in.Phone,
    }
    if err := s.Users.Create(ctx, &u); err != nil {
        return nil, "", time.Time{}, err
    }

    token, exp, err := s.JWT.GenerateToken(u.ID.Hex(), u.Email)
    if err != nil {
        return nil, "", time.Time{}, err
    }

    resp := &models.AuthResponse{ID: u.ID.Hex(), Name: u.Name, Email: u.Email, Phone: u.Phone}
    return resp, token, exp, nil
}

// Login autentifică un utilizator existent
func (s *AuthService) Login(ctx context.Context, in models.AuthLoginRequest) (*models.AuthResponse, string, time.Time, error) {
    if in.Email == "" || in.Password == "" {
        return nil, "", time.Time{}, errors.New("email and password are required")
    }
    u, err := s.Users.GetByEmail(ctx, in.Email)
    if err != nil {
        return nil, "", time.Time{}, errors.New("invalid credentials")
    }
    if !utils.CheckPassword(u.Password, in.Password) {
        return nil, "", time.Time{}, errors.New("invalid credentials")
    }
    token, exp, err := s.JWT.GenerateToken(u.ID.Hex(), u.Email)
    if err != nil {
        return nil, "", time.Time{}, err
    }
    resp := &models.AuthResponse{ID: u.ID.Hex(), Name: u.Name, Email: u.Email, Phone: u.Phone}
    return resp, token, exp, nil
}
