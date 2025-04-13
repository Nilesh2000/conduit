package service

import (
	"errors"

	"conduit/internal/repository"
)

// ProfileRepository is an interface for the profile repository
type ProfileRepository interface {
	GetByUsername(username string, currentUserID int64) (*repository.Profile, error)
	FollowUser(followerID int64, followingName string) (*repository.Profile, error)
}

// profileService implements the profileService interface
type profileService struct {
	profileRepository ProfileRepository
}

// NewProfileService creates a new profile service
func NewProfileService(profileRepository ProfileRepository) *profileService {
	return &profileService{
		profileRepository: profileRepository,
	}
}

// GetProfile gets a profile by username
func (s *profileService) GetProfile(username string, currentUserID int64) (*Profile, error) {
	profile, err := s.profileRepository.GetByUsername(username, currentUserID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	return &Profile{
		Username:  profile.Username,
		Bio:       profile.Bio,
		Image:     profile.Image,
		Following: profile.Following,
	}, nil
}

// FollowUser follows a user
func (s *profileService) FollowUser(followerID int64, followingName string) (*Profile, error) {
	profile, err := s.profileRepository.FollowUser(followerID, followingName)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrCannotFollowSelf):
			return nil, ErrCannotFollowSelf
		default:
			return nil, ErrInternalServer
		}
	}

	return &Profile{
		Username:  profile.Username,
		Bio:       profile.Bio,
		Image:     profile.Image,
		Following: profile.Following,
	}, nil
}
