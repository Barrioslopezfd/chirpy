package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
    mux := http.NewServeMux()
    server := &http.Server{
        Handler: mux,
        Addr: ":8080",
    }
    apiCfg := &apiConfig{}
    mux.Handle(
        "/app/", 
        apiCfg.middlewareMetrics(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
    )
    mux.HandleFunc("GET /api/healthz", handleResponse)
    mux.HandleFunc("POST /api/validate_chirp", handleChirpValidation)
    mux.HandleFunc("GET /admin/metrics", apiCfg.handleNumberOfServerHits)
    mux.HandleFunc("POST /admin/reset", apiCfg.handleResetServerHits)
    err:=server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }
}


