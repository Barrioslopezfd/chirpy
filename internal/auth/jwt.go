package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error){
    token := jwt.NewWithClaims(
	jwt.SigningMethodHS256,
	jwt.RegisteredClaims{
	    Issuer:	"chirpy",
	    IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
	    ExpiresAt: jwt.NewNumericDate(time.Now().Add(1*time.Hour).UTC()),
	    Subject: userID.String(),
	})
    signedToken, err := token.SignedString([]byte(tokenSecret))
    if err != nil {
	return "", err
    }
    return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token)(interface{}, error) {
	return []byte(tokenSecret), nil
    })
    if err != nil {
	return uuid.Nil, err
    }
    id, err := token.Claims.GetSubject()
    if err != nil {
	return uuid.Nil, err
    }
    sid, err := uuid.Parse(id)
    if err != nil {
	return uuid.Nil, err
    }

    return sid, nil
}

func GetBearerToken(headers http.Header) (string, error){
    tkn := headers.Get("Authorization")
    if tkn == "" {
	return "", errors.New("Unable to get authorization")
    }
    splitedTkn := strings.Split(tkn, " ")
    return splitedTkn[1], nil
}

func MakeRefreshToken() (string, error){
    data := make([]byte, 32)
    _, err := rand.Read(data)
    if err != nil {
	return "", err
    }
    return hex.EncodeToString(data), nil
}
