// Package handlers
package handlers

import (
	"net/http"

	"github.com/Wladim1r/profile/internal/models"
	"github.com/Wladim1r/proto-crypto/gen/protos/auth-portfile"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type client struct {
	conn auth.AuthClient
}

func NewClient(c auth.AuthClient) *client {
	return &client{conn: c}
}

func (cl *client) Registration(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := cl.conn.Register(
		c.Request.Context(),
		&auth.AuthRequest{Name: req.Name, Password: req.Password},
	)
	if err != nil {
		c.JSON(int(resp.GetStatus()), gin.H{
			"error": err.Error(),
		})
		return
	}

	switch resp.Status {
	case http.StatusCreated:
		c.JSON(http.StatusCreated, gin.H{
			"message": "user successful created 🎊🤩",
		})
		return
	case http.StatusConflict:
		c.JSON(http.StatusConflict, gin.H{
			"message": "user already exsited 💩",
		})
		return
	}
}

func (cl *client) Login(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var header metadata.MD

	resp, err := cl.conn.Login(
		c.Request.Context(),
		&auth.AuthRequest{Name: req.Name, Password: req.Password},
		grpc.Header(&header),
	)
	if err != nil {
		c.JSON(int(resp.GetStatus()), gin.H{
			"error": err.Error(),
		})
		return
	}

	serverHeader, ok := header["x-user-id"]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "there is no 'x-user-id' in the gRPC headers",
		})
		return
	}

	accessToken, ok := header["x-access-token"]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "there is no access token in the gRPC headers",
		})
		return
	}

	refreshToken, ok := header["x-refresh-token"]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "there is no refresh token in the gRPC headers",
		})
		return
	}

	c.SetCookie("refreshToken", refreshToken[0], 300, "/", "localhost", false, true)
	c.SetCookie("userID", serverHeader[0], 300, "/", "localhost", false, true)

	tStruct := struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}{
		Access:  accessToken[0],
		Refresh: refreshToken[0],
	}

	c.JSON(int(resp.GetStatus()), gin.H{
		"msg":                  "Login success!🫦",
		"Here is your tokens🌐": tStruct,
	})
}

func (cl *client) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "molodec! 👍",
	})
}

func (cl *client) Refresh(c *gin.Context) {
	userID, err := c.Cookie("userID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "there is no 'userID' in the cookies: " + err.Error(),
		})
		return
	}
	refresh, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "there is no 'refreshToken' in the cookies: " + err.Error(),
		})
		return
	}

	ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs(
		"x-user-id", userID,
		"x-refresh-token", refresh,
	))

	resp, err := cl.conn.Refresh(ctx, &auth.EmptyRequest{})
	if err != nil {
		c.JSON(int(resp.GetStatus()), gin.H{
			"error": err.Error(),
		})
		return
	}

	tStruct := struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}{
		Access:  resp.GetAccess(),
		Refresh: resp.GetRefresh(),
	}

	c.JSON(int(resp.GetStatus()), gin.H{
		"msg":         "user succesfully logouted",
		"your tokens": tStruct,
	})
}

func (cl *client) Logout(c *gin.Context) {
	refresh, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "there is no 'refreshToken' in the cookies: " + err.Error(),
		})
		return
	}

	authHeader := c.GetHeader("Authorization")

	ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs(
		"x-refresh-token", refresh,
		"x-authorization-header", authHeader,
	))

	resp, err := cl.conn.Logout(ctx, &auth.EmptyRequest{})
	if err != nil {
		c.JSON(int(resp.GetStatus()), gin.H{
			"error": err.Error(),
		})
		return
	}

	c.SetCookie("userID", "", -1, "/", "localhost", false, true)
	c.SetCookie("refreshToken", "", -1, "/", "localhost", false, true)

	c.JSON(int(resp.GetStatus()), gin.H{
		"msg": "user succesfully logouted",
	})
}
