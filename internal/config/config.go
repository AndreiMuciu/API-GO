package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
    Port     string
    MongoURI string
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

    uri = strings.Replace(uri, "<db_password>", password, 1)
    return &Config{
        Port:     port,
        MongoURI: uri,
    }, nil
}