package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dcrauwels/goqueue/admin"
	"github.com/dcrauwels/goqueue/api"
	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/jaevor/go-nanoid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// load .env into env variables
	godotenv.Load()
	// open db
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
	pidGenerator, err := nanoid.Standard(publicIDLength)
	if err != nil {
		log.Printf("public ID length not between 2 and 255: %v", err)
		panic(err)
	}
	resetTime, err := strutils.GetTimestampEnvironmentVariable("RESETTIME")
	if err != nil && !errors.Is(err, strutils.ErrNoValueFound) {
		log.Printf("Environment variable RESETTIME could not be parsed: %v", err)
		panic(err)
	}

	// init broker
	broker := api.NewBroker()
	go broker.Run()

	apiCfg := api.ApiConfig{
		DB:                   dbQueries,
		Broker:               broker,
		Secret:               os.Getenv("SECRET"),
		Env:                  os.Getenv("ENV"),
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		PublicIDGenerator:    pidGenerator,
		PublicIDLength:       publicIDLength,
		ResetTime:            resetTime,
	}

	// servemux
	mux := http.NewServeMux()

	/// register handlers from api package
	//handler_status.go
	mux.HandleFunc("GET /api/healthz", apiCfg.ReadinessHandler) // ok
	//handler_users.go
	mux.Handle("POST /api/users", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostUsers)))                          // ok
	mux.Handle("PUT /api/me", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutMe)))                                  // ok
	mux.Handle("PUT /api/me/desk", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutMeToDesk)))                       // ok
	mux.Handle("PUT /api/users/{user_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutUsersByPublicID))) // ok
	mux.Handle("GET /api/users", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetUsers)))                            // ok
	mux.Handle("GET /api/me", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetMe)))                                  // ok
	mux.HandleFunc("GET /api/users/{user_public_id}", apiCfg.HandlerGetUsersByID)                                                // ok
	//mux.HandleFunc("DELETE /api/users", apiCfg.HandlerDeleteUsers) NYI do I even want this
	//handler_auth.go
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLoginUser)                                                                     // ok
	mux.HandleFunc("GET /api/refresh", apiCfg.HandlerGetRefreshTokens)                                                             // ok (requires dev environment)
	mux.Handle("POST /api/refresh", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRefreshUser)))                        // not ok! I need to think about how this is going to work in relation to the auth middleware which already implements token rotation and access token generation from refresh tokens
	mux.Handle("POST /api/logout", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerLogoutUser)))                          // ok
	mux.Handle("POST /api/revoke", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRevokeAllRefreshTokens)))              // ok
	mux.Handle("POST /api/revoke/{user_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerRevokeRefreshToken))) // ok
	//handler_visitors.go
	mux.HandleFunc("POST /api/visitors", apiCfg.HandlerPostVisitors)                                                                       // ok
	mux.Handle("PUT /api/visitors/{visitor_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutVisitorsByPublicID)))  // ok
	mux.Handle("PUT /api/visitors/{visitor_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerCallVisitorsByPublicID))) // ok
	mux.Handle("GET /api/visitors", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetVisitors)))                                // ok
	mux.HandleFunc("GET /api/visitors/{visitor_public_id}", apiCfg.HandlerGetVisitorsByPublicID)                                           // ok
	mux.Handle("GET /api/visitors/events", apiCfg.AuthUserMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upid, _ := r.Context().Value(auth.UserPublicIDContextKey).(string)
		if upid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		apiCfg.Broker.ServeHTTP(w, r)
	}))) // ok > had to write a small wrapper to keep unauthorized access out of the broker
	mux.Handle("GET /api/visitors/queue", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetQueue)))             // ok
	mux.Handle("POST /api/visitors/call-next", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerCallNextVisitor))) // ok
	//handler_desks.go
	mux.Handle("POST /api/desks", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostDesks)))                          // ok
	mux.Handle("PUT /api/desks/{desk_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutDesksByPublicID))) // ok
	mux.Handle("GET /api/desks", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetDesks)))                            // ok
	mux.HandleFunc("GET /api/desks/{desk_public_id}", apiCfg.HandlerGetDesksByPublicID)                                          // ok
	//handler_purposes.go
	mux.Handle("POST /api/purposes", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostPurposes)))                       // ok
	mux.Handle("PUT /api/purposes/{purpose_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutPurposesByID))) // ok
	mux.HandleFunc("GET /api/purposes", apiCfg.HandlerGetPurposes)                                                                  // ok no auth needed
	mux.HandleFunc("GET /api/purposes/{purpose_public_id}", apiCfg.HandlerGetPurposesByID)                                          // NYI is this needed? Maybe GetPurposesByName instead?
	//handler_servicelogs.go
	mux.Handle("POST /api/servicelogs", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPostServicelogs)))                          // NYI
	mux.Handle("PUT /api/servicelogs/{servicelog_public_id}", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerPutServicelogsByID))) // NYI
	mux.Handle("GET /api/servicelogs", apiCfg.AuthUserMiddleware(http.HandlerFunc(apiCfg.HandlerGetServicelogs)))                            // NYI
	mux.HandleFunc("GET /api/servicelogs/{servicelog_public_id}", apiCfg.HandlerGetServicelogsByPublicID)                                    // NYI
	mux.Handle("GET /api/me/active-service", apiCfg.AuthUserMiddleware(http.HandlerFunc(api.HandlerGetMeActiveService)))                     // NYI

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
