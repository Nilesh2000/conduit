package repository

import "time"

type Comment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	Author    Profile   `json:"author"`
	Article   Article   `json:"article"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
