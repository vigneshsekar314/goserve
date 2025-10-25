package main

import (
	"time"

	"github.com/google/uuid"
)

type createAndLoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createLoginUserRes struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

const EMAILPASSERR string = "Incorrect email or password"
