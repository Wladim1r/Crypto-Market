// Package handlers
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Wladim1r/auth/internal/api/repository"
	"github.com/Wladim1r/auth/internal/api/service"
	"github.com/Wladim1r/auth/lib/errs"
	"github.com/Wladim1r/auth/lib/getenv"
	"github.com/Wladim1r/auth/lib/hashpwd"
	"github.com/Wladim1r/proto-crypto/gen/protos/auth-portfile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type handler struct {
	auth.UnimplementedAuthServer
	us service.UserService
	ts service.TokenService
	ur repository.UserRepository
}

func RegisterServer(
	gRPC *grpc.Server,
	uServ service.UserService,
	tServ service.TokenService,
	uRepo repository.UserRepository,
) {
	auth.RegisterAuthServer(gRPC, &handler{
		us: uServ,
		ts: tServ,
		ur: uRepo,
	})
}

func (h *handler) Register(ctx context.Context, req *auth.AuthRequest) (*auth.AuthResponse, error) {
	name := req.GetName()
	password := req.GetPassword()

	err := h.us.CheckUserExistsByName(name)

	if err != nil {
		switch {
		case errors.Is(err, errs.ErrRecordingWNF):
			hashPwd, err := hashpwd.HashPwd([]byte(password))
			if err != nil {
				return &auth.AuthResponse{
					Status: http.StatusInternalServerError,
				}, fmt.Errorf("Could not hash password: %w", err)
			}

			if err := h.us.CreateUser(name, hashPwd); err != nil {
				switch {
				case errors.Is(err, errs.ErrRecordingWNC):
					return &auth.AuthResponse{
						Status: http.StatusInternalServerError,
					}, fmt.Errorf("Could not create user rawsAffected=0: %w", err)

				default:
					return &auth.AuthResponse{
						Status: http.StatusInternalServerError,
					}, fmt.Errorf("Could not create user: %w", err)
				}
			}

			return &auth.AuthResponse{
				Status: http.StatusCreated,
			}, nil

		default:
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		}
	}

	return &auth.AuthResponse{
		Status: http.StatusConflict,
	}, nil
}

func (h *handler) Login(ctx context.Context, req *auth.AuthRequest) (*auth.AuthResponse, error) {
	name := req.GetName()
	password := req.GetPassword()

	userID, err := checkUserExists(name, password, h.ur)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrRecordingWNF):
			return &auth.AuthResponse{
				Status: http.StatusNotFound,
			}, err
		case errors.Is(err, errs.ErrDB):
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		default:
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("unknown error: %w", err)
		}
	}

	refreshTTL := getenv.GetTime("REFRESH_TTL", 150*time.Second)

	access, refresh, err := h.ts.SaveToken(userID, time.Now().Add(refreshTTL))
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrRecordingWNF):
			return &auth.AuthResponse{
				Status: http.StatusNotFound,
			}, fmt.Errorf("could not found user: %w", err)
		case errors.Is(err, errs.ErrRecordingWNC):
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("could not create token: %w", err)
		case errors.Is(err, errs.ErrDB):
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		case errors.Is(err, errs.ErrSignToken):
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("jwt error: %w", err)
		default:
			return &auth.AuthResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("unknown error: %w", err)
		}
	}

	userIDstr := strconv.Itoa(int(userID))

	userIDHeader := metadata.Pairs(
		"x-user-id", userIDstr,
		"x-access-token", access,
		"x-refresh-token", refresh,
	)
	grpc.SendHeader(ctx, userIDHeader)

	return &auth.AuthResponse{
		Status: http.StatusOK,
	}, nil
}

func (h *handler) Refresh(
	ctx context.Context,
	req *auth.EmptyRequest,
) (*auth.TokenResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &auth.TokenResponse{
			Status: http.StatusInternalServerError,
		}, fmt.Errorf("could not found metadata in grpc context")
	}

	userIDstr := md.Get("x-user-id")
	userID, err := strconv.Atoi(userIDstr[0])
	if err != nil {
		panic(err)
	}

	refreshToken := md.Get("x-refresh-token")[0]

	refreshTTL := getenv.GetTime("REFRESH_TTL", 150*time.Second)

	access, refresh, err := h.ts.RefreshAccessToken(
		refreshToken,
		userID,
		time.Now().Add(refreshTTL),
	)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrRecordingWNF):
			return &auth.TokenResponse{
				Status: http.StatusNotFound,
			}, fmt.Errorf("could not found user: %w", err)
		case errors.Is(err, errs.ErrTokenTTL):
			return &auth.TokenResponse{
				Status: http.StatusForbidden,
			}, err
		case errors.Is(err, errs.ErrRecordingWNC):
			return &auth.TokenResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("could not create token: %w", err)
		case errors.Is(err, errs.ErrDB):
			return &auth.TokenResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		case errors.Is(err, errs.ErrSignToken):
			return &auth.TokenResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("jwt error: %w", err)
		default:
			return &auth.TokenResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("unknown error: %w", err)
		}
	}

	return &auth.TokenResponse{
		Status:  http.StatusOK,
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (h *handler) Logout(ctx context.Context, req *auth.EmptyRequest) (*auth.LoginResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &auth.LoginResponse{
			Status: http.StatusInternalServerError,
		}, fmt.Errorf("could not found metadata in grpc context")
	}

	refreshToken := md.Get("x-refresh-token")[0]
	authHeader := md.Get("x-authorization-header")[0]

	if err := checkAuth(authHeader, h.ur); err != nil {
		switch {
		case errors.Is(err, errs.ErrEmptyAuthHeader):
			return &auth.LoginResponse{
				Status: http.StatusUnauthorized,
			}, fmt.Errorf("user is not registered: %w", err)
		case errors.Is(err, errs.ErrInvalidToken):
			return &auth.LoginResponse{
				Status: http.StatusForbidden,
			}, err
		case errors.Is(err, errs.ErrTokenTTL):
			return &auth.LoginResponse{
				Status: http.StatusForbidden,
			}, err
		case errors.Is(err, errs.ErrRecordingWNF):
			return &auth.LoginResponse{
				Status: http.StatusNotFound,
			}, fmt.Errorf("user does not exist: %w", err)
		case errors.Is(err, errs.ErrDB):
			return &auth.LoginResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		default:
			return &auth.LoginResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("unknown error: %w", err)
		}
	}

	if err := h.ts.DeleteToken(refreshToken); err != nil {
		switch {
		case errors.Is(err, errs.ErrDB):
			return &auth.LoginResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("db error: %w", err)
		default:
			return &auth.LoginResponse{
				Status: http.StatusInternalServerError,
			}, fmt.Errorf("unknown error: %w", err)
		}
	}

	return &auth.LoginResponse{
		Status: http.StatusOK,
	}, nil
}
