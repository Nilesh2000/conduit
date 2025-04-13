package repository

// Profile represents a user profile in the repository
type Profile struct {
	ID        int64
	Username  string
	Bio       string
	Image     string
	Following bool
}
