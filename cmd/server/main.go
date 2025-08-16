package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"API-GO/internal/config"
	"API-GO/internal/database"
	"API-GO/internal/logger"
	"API-GO/internal/middleware"
	"API-GO/internal/repository"
	"API-GO/internal/router"
	"API-GO/internal/services"
	"API-GO/internal/utils"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main(){
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	// Configure logging level
	logger.SetLevelFromString(cfg.LogLevel)

	db, err := database.Connect(cfg.MongoURI)
	if err != nil {
		panic(err)
	}
	// Creează indecși unici
	if err := database.CreateIndexes(db); err != nil {
    	log.Printf("Warning: failed to create indexes: %v", err)
	}
	defer func (){
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
       	if err := db.Disconnect(ctx); err != nil {
           	log.Printf("db disconnect error: %v", err)
       	}
	}()


	// Root router
	root := mux.NewRouter()
	root.Use(middleware.RequestLogger)

	// Repositories & services
	userRepo := repository.NewMongoUserRepository(db)
	jwtManager := &utils.JWTManager{Secret: []byte(cfg.JWTSecret), AccessTTL: time.Duration(cfg.JWTTTLMinutes) * time.Minute, CookieName: cfg.CookieName, SecureCookies: cfg.CookieSecure}
	authSvc := services.NewAuthService(userRepo, jwtManager)

	// Routere
	userRouter := router.NewUsersRouter(userRepo) // CRUD users prin repository
	// Books repository & router
	bookRepo := repository.NewMongoBookRepository(db)
	bookRouter := router.NewBooksRouter(bookRepo)
	authRouter := router.NewAuthRouter(authSvc, cfg.CookieName, cfg.CookieSecure)

	// Montează distinct pentru a evita conflictul dintre două PathPrefix identice
	root.PathPrefix("/api-go/v1/users").Handler(http.StripPrefix("/api-go/v1", userRouter))
	root.PathPrefix("/api-go/v1/books").Handler(http.StripPrefix("/api-go/v1", bookRouter))
	root.PathPrefix("/api-go/v1/auth").Handler(http.StripPrefix("/api-go/v1", authRouter))

    // Pentru viitor - alte routere
    // productRouter := productRouter.New(db)
    // applyAPIPrefix(root, productRouter)

	log.Printf("Server running on port %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, root); err != nil {
        log.Fatalf("server error: %v", err)
    }

}