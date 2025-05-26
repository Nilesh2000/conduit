package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Nilesh2000/conduit/internal/repository"
)

// MockProfileRepository is a mock implementation of the ProfileRepository interface
type MockProfileRepository struct {
	getByUsernameFunc func(ctx context.Context, username string, currentUserID int64) (*repository.Profile, error)
	followUserFunc    func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error)
	unfollowUserFunc  func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error)
	isFollowingFunc   func(ctx context.Context, followerID int64, followingID int64) (bool, error)
}

var _ ProfileRepository = (*MockProfileRepository)(nil)

// GetByUsername returns a profile by username
func (m *MockProfileRepository) GetByUsername(
	ctx context.Context,
	username string,
	currentUserID int64,
) (*repository.Profile, error) {
	return m.getByUsernameFunc(ctx, username, currentUserID)
}

// FollowUser follows a user
func (m *MockProfileRepository) FollowUser(
	ctx context.Context,
	followerID int64,
	followingName string,
) (*repository.Profile, error) {
	return m.followUserFunc(ctx, followerID, followingName)
}

// UnfollowUser unfollows a user
func (m *MockProfileRepository) UnfollowUser(
	ctx context.Context,
	followerID int64,
	followingName string,
) (*repository.Profile, error) {
	return m.unfollowUserFunc(ctx, followerID, followingName)
}

// IsFollowing checks if a user is following another user
func (m *MockProfileRepository) IsFollowing(
	ctx context.Context,
	followerID int64,
	followingID int64,
) (bool, error) {
	return m.isFollowingFunc(ctx, followerID, followingID)
}

// Test_profileService_GetByUsername tests the GetByUsername method of the profileService
func Test_profileService_GetProfile(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		currentUserID int64
		setupMock     func() *MockProfileRepository
		expectedError error
		validate      func(*testing.T, *Profile)
	}{
		{
			name:          "Profile found (not following)",
			username:      "testuser",
			currentUserID: 2,
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					getByUsernameFunc: func(ctx context.Context, username string, currentUserID int64) (*repository.Profile, error) {
						if username != "testuser" {
							t.Errorf("Expected username to be testuser, got %q", username)
						}
						if currentUserID != 2 {
							t.Errorf("Expected currentUserID to be 1, got %d", currentUserID)
						}

						return &repository.Profile{
							ID:        1,
							Username:  "testuser",
							Bio:       "Test bio",
							Image:     "https://example.com/testuser.jpg",
							Following: false,
						}, nil
					},
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, profile *Profile) {
				if profile.Username != "testuser" {
					t.Errorf("Expected username to be testuser, got %q", profile.Username)
				}
				if profile.Bio != "Test bio" {
					t.Errorf("Expected bio to be Test bio, got %q", profile.Bio)
				}
				if profile.Image != "https://example.com/testuser.jpg" {
					t.Errorf(
						"Expected image to be https://example.com/testuser.jpg, got %q",
						profile.Image,
					)
				}
				if profile.Following {
					t.Errorf("Expected following to be false, got true")
				}
			},
		},
		{
			name:          "Profile found (following)",
			username:      "followeduser",
			currentUserID: 1,
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					getByUsernameFunc: func(ctx context.Context, username string, currentUserID int64) (*repository.Profile, error) {
						if username != "followeduser" {
							t.Errorf("Expected username to be followeduser, got %q", username)
						}
						if currentUserID != 1 {
							t.Errorf("Expected currentUserID to be 1, got %d", currentUserID)
						}

						return &repository.Profile{
							ID:        2,
							Username:  "followeduser",
							Bio:       "Another bio",
							Image:     "https://example.com/another.jpg",
							Following: true,
						}, nil
					},
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, profile *Profile) {
				if profile.Username != "followeduser" {
					t.Errorf("Expected username to be followeduser, got %q", profile.Username)
				}
				if profile.Bio != "Another bio" {
					t.Errorf("Expected bio to be Another bio, got %q", profile.Bio)
				}
				if profile.Image != "https://example.com/another.jpg" {
					t.Errorf(
						"Expected image to be https://example.com/another.jpg, got %q",
						profile.Image,
					)
				}
				if !profile.Following {
					t.Errorf("Expected following to be true, got false")
				}
			},
		},
		{
			name:          "User not found",
			username:      "nonexistentuser",
			currentUserID: 1,
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					getByUsernameFunc: func(ctx context.Context, username string, currentUserID int64) (*repository.Profile, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validate:      nil,
		},
		{
			name:          "Repository error",
			username:      "testuser",
			currentUserID: 1,
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					getByUsernameFunc: func(ctx context.Context, username string, currentUserID int64) (*repository.Profile, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validate:      nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockRepo := tt.setupMock()

			// Create service
			service := NewProfileService(mockRepo)

			// Call GetProfile
			profile, err := service.GetProfile(
				context.Background(),
				tt.username,
				tt.currentUserID,
			)

			// Validate results
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Validate profile if expected
			if err == nil && tt.validate != nil {
				tt.validate(t, profile)
			}
		})
	}
}

// Test_profileService_FollowUser tests the FollowUser method of the profileService
func Test_profileService_FollowUser(t *testing.T) {
	tests := []struct {
		name          string
		followerID    int64
		followingName string
		setupMock     func() *MockProfileRepository
		expectedError error
		validate      func(*testing.T, *Profile)
	}{
		{
			name:          "Successfully follow user",
			followerID:    1,
			followingName: "usertofollow",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						if followerID != 1 {
							t.Errorf("Expected followerID to be 1, got %d", followerID)
						}
						if followingName != "usertofollow" {
							t.Errorf(
								"Expected followingName to be usertofollow, got %q",
								followingName,
							)
						}

						return &repository.Profile{
							ID:        2,
							Username:  "usertofollow",
							Bio:       "Their bio",
							Image:     "https://example.com/their-image.jpg",
							Following: true,
						}, nil
					},
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, profile *Profile) {
				if profile.Username != "usertofollow" {
					t.Errorf("Expected username to be usertofollow, got %q", profile.Username)
				}
				if profile.Bio != "Their bio" {
					t.Errorf("Expected bio to be Their bio, got %q", profile.Bio)
				}
				if profile.Image != "https://example.com/their-image.jpg" {
					t.Errorf(
						"Expected image to be https://example.com/their-image.jpg, got %q",
						profile.Image,
					)
				}
				if !profile.Following {
					t.Errorf("Expected following to be true, got false")
				}
			},
		},
		{
			name:          "User not found",
			followerID:    1,
			followingName: "nonexistentuser",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validate:      nil,
		},
		{
			name:          "Cannot follow yourself",
			followerID:    1,
			followingName: "currentuser",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrCannotFollowSelf
					},
				}
			},
			expectedError: ErrCannotFollowSelf,
			validate:      nil,
		},
		{
			name:          "Repository error",
			followerID:    1,
			followingName: "usertofollow",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validate:      nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockRepo := tt.setupMock()

			// Create service
			service := NewProfileService(mockRepo)

			// Call FollowUser
			profile, err := service.FollowUser(
				context.Background(),
				tt.followerID,
				tt.followingName,
			)

			// Validate results
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Validate profile if expected
			if tt.validate != nil {
				tt.validate(t, profile)
			}
		})
	}
}

// Test_profileService_UnfollowUser tests the UnfollowUser method of the profileService
func Test_profileService_UnfollowUser(t *testing.T) {
	tests := []struct {
		name          string
		followerID    int64
		followingName string
		setupMock     func() *MockProfileRepository
		expectedError error
		validate      func(*testing.T, *Profile)
	}{
		{
			name:          "Successfully unfollow user",
			followerID:    1,
			followingName: "usertounfollow",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						if followerID != 1 {
							t.Errorf("Expected followerID to be 1, got %d", followerID)
						}
						if followingName != "usertounfollow" {
							t.Errorf(
								"Expected followingName to be usertounfollow, got %q",
								followingName,
							)
						}

						return &repository.Profile{
							ID:        2,
							Username:  "usertounfollow",
							Bio:       "Their image",
							Image:     "https://example.com/their-image.jpg",
							Following: false,
						}, nil
					},
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, profile *Profile) {
				if profile.Username != "usertounfollow" {
					t.Errorf("Expected username to be usertounfollow, got %q", profile.Username)
				}
				if profile.Bio != "Their image" {
					t.Errorf("Expected bio to be Their image, got %q", profile.Bio)
				}
				if profile.Image != "https://example.com/their-image.jpg" {
					t.Errorf(
						"Expected image to be https://example.com/their-image.jpg, got %q",
						profile.Image,
					)
				}
				if profile.Following {
					t.Errorf("Expected following to be false, got true")
				}
			},
		},
		{
			name:          "User not found",
			followerID:    1,
			followingName: "nonexistentuser",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validate:      nil,
		},
		{
			name:          "Cannot unfollow yourself",
			followerID:    1,
			followingName: "currentuser",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrCannotFollowSelf
					},
				}
			},
			expectedError: ErrCannotFollowSelf,
			validate:      nil,
		},
		{
			name:          "Repository error",
			followerID:    1,
			followingName: "usertounfollow",
			setupMock: func() *MockProfileRepository {
				return &MockProfileRepository{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*repository.Profile, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validate:      nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockRepo := tt.setupMock()

			// Create service
			service := NewProfileService(mockRepo)

			// Call UnfollowUser
			profile, err := service.UnfollowUser(
				context.Background(),
				tt.followerID,
				tt.followingName,
			)

			// Validate results
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Validate profile if expected
			if tt.validate != nil {
				tt.validate(t, profile)
			}
		})
	}
}
