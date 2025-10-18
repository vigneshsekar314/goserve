package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ErrorJson struct {
	Error string `json:"error"`
}

type ValidJson struct {
	CleanedBody string `json:"cleaned_body"`
}

func validate_chirp(req ChirpRequest) (ValidJson, error) {

	if len(req.Body) > 140 {
		return ValidJson{}, errors.New("Chirp is too long")
	}
	cleaned_msg := cleanMsg(req.Body)
	valid_json := ValidJson{
		CleanedBody: cleaned_msg,
	}
	return valid_json, nil
}

func cleanMsg(subject string) string {
	cleaned := make([]string, 0)
	banned_words := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range strings.Fields(subject) {
		var cleaned_word string = word
		for _, banned := range banned_words {
			if strings.EqualFold(word, banned) {
				cleaned_word = "****"
				break
			}
		}
		cleaned = append(cleaned, cleaned_word)
	}
	return strings.Join(cleaned, " ")
}

func GetMarshaledValue[T any](value T) ([]byte, error) {
	msg, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error marshalling value: %w", err)
	}
	return msg, nil
}
