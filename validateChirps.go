package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
    tooWrong := "Something went wrong"
    tooLong := "Chirp is too long"

    type parameters struct {
        Body string `json:"body"`
    }

    type cleanedParam struct {
        Cleaned_body string `json:"cleaned_body"`
    }

    decoder := json.NewDecoder(r.Body)
    param := parameters{}
    err := decoder.Decode(&param)
    if err != nil {
        responseWithError(w, 400, tooWrong, err)
        return
    }
    if len(param.Body) > 140 {
        responseWithError(w, 400, tooLong, errors.New("Too many arguments"))
        return
    }

    cleanParam := cleanedParam{}
    cleanParam.Cleaned_body, err = censorship(param.Body)
    if err != nil {
        log.Println(err)
        return
    }

    dat, err := json.Marshal(cleanParam)
    if err != nil {
        responseWithError(w, 400, tooWrong, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
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

func responseWithError(w http.ResponseWriter, code int, msg string, err error) {
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
    log.Printf("%s: %v", msg, err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(datWrong)
}
