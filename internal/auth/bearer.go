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
