package kafka

import "time"

type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

const (
	EventUserRegistered = "user.registered"
	EventUserLoggedIn   = "user.logged_in"
	EventPasswordReset  = "user.password_reset"
	EventProductCreated = "product.created"
	EventProductUpdated = "product.updated"
	EventProductDeleted = "product.deleted"
	EventHTTPRequest    = "http.request"
)
