package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Barrioslopezfd/chirpy/internal/auth"
	"github.com/Barrioslopezfd/chirpy/internal/database"
)

type parameter struct {
    Password    string  `json:"password"`
    Email   string  `json:"email"`
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
    token, err := auth.MakeJWT(usr.ID, cfg.jwtSecret)
    if err != nil {
        responseWithError(w, 500, fmt.Sprintf("JWT MAKING ERROR: %v", err))
        return
    }
    refTok, err := auth.MakeRefreshToken()
    if err != nil {
        responseWithError(w, 500, fmt.Sprint(err))
        return
    }

    _ , err=cfg.db.CreateRefreshToken(r.Context(), database. CreateRefreshTokenParams{
        Token: refTok,
        UserID: usr.ID,
        ExpiresAt: time.Now().Add(1440*time.Hour),
    })

    if err != nil {
        responseWithError(w, 500, fmt.Sprintf("ERROR CREATING REFRESH TOKEN: %v", err))
        return
    }

    usrJson, err := json.Marshal(User{
        ID: usr.ID,
        CreatedAt: usr.CreatedAt,
        UpdatedAt: usr.UpdatedAt,
        Email: usr.Email,
        Token: token,
        RefreshToken: refTok,
    })
    if err != nil {
        responseWithError(w, 500, "Error marshaling json")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Authorization", token)
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

func (cfg *apiConfig) HandleUserInfo(w http.ResponseWriter, r *http.Request){

    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        responseWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    type parameter struct{
        Password    string
        Email       string
    }
    decoder := json.NewDecoder(r.Body)
    var param parameter
    err = decoder.Decode(&param)
    if err != nil {
        responseWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    hash,err:=auth.HashPassword(param.Password)
    if err != nil {
        responseWithError(w, http.StatusInternalServerError, fmt.Sprint("Error hashing Password: ", err))
        return
    }
    usrID, err:=auth.ValidateJWT(token, cfg.jwtSecret)
    if err != nil {
        responseWithError(w, http.StatusUnauthorized, fmt.Sprint("Error validating JWT: ",err))
        return
    }
    change := database.ChangeEmailAndPasswordParams{
        ID: usrID,
        Email: param.Email,
        HashedPassword: hash,
    }
    err=cfg.db.ChangeEmailAndPassword(r.Context(), change)
    if err != nil {
        responseWithError(w, http.StatusInternalServerError, fmt.Sprint("Error changing email and password: ", err))
        return
    }

    type data struct {
        Token   string  `json:"token"`
        Email   string  `json:"email"`
    }

    dat, err := json.Marshal(data{
        Token:  token,
        Email:  param.Email,
    })

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(dat)
}
