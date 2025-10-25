package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	auth_token := headers.Get("Authorization")
	auth_strs := strings.Fields(auth_token)
	if len(auth_strs) < 2 {
		return "", errors.New("Invalid / No Authorization header found\n")
	}
	return strings.Trim(auth_strs[1], " "), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	api_key := headers.Get("Authorization")
	api_strs := strings.Fields(api_key)
	if len(api_strs) < 2 {
		return "", errors.New("Invalid / No API key found in header\n")
	}
	return strings.Trim(api_strs[1], " "), nil
}
