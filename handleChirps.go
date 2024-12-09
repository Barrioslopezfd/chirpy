package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Barrioslopezfd/chirpy/internal/auth"
	"github.com/Barrioslopezfd/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
    ID          uuid.UUID   `json:"id"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    Body        string      `json:"body"`
    UserID      uuid.UUID   `json:"user_id"`
}

func (cfg *apiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {

    type valid_ch struct {
        Body        string      `json:"body"`
        UserID       string   `json:"user_id"`
    }

    var ch valid_ch
    decoder := json.NewDecoder(r.Body)
    err:=decoder.Decode(&ch)
    if err != nil {
        responseWithError(w, 500, "Failed decoding json")
        return
    }
    valid_body, err := cfg.validateChirp(ch.Body)
    if err != nil {
        stringifiedError := fmt.Sprint(err)
        responseWithError(w, 400, stringifiedError)
        return
    }

    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        responseWithError(w, 401, fmt.Sprint(err))
        return
    }
    UID, err := auth.ValidateJWT(token, cfg.jwtSecret)
    if err != nil {
        responseWithError(w, 401, "Unauthorized")
        return
    }

    chirp_param:= database.CreateChirpParams{
        Body:   valid_body,
        UserID: UID,
    }

    dbchirp, err := cfg.db.CreateChirp(r.Context(), chirp_param)
    if err != nil {
        responseWithError(w, 500, "Error creating Chirp")
        return
    }

    Chirp := Chirp{
        ID: dbchirp.ID,
        CreatedAt: dbchirp.CreatedAt,
        UpdatedAt: dbchirp.UpdatedAt,
        Body: dbchirp.Body,
        UserID: dbchirp.UserID,
    }

    chJson, err := json.Marshal(Chirp)
    if err != nil {
        responseWithError(w, 500, "Error marshallig json")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    w.Write(chJson)
} 

func (cfg *apiConfig) validateChirp(Chirp string) (cleaned_body string, err error)  {

    if len(Chirp) > 140 {
        return "", fmt.Errorf("Chirp is too long")
    }
    cleaned_body, err = censorship(Chirp)
    if err != nil {
        return "", err
    }

    return cleaned_body, nil
}

func (cfg *apiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
    var myChirps []Chirp

    dbChirp,err:=cfg.db.GetChirps(r.Context())
    if err != nil {
        responseWithError(w, 400, "Error getting chirps")
        return
    }
    for i := range dbChirp {
        myChirps = append(myChirps, Chirp{
            ID: dbChirp[i].ID,
            CreatedAt: dbChirp[i].CreatedAt,
            UpdatedAt: dbChirp[i].UpdatedAt,
            Body: dbChirp[i].Body,
            UserID: dbChirp[i].UserID,
        })
    }
    ret, err := json.Marshal(myChirps)
    if err != nil {
        responseWithError(w, 400, "Error marshaling json")
        return 
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(ret)

}

func (cfg *apiConfig) GetChirpsSingle(w http.ResponseWriter, r *http.Request) {

    cid := r.PathValue("ChirpID")
    cuuid, err:=uuid.Parse(cid)
    if err != nil {
        responseWithError(w, 500, "Failed stringifying to uuid")
        return
    }

    dbChirp,err:=cfg.db.GetChirpsSingle(r.Context(), cuuid)
    if err != nil {
        responseWithError(w, 404, "Chirp not found")
        return
    }
    myChirp := Chirp{
        ID:         dbChirp.ID,
        CreatedAt:  dbChirp.CreatedAt,
        UpdatedAt:  dbChirp.UpdatedAt,
        Body:       dbChirp.Body,
        UserID:     dbChirp.UserID,
    }

    dat,err := json.Marshal(myChirp)
    if err != nil {
        responseWithError(w, 500, "Error marshaling json")
        return 
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(dat)

}

func censorship(s string) (string, error) {
    if s == "" {
        return s, fmt.Errorf("Failed to censor, received an empty string")
    }

    slice := strings.Split(s, " ")
    badWords := []string{"kerfuffle", "sharbert", "fornax"}
    for i := range slice {
        if strings.ToLower(slice[i]) == badWords[0] || strings.ToLower(slice[i]) == badWords[1] || strings.ToLower(slice[i]) == badWords[2]{
            slice[i] = "****"
        }
    }
    return strings.Join(slice, " "), nil
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
    type retError struct {
        Error string `json:"error"`
    }
    datErr := retError{
        Error:msg,
    }
    datWrong, err := json.Marshal(datErr)
    if err != nil {
        log.Printf("Error marshaling json: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        return
    }
    log.Printf("%s", msg)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(datWrong)
}

func (cfg *apiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request){
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        responseWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    uid, err := auth.ValidateJWT(token, cfg.jwtSecret)
    if err != nil {
        responseWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }
    cid := r.PathValue("ChirpID")
    cuid, err:=uuid.Parse(cid)
    if err != nil {
        responseWithError(w, http.StatusBadRequest, "Bad Request")
        return
    }
    chirp, err := cfg.db.GetChirpsSingle(r.Context(), cuid)
    if err != nil {
        responseWithError(w, http.StatusNotFound, "Not Found")
        return
    }
    if chirp.UserID != uid {
        responseWithError(w, http.StatusForbidden, "Forbidden Action")
    }
    err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
        ID: cuid,
        UserID: uid,
    })
    if err != nil {
        responseWithError(w, http.StatusForbidden, fmt.Sprint(err))
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
