package entities

import (
	"time"

	"github.com/google/uuid"
)

// Crawler data
type Crawler struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
