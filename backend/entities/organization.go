package entities

import (
	"database/sql"
	"time"
)

type Organization struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Type        string         `json:"type"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}
