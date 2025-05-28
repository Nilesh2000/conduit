package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// MockTagRepository is a mock implementation of the TagRepository interface
type MockTagRepository struct {
	getFunc func(ctx context.Context) ([]string, error)
}

// GetTags is a mock implementation of the GetTags method
func (m *MockTagRepository) Get(ctx context.Context) ([]string, error) {
	return m.getFunc(ctx)
}

func Test_tagService_GetTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func() *MockTagRepository
		expectedErr error
		validate    func(*testing.T, []string)
	}{
		{
			name: "Successful retrieval",
			setupMock: func() *MockTagRepository {
				return &MockTagRepository{
					getFunc: func(ctx context.Context) ([]string, error) {
						return []string{"tag1", "tag2"}, nil
					},
				}
			},
			expectedErr: nil,
			validate: func(t *testing.T, tags []string) {
				if !reflect.DeepEqual(tags, []string{"tag1", "tag2"}) {
					t.Errorf("Expected tags %v, got %v", []string{"tag1", "tag2"}, tags)
				}
			},
		},
		{
			name: "Internal server error",
			setupMock: func() *MockTagRepository {
				return &MockTagRepository{
					getFunc: func(ctx context.Context) ([]string, error) {
						return nil, ErrInternalServer
					},
				}
			},
			expectedErr: ErrInternalServer,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockRepository := tt.setupMock()

			// Create service with mock repository
			tagService := NewTagService(mockRepository)

			// Create context
			ctx := context.Background()

			// Call method
			tags, err := tagService.GetTags(ctx)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate tags if expected
			if err == nil && tt.validate != nil {
				tt.validate(t, tags)
			}
		})
	}
}
