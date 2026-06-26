package models

import "time"

// TicketStatus represents the allowed lifecycle states of a ticket.
type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
)

// IsValidStatus reports whether s is one of the supported ticket statuses.
func IsValidStatus(s string) bool {
	switch TicketStatus(s) {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	}
	return false
}

// allowedTransitions defines the permitted forward-only status flow:
// open -> in_progress -> closed. A closed ticket can never be reopened
// or moved back to any earlier state.
var allowedTransitions = map[TicketStatus][]TicketStatus{
	StatusOpen:       {StatusInProgress},
	StatusInProgress: {StatusClosed},
	StatusClosed:     {},
}

// CanTransition reports whether moving from `from` to `to` is a valid
// status transition under the required status flow rules.
func CanTransition(from, to TicketStatus) bool {
	if from == to {
		return false
	}

	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	for _, n := range next {
		if n == to {
			return true
		}
	}

	return false
}

// Ticket represents a support ticket owned by exactly one user.
type Ticket struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Title       string       `json:"title" gorm:"not null"`
	Description string       `json:"description"`
	Status      TicketStatus `json:"status" gorm:"type:text;default:open;not null"`
	UserID      uint         `json:"user_id" gorm:"not null;index"`

	// Owner of the ticket.
	User User `json:"-" gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}