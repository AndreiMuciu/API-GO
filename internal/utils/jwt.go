package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    Secret         []byte
    AccessTTL      time.Duration
    CookieName     string
    SecureCookies  bool
}

type UserClaims struct {
    UserID string `json:"uid"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func (m *JWTManager) GenerateToken(userID, email string) (string, time.Time, error) {
    now := time.Now()
    exp := now.Add(m.AccessTTL)
    claims := UserClaims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(exp),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := token.SignedString(m.Secret)
    return s, exp, err
}

func (m *JWTManager) ParseToken(tokenStr string) (*UserClaims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
        return m.Secret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
        return claims, nil
    }
    return nil, jwt.ErrTokenInvalidClaims
}
