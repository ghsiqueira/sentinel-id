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
	Repo        *repositories.UserRepository
	SessionRepo *repositories.SessionRepository
}

func NewUserService(repo *repositories.UserRepository, sessionRepo *repositories.SessionRepository) *UserService {
	return &UserService{
		Repo:        repo,
		SessionRepo: sessionRepo,
	}
}

func (s *UserService) Register(input models.UserRegisterDTO) (*models.User, error) {
	existingUser, err := s.Repo.GetByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	existingUser, err = s.Repo.GetByCPF(input.CPF)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("cpf already registered")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           uuid.New(),
		FullName:     input.FullName,
		Email:        input.Email,
		CPF:          input.CPF,
		PasswordHash: hashedPassword,
		MfaEnabled:   false,
		CreatedAt:    time.Now(),
	}

	err = s.Repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) Login(input models.UserLoginDTO, userAgent string, ip string) (string, string, error) {
	user, err := s.Repo.GetByEmail(input.Email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials")
	}

	match, err := utils.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}
	if !match {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		DeviceName:   userAgent,
		IPAddress:    ip,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
	}

	err = s.SessionRepo.CreateSession(session)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := utils.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	if claims["type"] != "refresh" {
		return "", errors.New("invalid token type")
	}

	session, err := s.SessionRepo.GetSessionByToken(refreshTokenString)
	if err != nil {
		return "", errors.New("session not found")
	}

	if session.IsRevoked {
		return "", errors.New("session revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}

	newAccessToken, err := utils.GenerateAccessToken(session.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *UserService) Logout(refreshTokenString string) error {
	return s.SessionRepo.RevokeSession(refreshTokenString)
}

func (s *UserService) LogoutAll(userID string) error {
	return s.SessionRepo.RevokeAllUserSessions(userID)
}

func (s *UserService) ListSessions(userID string) ([]models.Session, error) {
	return s.SessionRepo.GetSessionsByUserID(userID)
}
