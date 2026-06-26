package models

import "time"

// User represents a registered account in the system.
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// One user can own multiple tickets.
	Tickets []Ticket `json:"-" gorm:"foreignKey:UserID"`
}