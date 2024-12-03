package main 

import (
    "net/http"
    "fmt"
)

func handleResponse(w http.ResponseWriter, _ *http.Request) {
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleNumberOfServerHits(w http.ResponseWriter, _ *http.Request) {
    w.Header().Add("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte(fmt.Sprintf(
        `<html>
            <body>
                <h1>Welcome, Chirpy Admin</h1>
                <p>Chirpy has been visited %v times!</p>
            </body>
        </html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleResetServerHits(w http.ResponseWriter, _ *http.Request) {
    cfg.fileserverHits.Store(0)
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
        cfg.fileserverHits.Add(1)
    })
}

