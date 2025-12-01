package handlers

import (
	"fmt"
	"strings"
	"time"

	repo "github.com/Wladim1r/auth/internal/api/repository"
	"github.com/Wladim1r/auth/lib/errs"
	"github.com/Wladim1r/auth/lib/getenv"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func verifyToken(token string, r repo.UserRepository) error {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(getenv.GetString("SECRET_KEY", "default_secret_key")), nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}
	if !jwtToken.Valid {
		return errs.ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("could not parse jwtToken into MapClaims struct")
	}

	expRaw, ok := claims["exp"]
	if !ok {
		return fmt.Errorf("field 'expiration' did not found")
	}
	exp64, ok := expRaw.(float64)
	if !ok {
		return fmt.Errorf("failed to parse expRaw into int64")
	}
	exp := int64(exp64)

	if time.Now().Unix() > exp {
		return errs.ErrTokenTTL
	}

	userIDRaw, ok := claims["sub"]
	if !ok {
		return fmt.Errorf("field 'userID' did not found")
	}
	userID, ok := userIDRaw.(float64)
	if !ok {
		return fmt.Errorf("failed to parse 'userIDRaw' into float64")
	}

	return r.CheckUserExistsByID(uint(userID))
}

func checkAuth(authHeadVal string, r repo.UserRepository) error {
	if authHeadVal == "" {
		return errs.ErrEmptyAuthHeader
	}

	bearerToken := strings.Split(authHeadVal, " ")
	if len(bearerToken) != 2 {
		return fmt.Errorf("Invalid token format")
	}

	if err := verifyToken(bearerToken[1], r); err != nil {
		return err
	}

	return nil
}

func checkUserExists(name, pwd string, db repo.UserRepository) (uint, error) {
	userID, password, err := db.SelectPwdByName(name)
	if err != nil {
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(pwd)); err != nil {
		return 0, fmt.Errorf("🚫🟰 passwords not equal: %w", err)
	}

	return userID, nil
}
