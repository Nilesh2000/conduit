package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type MockUserService struct {
	RegisterFunc func(username, email, password string) (*User, error)
}

func (m *MockUserService) Register(username, email, password string) (*User, error) {
	return m.RegisterFunc(username, email, password)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		mockRegister     func(username, email, password string) (*User, error)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid Registration",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password"
				}
			}`,
			mockRegister: func(username, email, password string) (*User, error) {
				if username != "testuser" || email != "test@example.com" || password != "password" {
					t.Errorf("Expected Register(%q, %q, %q), got Register(%q, %q, %q)", "testuser", "test@example.com", "password", username, email, password)
				}
				return &User{
					Email:    email,
					Token:    "jwt.token.here",
					Username: username,
					Bio:      "",
					Image:    "",
				}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: UserResponse{
				User: User{
					Email:    "test@example.com",
					Token:    "jwt.token.here",
					Username: "testuser",
					Bio:      "",
					Image:    "",
				},
			},
		},
		{
			name: "Invalid JSON",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
			}`,
			mockRegister: func(username, email, password string) (*User, error) {
				t.Errorf("Register should not be called for invalid JSON")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
				}{Body: []string{"Invalid request body"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockUserService{
				RegisterFunc: tt.mockRegister,
			}

			req := httptest.NewRequest("POST", "/api/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := RegisterHandler(mockUserService)
			handler.ServeHTTP(rr, req)

			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			var got interface{}
			if tt.expectedStatus == http.StatusCreated {
				var resp UserResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			} else {
				var resp ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}
