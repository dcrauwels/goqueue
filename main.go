package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	db     *database.Queries
	secret string
}

func main() {
	// load .env into env variables
	godotenv.Load()
	//opendb
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Println(err)
		return
	}
	dbQueries := database.New(db)

	// set up apiConfig
	apiCfg := apiConfig{
		db:     dbQueries,
		secret: os.Getenv("SECRET"),
	}

	// servemux
	mux := http.NewServeMux()

	// register handlers
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	//handler_users.go
	mux.HandleFunc("POST /api/users", apiCfg.handlerPostUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerPutUsers)
	mux.HandleFunc("DELETE /api/users", apiCfg.handlerDeleteUsers)
	//handler_auth.go
	// login
	// refresh
	// revoke
	//handler_visitors.go

	// fileserver whenever

	// server
	s := http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		ReadTimeout:                  30 * time.Second,
		WriteTimeout:                 60 * time.Second,
		IdleTimeout:                  120 * time.Second,
	}

	err = s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}

}
