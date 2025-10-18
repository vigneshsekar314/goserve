package main

import (
	"time"

	"github.com/google/uuid"
)

type ChirpRequest struct {
	Body   string    `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}

type ChirpResponse struct {
	Id        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
