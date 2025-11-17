package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ServiceName string    `db:"service_name" json:"service_name"`
	Price       int       `db:"price" json:"price"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	StartMonth  int       `db:"start_month" json:"start_month"`
	StartYear   int       `db:"start_year" json:"start_year"`
	EndMonth    *int      `db:"end_month,omitempty" json:"end_month,omitempty"`
	EndYear     *int      `db:"end_year,omitempty" json:"end_year,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
