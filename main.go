package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dcrauwels/goqueue/api"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	// set up ApiConfig
	apiCfg := api.ApiConfig{
		DB:     dbQueries,
		Secret: os.Getenv("SECRET"),
	}

	// servemux
	mux := http.NewServeMux()

	// register handlers from api package
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", apiCfg.ReadinessHandler)
	//handler_users.go
	mux.HandleFunc("POST /api/users", apiCfg.HandlerPostUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.HandlerPutUsers)
	//mux.HandleFunc("DELETE /api/users", apiCfg.HandlerDeleteUsers)
	//handler_auth.go
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLoginUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefresh)
	mux.HandleFunc("POST /api/logout", apiCfg.HandlerLogoutUser)
	//handler_visitors.go
	// create visitor
	// assign visitor to desk
	// set visitor complete

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
