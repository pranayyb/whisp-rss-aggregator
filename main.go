package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pranayyb/whisp-rss-aggregator/internal/db"
)

type apiConfig struct {
	DB *db.Queries
}

func main() {
	fmt.Println("Hello, Whisp RSS Aggregator!")

	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT environment variable not set")
	}

	dbString := os.Getenv("DB_URL")
	if dbString == "" {
		log.Fatal("DB URL environment variable not set")
	}

	conn, err := sql.Open("postgres", dbString)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	apiCfg := apiConfig{
		DB: db.New(conn),
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1router := chi.NewRouter()
	v1router.Get("/readiness", handlerReadiness)
	v1router.Get("/err", handlerError)
	v1router.Post("/create_user", apiCfg.handlerCreateUser)
	v1router.Get("/get_user", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1router.Post("/create_feed", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1router.Get("/feeds", apiCfg.handlerGetFeed)

	router.Mount("/v1", v1router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	fmt.Println("Starting server on port", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
	fmt.Println("PORT: ", portString)
}
