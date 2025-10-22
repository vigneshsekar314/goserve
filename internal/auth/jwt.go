package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userid uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	utcNow := time.Now().UTC()
	expiryTime := utcNow.Add(expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(utcNow),
		ExpiresAt: jwt.NewNumericDate(expiryTime),
		Subject:   userid.String(),
	})
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("error in signing token: %s\n", err)
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimData := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimData, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("error in parsing jwt: %s\n", err)
		return uuid.UUID{}, err
	}
	user_id, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("error in getting subject%s\n", err)
		return uuid.UUID{}, err
	}
	user_uuid, err := uuid.Parse(user_id)
	if err != nil {
		log.Printf("error in parsing uuid %s\n", err)
		return uuid.UUID{}, err
	}
	return user_uuid, nil
}
