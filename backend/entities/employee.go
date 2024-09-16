package entities

import (
	"database/sql"
	"time"
)

type Employee struct {
	Id        string         `json:"uuid"`
	Username  string         `json:"username"`
	FirstName sql.NullString `json:"firstName"`
	LastName  sql.NullString `json:"lastName"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
