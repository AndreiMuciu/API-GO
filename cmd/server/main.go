package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"API-GO/internal/config"
	"API-GO/internal/database"
	"API-GO/internal/router"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// applyAPIPrefix aplică prefix-ul /api-go/v1 la orice router
func applyAPIPrefix(r *mux.Router, handler http.Handler) {
    r.PathPrefix("/api-go/v1").Handler(http.StripPrefix("/api-go/v1", handler))
}

func main(){
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

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

    // Aplică prefix-ul la user router
    userRouter := router.New(db)
    applyAPIPrefix(root, userRouter)

    // Pentru viitor - alte routere
    // productRouter := productRouter.New(db)
    // applyAPIPrefix(root, productRouter)

    log.Printf("Server running on %s", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, root); err != nil {
        log.Fatalf("server error: %v", err)
    }

}