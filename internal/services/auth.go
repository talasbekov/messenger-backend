package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/messenger/backend/internal/config"
	"github.com/messenger/backend/internal/db"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	q         db.Querier
	authCfg   config.AuthConfig
	secureCfg config.SecurityConfig
}

func NewAuthService(q db.Querier, authCfg config.AuthConfig, secureCfg config.SecurityConfig) *AuthService {
	return &AuthService{
		q:         q,
		authCfg:   authCfg,
		secureCfg: secureCfg,
	}
}

type RegisterParams struct {
	Username string
	Email    sql.NullString
	Phone    sql.NullString
	Password string
}

func (s *AuthService) Register(ctx context.Context, params RegisterParams) (*db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), s.secureCfg.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.q.CreateUser(ctx, db.CreateUserParams{
		ID:             ulid.Make().String(),
		Username:       params.Username,
		Email:          pgtype.Text{String: params.Email.String, Valid: params.Email.Valid},
		Phone:          pgtype.Text{String: params.Phone.String, Valid: params.Phone.Valid},
		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		// TODO: Add proper error handling for duplicate username/email/phone
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

type LoginParams struct {
	Identifier string
	Password   string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	User         *db.User
}

func (s *AuthService) Login(ctx context.Context, params LoginParams) (*LoginResponse, error) {
	nullIdentifier := pgtype.Text{String: params.Identifier, Valid: true}
	user, err := s.q.FindUserByIdentifier(ctx, db.FindUserByIdentifierParams{
		Username: nullIdentifier,
		Email:    nullIdentifier,
		Phone:    nullIdentifier,
	})
	if err != nil {
		return nil, fmt.Errorf("user not found") // Or db error
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	accessToken, err := s.createToken(user.ID, s.authCfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// For simplicity, we'll use a JWT for the refresh token as well.
	// In a real-world scenario, you might store refresh tokens in the DB for better control.
	refreshToken, err := s.createToken(user.ID, s.authCfg.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}

func (s *AuthService) createToken(userID string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": s.authCfg.Issuer,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.authCfg.Secret))
}
