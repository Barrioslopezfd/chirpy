package main

import (
	"fmt"
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
        apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
    )
    mux.HandleFunc("GET /api/healthz", handleResponse)
    mux.HandleFunc("GET /api/metrics", apiCfg.handleNumberOfServerHits)
    mux.HandleFunc("POST /api/reset", apiCfg.handleResetServerHits)
    err:=server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }
}

func handleResponse(w http.ResponseWriter, _ *http.Request) {

    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleNumberOfServerHits(w http.ResponseWriter, _ *http.Request) {
    fmt.Println("Hits: ", cfg.fileserverHits.Load())
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte(fmt.Sprint("Hits: ", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleResetServerHits(w http.ResponseWriter, _ *http.Request) {
    cfg.fileserverHits.Store(0)
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
        cfg.fileserverHits.Add(1)
    })
}
