package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Barrioslopezfd/chirpy/internal/auth"
)

func (cfg *apiConfig) RefreshAccessToken(w http.ResponseWriter, r *http.Request){
    type tokenResponse struct {
	Token	string	`json:"token"`
    }
    refToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
	responseWithError(w, 500, fmt.Sprintf("ERROR GETTING TOKEN: %v", err))
	return
    }
    usrToken, err := cfg.db.GetRefreshToken(r.Context(), refToken)
    if err != nil || usrToken.ExpiresAt.Before(time.Now()) || usrToken.RevokedAt.Valid {
	responseWithError(w, 401, "Unauthorized")
	return
    }
    token, err := auth.MakeJWT(usrToken.UserID, cfg.jwtSecret)
    if err != nil {
	responseWithError(w, 500, fmt.Sprint("ERROR MAKING JWT: ", err))
	return
    }
    data, err := json.Marshal(tokenResponse{
	Token: fmt.Sprint(token),
    })
    if err != nil {
	responseWithError(w, 500, fmt.Sprintf("ERROR MARSHALLING: %v", err))
	return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    w.Write(data)
}
func (cfg *apiConfig) RevokeRefreshToken(w http.ResponseWriter, r *http.Request){
    refToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
	responseWithError(w, 500, fmt.Sprintf("ERROR GETTING TOKEN: %v", err))
	return
    }
    cfg.db.RevokeRefreshToken(r.Context(), refToken)
    w.WriteHeader(204)
}
