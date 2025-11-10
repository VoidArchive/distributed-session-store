package session

import "time"

type Session struct {
	ID        string
	Data      map[string]any
	ExpiresAt time.Time
}

type Store interface {
	Get(id string) (*Session, error)
	Set(id string, data map[string]any, ttl time.Duration) error
	Delete(id string) error
}
