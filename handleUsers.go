package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

    type parameter struct {
        Email   string  `json:"email"`
    }

    decoder := json.NewDecoder(r.Body)
    var param parameter
    err := decoder.Decode(&param)
    if err != nil {
        responseWithError(w, 500, "Error decoding json")
        return
    }
    usr, err := cfg.db.CreateUser(r.Context(), param.Email)
    if err != nil {
        responseWithError(w, 500, "Error creating user")
        return
    }

    user_t := User{
        ID: usr.ID,
        CreatedAt: usr.CreatedAt,
        UpdatedAt: usr.UpdatedAt,
        Email: usr.Email,
    }

    usrJson, err := json.Marshal(user_t)
    if err != nil {
        responseWithError(w, 500, "Error marshallig json")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    w.Write(usrJson)

}

func (cfg *apiConfig) resetUsers(w http.ResponseWriter, r *http.Request) {
    if cfg.env != "dev" {
        w.WriteHeader(403)
        w.Write([]byte("403 Forbidden status"))
        return
    }
    err := cfg.db.Reset(r.Context())
    if err != nil {
        responseWithError(w, 500, "Failed reseting users")
        return
    }
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}
