package repository

// import (
// 	"fmt"

// 	"github.com/Wladim1r/auth/internal/models"
// 	"github.com/Wladim1r/auth/lib/errs"
// 	"gorm.io/gorm"
// )

// type UsersDB interface {
// 	CreateUser(user *models.User) error
// 	DeleteUserByID(userID uint) error
// 	SelectPwdByName(name string) (string, error)
// 	CheckUserExists(name string) error

// 	GetUserByName(name string) (*models.User, error)
// 	StoreRefreshToken(session *models.Session) error
// 	GetByRefreshTokenHash(token string) (*models.Session, error)
// 	DeleteByRefreshTokenHash(token string) error
// 	DeleteAllUserSessions(userID uint) error
// }

// type usersDB struct {
// 	db *gorm.DB
// }

// func NewRepository(db *gorm.DB) UsersDB {
// 	return &usersDB{
// 		db: db,
// 	}
// }


// func (db *usersDB) CreateUser(user *models.User) error {
// 	result := db.db.Create(&user)

// 	if result.RowsAffected == 0 {
// 		return errs.ErrRecordingWNC
// 	}

// 	if result.Error != nil {
// 		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return nil
// }

// func (db *usersDB) DeleteUserByID(userID uint) error {
// 	result := db.db.Where("id = ?", userID).Delete(&models.User{})
// 	if result.RowsAffected == 0 {
// 		return errs.ErrRecordingWND
// 	}

// 	if result.Error != nil {
// 		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return nil
// }

// func (db *usersDB) SelectPwdByName(name string) (string, error) {
// 	var user models.User

// 	result := db.db.Table("users").Select("password").Where("name = ?", name).Scan(&user)

// 	if result.RowsAffected == 0 {
// 		return "", errs.ErrRecordingWNF
// 	}

// 	if result.Error != nil {
// 		return "", fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return user.Password, nil
// }

// func (db *usersDB) CheckUserExists(name string) error {
// 	var user models.User

// 	result := db.db.Table("users").Select("id").Where("name = ?", name).Scan(&user)

// 	if result.RowsAffected == 0 {
// 		return errs.ErrRecordingWNF
// 	}

// 	if result.Error != nil {
// 		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return nil
// }

// func (db *usersDB) GetUserByName(name string) (*models.User, error) {
// 	var user models.User

// 	result := db.db.Where("name = ?", name).First(&user)

// 	if result.RowsAffected == 0 {
// 		return nil, errs.ErrRecordingWNF
// 	}

// 	if result.Error != nil {
// 		return nil, fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return &user, nil
// }

// func (db *usersDB) StoreRefreshToken(session *models.Session) error {
	
// 	result := db.db.Create(session)

// 	if result.Error != nil {
// 		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return nil
// }

// func (db *usersDB) GetByRefreshTokenHash(tokenHash string) (*models.Session, error) {
// 	var session models.Session

// 	result := db.db.Where("refresh_token = ?", tokenHash).First(&session)

// 	if result.Error != nil {
// 		return nil, fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}

// 	return &session, nil
// }

// func (db *usersDB) DeleteByRefreshTokenHash(tokenHash  string) error {
// 	result := db.db.Where("refresh_token = ?", tokenHash).Delete(&models.Session{})

// 	if err := result.Error; err != nil {
//         return fmt.Errorf("%w: %s", errs.ErrDB, err.Error())
//     }

// 	if result.RowsAffected == 0 {
// 		return errs.ErrRecordingWND
// 	}

// 	return nil
// }

// func (db *usersDB) DeleteAllUserSessions(userID uint) error {
// 	result := db.db.Where("user_id = ?", userID).Delete(&models.Session{})
// 	if result.Error != nil {
// 		return fmt.Errorf("%w: %s", errs.ErrDB, result.Error.Error())
// 	}
// 	return nil
// }