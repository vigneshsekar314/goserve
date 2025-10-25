package main

import "github.com/google/uuid"

type UpgradeRequest struct {
	Event string          `json:"event"`
	Data  UserIdContainer `json:"data"`
}

type UserIdContainer struct {
	UserId uuid.UUID `json:"user_id"`
}
