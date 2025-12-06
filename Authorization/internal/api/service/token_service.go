package service

import (
	"github.com/Wladim1r/auth/internal/api/repository"
	"github.com/Wladim1r/auth/internal/models"
	"github.com/google/uuid"
)

type TokenService interface {
	StoreRefreshToken(session *models.Session) error
	GetSessionByToken(token string) (*models.Session, error)
	DeleteSessionByToken(token string) error
	DeleteAllUserSessions(userID uuid.UUID) error
}

type tokenService struct {
	tokenRepo repository.TokenRepository
}

func NewTokenService(tokenRepo repository.TokenRepository) TokenService {
	return &tokenService{
		tokenRepo: tokenRepo,
	}
}

func (s *tokenService) StoreRefreshToken(session *models.Session) error {
	return s.tokenRepo.StoreRefreshToken(session)
}

func (s *tokenService) GetSessionByToken(token string) (*models.Session, error) {
	return s.tokenRepo.GetByRefreshTokenHash(token)
}

func (s *tokenService) DeleteSessionByToken(token string) error {
	return s.tokenRepo.DeleteByRefreshTokenHash(token)
}

func (s *tokenService) DeleteAllUserSessions(userID uuid.UUID) error {
	return s.tokenRepo.DeleteAllUserSessions(userID)
}
