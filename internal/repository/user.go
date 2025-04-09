package repository

// User represents a user in the system
type User struct {
	ID       int64
	Username string
	Email    string
	Password string
	Bio      string
	Image    string
}
