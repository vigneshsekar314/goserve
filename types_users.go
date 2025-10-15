package main

import (
	"time"

	"github.com/google/uuid"
)

type createUserReq struct {
	Email string `json:"email"`
}

type createUserRes struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
