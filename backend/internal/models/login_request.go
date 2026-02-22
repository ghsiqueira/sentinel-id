package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	ID        uuid.UUID  `json:"request_id"`
	Status    string     `json:"status"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}
