// Package repository
package repository

import (
	"errors"
	"fmt"

	"github.com/Wladim1r/auth/internal/models"
	"github.com/Wladim1r/auth/lib/errs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionsDB interface {
	StoreRefreshToken(session *models.Session) error
	GetByRefreshTokenHash(tokenHash string) (*models.Session, error)
	DeleteByRefreshTokenHash(tokenHash string) error
	DeleteAllUserSessions(userID uuid.UUID) error
}

type sessionsDB struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionsDB {
	return &sessionsDB{
		db: db,
	}
}

func (db *sessionsDB) StoreRefreshToken(session *models.Session) error {

	result := db.db.Create(session)

	if result.Error != nil {
		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
	}

	return nil
}

func (db *sessionsDB) GetByRefreshTokenHash(tokenHash string) (*models.Session, error) {
	var session models.Session

	if err := db.db.Where("refresh_token = ?", tokenHash).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrRecordingWNF
		}
		return nil, fmt.Errorf("%w: %s", errs.ErrDB, err.Error())
	}

	return &session, nil
}

func (db *sessionsDB) DeleteByRefreshTokenHash(tokenHash string) error {
	result := db.db.Where("refresh_token = ?", tokenHash).Delete(&models.Session{})

	if err := result.Error; err != nil {
		return fmt.Errorf("%w: %s", errs.ErrDB, err.Error())
	}

	if result.RowsAffected == 0 {
		return errs.ErrRecordingWND
	}

	return nil
}

func (db *sessionsDB) DeleteAllUserSessions(userID uuid.UUID) error {
	result := db.db.Where("user_id = ?", userID).Delete(&models.Session{})
	if result.Error != nil {
		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
	}
	return nil
}
