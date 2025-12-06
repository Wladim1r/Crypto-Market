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
	userService service.UserService
	tokenService service.TokenService
}

func NewHandler(ctx context.Context, userService service.UserService, tokenService service.TokenService) *handler {
	return &handler{
		ctx: ctx,
		userService: userService,
		tokenService: tokenService,
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

	if err := h.userService.RegisterUser(req.Name, req.Password); err != nil {
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

func (h *handler) Login(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	user, err := h.userService.LoginUser(req.Name, req.Password)
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

	refreshToken, err := hashpwd.GenerateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate refresh token.",
		})
		return
	}

	hashedRefreshToken := hashpwd.HashToken(refreshToken)

	session := models.Session{
		UserID:       user.ID,
		RefreshToken: hashedRefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.tokenService.StoreRefreshToken(&session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save session",
		})
		return
	}

	c.SetCookie("userID", user.ID.String(), 3600*24*7, "/", "localhost", true, false)
	c.SetCookie("refreshToken", refreshToken, 3600*24*7, "/", "localhost", true, true)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login success!🫦",
		"access_token": accessToken,
	})

}

func (h *handler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errs.ErrInvalidToken.Error(),
		})
		return
	}

	session, err := h.tokenService.GetSessionByToken(hashpwd.HashToken(refreshToken))
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

	c.SetCookie("refreshToken", "", -1, "/", "localhost", true, true)

	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "refresh token expired😪",
		})
		return
	}

	if err := h.tokenService.DeleteSessionByToken(hashpwd.HashToken(refreshToken)); err != nil {
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

	newRefreshToken, err := hashpwd.GenerateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate refresh token.",
		})
		return
	}

	newHashedRefreshToken := hashpwd.HashToken(newRefreshToken)

	newSession := models.Session{
		UserID:       session.UserID,
		RefreshToken: newHashedRefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.tokenService.StoreRefreshToken(&newSession); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.SetCookie("refreshToken", newRefreshToken, 3600*24*7, "/", "localhost", true, true)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Good boy💋",
		"access_token": newAccessToken,
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

	if err := h.userService.DeleteUserByID(userID); err != nil {
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
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errs.ErrInvalidToken.Error(),
		})
		return
	}

	if err := h.tokenService.DeleteSessionByToken(hashpwd.HashToken(refreshToken)); err != nil {
		if !errors.Is(err, errs.ErrRecordingWND) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not invalidate session",
			})
			return
		}
	}

	c.SetCookie("userID", "", -1, "/", "localhost", true, false)
	c.SetCookie("refreshToken", "", -1, "/", "localhost", true, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "logout successful",
	})
}
