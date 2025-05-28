package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

// MockTagService is a mock implementation of the TagService interface
type MockTagService struct {
	getTagsFunc func(ctx context.Context) ([]string, error)
}

// GetTags is a mock implementation of the GetTags method
func (m *MockTagService) GetTags(ctx context.Context) ([]string, error) {
	return m.getTagsFunc(ctx)
}

func TestTagHandler_GetTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupMock        func() *MockTagService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successful tag retrieval",
			setupMock: func() *MockTagService {
				return &MockTagService{
					getTagsFunc: func(ctx context.Context) ([]string, error) {
						return []string{"tag1", "tag2"}, nil
					},
				}
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: TagResponse{Tags: []string{"tag1", "tag2"}},
		},
		{
			name: "Internal server error",
			setupMock: func() *MockTagService {
				return &MockTagService{
					getTagsFunc: func(ctx context.Context) ([]string, error) {
						return nil, service.ErrInternalServer
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewTagHandler(mockService)

			// Create Request
			req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.GetTags()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp TagResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}
