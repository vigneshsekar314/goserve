package auth

const SECRET string = "goserveSecret&#()"

type RefreshTokenResponse struct {
	Token string `json:"token"`
}
