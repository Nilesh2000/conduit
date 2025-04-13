package service

import (
	"conduit/internal/repository"
	"errors"
)

// ProfileRepository is an interface for the profile repository
type ProfileRepository interface {
	GetByUsername(username string) (repository.Profile, error)
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
	profile, err := s.profileRepository.GetByUsername(username)
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
