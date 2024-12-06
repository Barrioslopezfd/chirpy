package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hashString string, err error){
    hashPass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    if err != nil {
        return "", err
    }
    return string(hashPass), nil
}

func CheckPasswordHash(password, hash string) error {
    err:=bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        return err
    }
    return nil
}
