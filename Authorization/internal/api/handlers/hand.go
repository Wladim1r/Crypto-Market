// Package handlers
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Wladim1r/auth/internal/api/service"
	"github.com/Wladim1r/auth/internal/models"
	"github.com/Wladim1r/auth/lib/errs"
	"github.com/Wladim1r/auth/lib/getenv"
	"github.com/Wladim1r/auth/lib/hashpwd"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type handler struct {
	ctx context.Context
	s   service.Service
}

func NewHandler(ctx context.Context, service service.Service) *handler {
	return &handler{
		ctx: ctx,
		s:   service,
		// rdb: rdb,
	}
}

func (h *handler) Registration(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.s.RegisterUser(req.Name, req.Password); err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "user already exists💩",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully🎊🤩"})

}

func createJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString([]byte(getenv.GetString("SECRET_KEY", "default_secret_key")))
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt: %w", err)
	}

	return signedToken, nil
}

func createRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString([]byte(getenv.GetString("SECRET_KEY", "default_secret_key")))

	if err != nil {
		return "", fmt.Errorf("failed to sign refresh jwt: %w", err)
	}

	return signedToken, nil
}

func (h *handler) Login(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	user, err := h.s.LoginUser(req.Name, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrRecordingWNF) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	accessToken, err := createJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	refreshToken, err := createRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	session := models.Session{
		UserID:       user.ID,
		RefreshToken: hashpwd.HashToken(refreshToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.s.StoreRefreshToken(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login success!🫦",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *handler) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request😕",
		})
		return
	}

	session, err := h.s.GetSessionByToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, errs.ErrRecordingWNF) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid or expired refresh token😱",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error on getting token"})
		}
	}

	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "refresh token expired😪",
		})
		return
	}

	if err := h.s.DeleteSessionByToken(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not delete old session😧",
		})
		return
	}

	newAccessToken, err := createJWT(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	newRefreshToken, err := createRefreshToken(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	newSession := models.Session{
		UserID:       session.UserID,
		RefreshToken: hashpwd.HashToken(newRefreshToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.s.StoreRefreshToken(&newSession); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Good boy💋",
		"access_token":  newAccessToken,
		"refrash_token": newRefreshToken,
	})
}

func (h *handler) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "molodec! 👍",
	})
}

func (h *handler) Delacc(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid token claims",
		})
		return
	}

	userID := userIDAny.(uuid.UUID)

	if err := h.s.DeleteUserByID(userID); err != nil {
		if errors.Is(err, errs.ErrRecordingWND) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user to delete not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "👍 user account and all sessions successfully deleted",
	})
}

func (h *handler) Logout(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if err := h.s.DeleteSessionByToken(req.RefreshToken); err != nil {
		if !errors.Is(err, errs.ErrRecordingWND) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not invalidate session",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logout successful",
	})
}
