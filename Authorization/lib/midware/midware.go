// Package midware
package midware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Wladim1r/auth/lib/getenv"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func verifyToken(c *gin.Context, token string) error {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(getenv.GetString("SECRET_KEY", "default_secret_key")), nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}
	if !jwtToken.Valid {
		return fmt.Errorf("token is not valid")
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
		return fmt.Errorf("token life time has expired")
	}

	sub, ok := claims["sub"]
	if !ok {
		return fmt.Errorf("sub claim not found")
	}

	userID, ok := sub.(string)
	if !ok {
		return fmt.Errorf("sub claim is not a number")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("sub claim is not a valid uuid: %w", err)
	}

	c.Set("userID", userUUID)
	return nil
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeadVal := c.GetHeader("Authorization")
		if authHeadVal == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeadVal, " ")
		if len(bearerToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		if err := verifyToken(c, bearerToken[1]); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
