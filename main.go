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
		Env:    os.Getenv("ENV"),
	}

	// servemux
	mux := http.NewServeMux()

	/// register handlers from api package
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", apiCfg.ReadinessHandler)
	//handler_users.go
	mux.HandleFunc("POST /api/users", apiCfg.HandlerPostUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.HandlerPutUsers)
	mux.HandleFunc("PUT /api/users/{user_id}", apiCfg.HandlerPutUsersByID)
	mux.HandleFunc("GET /api/users", apiCfg.HandlerGetUsers)
	mux.HandleFunc("GET /api/users/{user_id}", apiCfg.HandlerGetUsersByID)
	//mux.HandleFunc("DELETE /api/users", apiCfg.HandlerDeleteUsers) NYI do I even want this
	//handler_auth.go
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLoginUser)
	mux.HandleFunc("GET /api/refresh", apiCfg.HandlerGetRefreshTokens)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefreshUser)
	mux.HandleFunc("POST /api/logout", apiCfg.HandlerLogoutUser)
	//handler_visitors.go
	mux.HandleFunc("POST /api/visitors", apiCfg.HandlerPostVisitors)
	mux.HandleFunc("PUT /api/visitors/{visitor_id}", apiCfg.HandlerPutVisitorsByID)
	mux.HandleFunc("GET /api/visitors", apiCfg.HandlerGetVisitors)
	mux.HandleFunc("GET /api/visitors/{visitor_id}", apiCfg.HandlerGetVisitorsByID)
	//handler_purposes.go
	mux.HandleFunc("POST /api/purposes", apiCfg.HandlerPostPurposes)
	mux.HandleFunc("PUT /api/purposes/{purpose_id}", apiCfg.HandlerPutPurposesByID)
	mux.HandleFunc("GET /api/purposes", apiCfg.HandlerGetPurposes)
	mux.HandleFunc("GET /api/purposes/{purpose_id}", apiCfg.HandlerGetPurposesByID) // NYI is this needed? Maybe GetPurposesByName instead?
	//handler_servicelogs.go NYI
	/// register handlers from the admin package
	//handler_admin.go
	mux.HandleFunc("POST /admin/users", func(w http.ResponseWriter, r *http.Request) {
		admin.AdminCreateUser(w, r, apiCfg, apiCfg.DB)
	})

	// fileserver
	fS := http.FileServer(http.Dir("./frontend/"))
	mux.Handle("/", fS)

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
