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
	"github.com/dcrauwels/goqueue/strutils"
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
	accessTokenDuration, err := strutils.GetIntegerEnvironmentVariable("ACCESSTOKENDURATION")
	if err != nil {
		log.Printf("Environment variable ACCESSTOKENDURATION not provided: %v", err)
		panic(err)
	}
	refreshTokenDuration, err := strutils.GetIntegerEnvironmentVariable("REFRESHTOKENDURATION")
	if err != nil {
		log.Printf("Environment variable REFRESHTOKENDURATION not provided: %v", err)
		panic(err)
	}
	publicIDLength, err := strutils.GetIntegerEnvironmentVariable("PUBLICIDLENGTH")
	if err != nil {
		log.Printf("Environment variable PUBLICIDLENGTH not provided: %v", err)
		panic(err)
	}

	apiCfg := api.ApiConfig{
		DB:                   dbQueries,
		Secret:               os.Getenv("SECRET"),
		Env:                  os.Getenv("ENV"),
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		PublicIDLength:       publicIDLength,
	}

	// servemux
	mux := http.NewServeMux()

	/// register handlers from api package
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", apiCfg.ReadinessHandler) // ok
	//handler_users.go
	mux.Handle("POST /api/users", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostUsers)))                    // ok
	mux.Handle("PUT /api/users", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutUsers)))                      // ok
	mux.Handle("PUT /api/users/{public_user_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutUsersByID))) // ok
	mux.Handle("GET /api/users", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetUsers)))                      // ok
	mux.HandleFunc("GET /api/users/{public_user_id}", apiCfg.HandlerGetUsersByID)                                          // ok
	//mux.HandleFunc("DELETE /api/users", apiCfg.HandlerDeleteUsers) NYI do I even want this
	//handler_auth.go
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLoginUser)
	mux.HandleFunc("GET /api/refresh", apiCfg.HandlerGetRefreshTokens)                                                      // ok (requires dev environment)
	mux.Handle("POST /api/refresh", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRefreshUser)))                 // not ok! I need to think about how this is going to work in relation to the auth middleware which already implements token rotation and access token generation from refresh tokens
	mux.Handle("POST /api/logout", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerLogoutUser)))                   // ok
	mux.Handle("POST /api/revoke", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRevokeAllRefreshTokens)))       // ok
	mux.Handle("POST /api/revoke/{user_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRevokeRefreshToken))) // ok
	//handler_visitors.go
	mux.HandleFunc("POST /api/visitors", apiCfg.HandlerPostVisitors)
	mux.Handle("PUT /api/visitors/{visitor_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutVisitorsByID))) // ok
	mux.Handle("GET /api/visitors", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetVisitors)))                  // ok
	mux.Handle("GET /api/visitors/{visitor_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetVisitorsByID))) // ok
	//handler_purposes.go
	mux.Handle("POST /api/purposes", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostPurposes)))                // ok
	mux.Handle("PUT /api/purposes/{purpose_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutPurposesByID))) // ok
	mux.HandleFunc("GET /api/purposes", apiCfg.HandlerGetPurposes)                                                           // ok no auth needed
	mux.HandleFunc("GET /api/purposes/{purpose_id}", apiCfg.HandlerGetPurposesByID)                                          // NYI is this needed? Maybe GetPurposesByName instead?
	//handler_servicelogs.go
	mux.Handle("POST /api/servicelogs", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostServicelogs)))                   // NYI
	mux.Handle("PUT /api/servicelogs/{servicelog_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutServicelogsByID))) // NYI
	mux.Handle("GET /api/servicelogs", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetServicelogs)))                     // NYI
	//mux.HandleFunc("GET /api/servicelogs/{visitor_id}", apiCfg.HandlerGetServicelogsByVisitorID)                                   // NYI not convinced this is needed

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
