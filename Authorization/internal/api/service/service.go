// Package service
package service

import (
	"github.com/Wladim1r/auth/internal/api/repository"
	"github.com/Wladim1r/auth/internal/models"
	"github.com/Wladim1r/auth/lib/errs"
	"github.com/Wladim1r/auth/lib/hashpwd"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegisterUser(name string, password string) error
	LoginUser(name string, password string) (*models.User, error)
	DeleteUserByID(userID uuid.UUID) error
	StoreRefreshToken(session *models.Session) error
	GetSessionByToken(token string) (*models.Session, error)
	DeleteSessionByToken(token string) error
	DeleteAllUserSessions(userID uuid.UUID) error
}

type service struct {
	userRepo    repository.UsersDB
	sessionRepo repository.SessionsDB
}

func NewService(userRepo repository.UsersDB, sessionRepo repository.SessionsDB) Service {
	return &service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *service) RegisterUser(name string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &models.User{
		Name:     name,
		Password: string(hashedPassword),
	}
	return s.userRepo.CreateUser(user)
}

func (s *service) LoginUser(name string, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByName(name)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.ErrRecordingWNF
	}
	return user, nil
}

func (s *service) DeleteUserByID(userID uuid.UUID) error {
	return s.userRepo.DeleteUserByID(userID)
}

func (s *service) StoreRefreshToken(session *models.Session) error {
	return s.sessionRepo.StoreRefreshToken(session)
}

func (s *service) GetSessionByToken(token string) (*models.Session, error) {
	tokenHash := hashpwd.HashToken(token)
	return s.sessionRepo.GetByRefreshTokenHash(tokenHash)
}

func (s *service) DeleteSessionByToken(token string) error {
	tokenHash := hashpwd.HashToken(token)
	return s.sessionRepo.DeleteByRefreshTokenHash(tokenHash)
}

func (s *service) DeleteAllUserSessions(userID uuid.UUID) error {
	return s.sessionRepo.DeleteAllUserSessions(userID)
}
