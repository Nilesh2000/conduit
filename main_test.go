package main

type MockUserService struct {
	RegisterFunc func(username, email, password string) (*User, error)
}

func (m *MockUserService) Register(username, email, password string) (*User, error) {
	return m.RegisterFunc(username, email, password)
}
