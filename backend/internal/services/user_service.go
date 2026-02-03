package services

import (
	"errors"
	"sentinel-id/internal/models"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/utils"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	UserRepo    *repositories.UserRepository
	SessionRepo *repositories.SessionRepository
}

func NewUserService(userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository) *UserService {
	return &UserService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}
}

func (s *UserService) Register(input models.UserRegisterDTO) (*models.User, error) {
	existingUser, _ := s.UserRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           uuid.New(),
		FullName:     input.FullName,
		Email:        input.Email,
		CPF:          input.CPF,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		MfaEnabled:   false,
	}

	err = s.UserRepo.CreateUser(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) Login(input models.UserLoginDTO, userAgent string, clientIP string) (string, string, error) {
	user, err := s.UserRepo.FindByEmail(input.Email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	match, err := utils.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !match {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	session := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		DeviceInfo:   userAgent,
		IPAddress:    clientIP,
		ExpiresAt:    time.Now().Add(24 * time.Hour * 7),
		CreatedAt:    time.Now(),
	}

	err = s.SessionRepo.Create(session)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(refreshToken string) (string, error) {
	session, err := s.SessionRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("database error")
	}
	if session == nil {
		return "", errors.New("invalid refresh token")
	}

	if session.IsRevoked {
		return "", errors.New("session is revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}

	newAccessToken, err := utils.GenerateToken(session.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *UserService) Logout(refreshToken string) error {
	return s.SessionRepo.Revoke(refreshToken)
}

func (s *UserService) RevokeAllSessions(userID uuid.UUID) error {
	return s.SessionRepo.RevokeAll(userID)
}

func (s *UserService) ListUserSessions(userID uuid.UUID) ([]models.Session, error) {
	return s.SessionRepo.ListByUser(userID)
}

func (s *UserService) LogoutAll(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.SessionRepo.RevokeAll(uid)
}

func (s *UserService) ListSessions(userID string) ([]models.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.SessionRepo.ListByUser(uid)
}
