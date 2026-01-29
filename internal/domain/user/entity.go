package user

import "time"

type User struct {
	ID        uint64
	Name      string
	Email     string
	CreatedAt time.Time
}
