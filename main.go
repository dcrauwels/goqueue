package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dcrauwels/goqueue/admin"
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

	/// register handlers from api package
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", apiCfg.ReadinessHandler)
	//handler_users.go
	mux.HandleFunc("POST /api/users", apiCfg.HandlerPostUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.HandlerPutUsers)
	mux.HandleFunc("PUT /api/users/{user_id}", apiCfg.HandlerPutUsersByID) //NYI
	mux.HandleFunc("GET /api/users", apiCfg.HandlerGetUsers)
	mux.HandleFunc("GET /api/users/{user_id}", apiCfg.HandlerGetUsersByID)
	//mux.HandleFunc("DELETE /api/users", apiCfg.HandlerDeleteUsers) NYI do I even want this
	//handler_auth.go
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLoginUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefreshUser)
	mux.HandleFunc("POST /api/logout", apiCfg.HandlerLogoutUser)
	//handler_visitors.go
	mux.HandleFunc("POST /api/visitors", apiCfg.HandlerPostVisitors)
	mux.HandleFunc("PUT /api/visitors/{visitor_id}", apiCfg.HandlerPutVisitors) // NYI
	mux.HandleFunc("GET /api/visitors", apiCfg.HandlerGetVisitors)              // NYI
	//handler_servicelogs.go NYI
	/// register handlers from the admin package
	//handler_admin.go
	mux.HandleFunc("POST /admin/users", func(w http.ResponseWriter, r *http.Request) {
		admin.AdminCreateUser(w, r, apiCfg)
	}

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
