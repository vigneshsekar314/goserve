package auth

import (
	"log"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed_paswd, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Printf("password could not be hashed, %s\n", err)
		return "", err
	}
	return hashed_paswd, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	is_hash_match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Printf("password compare failed: %s\n", err)
		return false, err
	}
	return is_hash_match, nil
}
