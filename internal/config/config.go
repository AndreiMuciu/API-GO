package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
    Port     string
    MongoURI string
    JWTSecret string
    JWTTTLMinutes int
    CookieName string
    CookieSecure bool
}

func Load() (*Config, error) {
    uri := os.Getenv("MONGO_URI")
    if uri == "" {
        return nil, fmt.Errorf("MONGO_URI environment variable is not set")
    }
    password := os.Getenv("DB_PASSWORD")
    if password == "" {
        return nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
    }
    port := os.Getenv("PORT")
    if port == "" {
        return nil, fmt.Errorf("environment variable PORT is not set")          
    }
    // dacă nu începe cu ":", adaugă prefixul
    if !strings.HasPrefix(port, ":") {
        port = ":" + port
    }

    // JWT secret
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
    }
    // TTL (minutes), default 60
    jwtTTL := 60
    if v := os.Getenv("JWT_TTL_MINUTES"); v != "" {
        var parsed int
        fmt.Sscanf(v, "%d", &parsed)
        if parsed > 0 {
            jwtTTL = parsed
        }
    }
    // Cookie name default "access_token"
    cookieName := os.Getenv("COOKIE_NAME")
    if cookieName == "" {
        cookieName = "access_token"
    }
    // Cookie secure flag (default false for local dev)
    cookieSecure := false
    if v := os.Getenv("COOKIE_SECURE"); strings.ToLower(v) == "true" || v == "1" {
        cookieSecure = true
    }

    uri = strings.Replace(uri, "<db_password>", password, 1)
    return &Config{
        Port:     port,
        MongoURI: uri,
        JWTSecret: jwtSecret,
        JWTTTLMinutes: jwtTTL,
        CookieName: cookieName,
        CookieSecure: cookieSecure,
    }, nil
}