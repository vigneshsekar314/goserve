package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type chirp_req struct {
	Body string `json:"body"`
}

type errorJson struct {
	Error string `json:"error"`
}

type validJson struct {
	CleanedBody string `json:"cleaned_body"`
}

func validate_chirp(w http.ResponseWriter, r *http.Request) {
	var api_req chirp_req
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&api_req); err != nil {
		log.Printf("error decoding params: %s", err)
		w.WriteHeader(400)
		errJs := errorJson{
			Error: "Something went wrong",
		}
		errmsg, err := json.Marshal(errJs)
		if err != nil {
			log.Printf("error decoding error message JSON: %s", err)
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(400)
		w.Write(errmsg)
		return
	}

	if len(api_req.Body) > 140 {
		w.WriteHeader(400)
		longmsg := errorJson{
			Error: "Chirp is too long",
		}
		errMsg, err := GetMarshaledValue(longmsg)
		if err != nil {
			log.Printf("%s", err)
			return
		}
		w.Write(errMsg)
		return
	}
	cleaned_msg := cleanMsg(api_req.Body)
	valid_json := validJson{
		CleanedBody: cleaned_msg,
	}
	response, err := GetMarshaledValue(valid_json)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(response)
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
