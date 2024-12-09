package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/Barrioslopezfd/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
    fileserverHits  atomic.Int32
    db		    *database.Queries
    env		    string
    jwtSecret	    string
    polkaKey	    string
}

type User struct {
    ID		    uuid.UUID   `json:"id"`
    CreatedAt	    time.Time   `json:"created_at"`
    UpdatedAt	    time.Time   `json:"updated_at"`
    Email	    string	`json:"email"`
    Token	    string	`json:"token"`
    RefreshToken    string	`json:"refresh_token"`
    IsChirpyRed	    bool	`json:"is_chirpy_red"`
}

func main() {

    godotenv.Load()
    dbURL := os.Getenv("DB_URL")
    platform := os.Getenv("PLATFORM")
    jwtSecret := os.Getenv("JWT_SECRET")
    polkaKey := os.Getenv("POLKA_KEY")
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    dbQueries := database.New(db)
    apiCfg := &apiConfig{
	fileserverHits: atomic.Int32{},
	db:		dbQueries,
	env:		platform,
	jwtSecret:	jwtSecret,
	polkaKey:	polkaKey,
    }

    mux := http.NewServeMux()
    server := &http.Server{
        Handler: mux,
        Addr: ":8080",
    }
    mux.Handle(
        "/app/", 
        apiCfg.middlewareMetrics(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
    )

    mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)
    mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
    mux.HandleFunc("GET /api/chirps/{ChirpID}", apiCfg.GetChirpsSingle)
    mux.HandleFunc("DELETE /api/chirps/{ChirpID}", apiCfg.DeleteChirp)

    mux.HandleFunc("PUT /api/users", apiCfg.HandleUserInfo)
    mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
    mux.HandleFunc("POST /api/login", apiCfg.LoginUser)
    mux.HandleFunc("POST /api/refresh", apiCfg.RefreshAccessToken)
    mux.HandleFunc("POST /api/revoke", apiCfg.RevokeRefreshToken)

    mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeToChirpyRed)
    mux.HandleFunc("GET /api/healthz", handleResponse)
    mux.HandleFunc("GET /admin/metrics", apiCfg.handleNumberOfServerHits)
    mux.HandleFunc("POST /admin/reset", apiCfg.resetUsers)
    log.Fatal(server.ListenAndServe())
}


