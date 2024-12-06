package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Barrioslopezfd/chirpy/internal/auth"
	"github.com/Barrioslopezfd/chirpy/internal/database"
)

type parameter struct {
    Password    string  `json:"password"`
    Email       string  `json:"email"`
}

func (cfg *apiConfig) LoginUser(w http.ResponseWriter, r *http.Request){
    decoder := json.NewDecoder(r.Body)
    var param parameter
    err := decoder.Decode(&param)
    if err != nil {
        responseWithError(w, 500, "Error decoding json")
        return
    }
    usr, err := cfg.db.GetUserByEmail(r.Context(), param.Email)
    if err != nil {
        responseWithError(w, 404, "Incorrect email")
        return
    }
    err = auth.CheckPasswordHash(param.Password, usr.HashedPassword)
    if err != nil {
        responseWithError(w, 401, "Incorrect email or password")
        return 
    }
    usrJson, err := json.Marshal(User{
        ID: usr.ID,
        CreatedAt: usr.CreatedAt,
        UpdatedAt: usr.UpdatedAt,
        Email: usr.Email,
    })
    if err != nil {
        responseWithError(w, 500, "Error marshaling json")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    w.Write([]byte(usrJson))
}

func (cfg *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {


    decoder := json.NewDecoder(r.Body)
    var param parameter
    err := decoder.Decode(&param)
    if err != nil {
        responseWithError(w, 500, "Error decoding json")
        return
    }
    hashPass, err := auth.HashPassword(param.Password)
    if err != nil {
        Serr := fmt.Sprint(err)
        responseWithError(w, 500, Serr)
        return
    }
    usr, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
        HashedPassword: hashPass,
        Email:  param.Email,
    })
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

