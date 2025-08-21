package services

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/messenger/backend/internal/config"
	"github.com/messenger/backend/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthService() *AuthService {
	authCfg := config.AuthConfig{
		Secret: "test-secret",
	}
	secureCfg := config.SecurityConfig{
		BCryptCost: 10,
	}
	return NewAuthService(testQueries, authCfg, secureCfg)
}

func TestRegister_Success(t *testing.T) {
	service := setupAuthService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	params := RegisterParams{
		Username: "testuser",
		Password: "password123",
	}

	user, err := service.Register(ctx, params)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)

	dbUser, err := testQueries.FindUserByIdentifier(ctx, db.FindUserByIdentifierParams{
		Username: pgtype.Text{String: "testuser", Valid: true},
	})
	require.NoError(t, err)
	assert.Equal(t, user.ID, dbUser.ID)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	service := setupAuthService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	params1 := RegisterParams{Username: "duplicate", Password: "password123"}
	_, err := service.Register(ctx, params1)
	require.NoError(t, err)

	params2 := RegisterParams{Username: "duplicate", Password: "password456"}
	_, err = service.Register(ctx, params2)
	require.Error(t, err)
}

func TestLogin_Success(t *testing.T) {
	service := setupAuthService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	regParams := RegisterParams{Username: "loginuser", Password: "password123"}
	user, err := service.Register(ctx, regParams)
	require.NoError(t, err)

	loginParams := LoginParams{Identifier: "loginuser", Password: "password123"}
	resp, err := service.Login(ctx, loginParams)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.Username, resp.User.Username)
}

func TestLogin_WrongPassword(t *testing.T) {
	service := setupAuthService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	regParams := RegisterParams{Username: "loginuser", Password: "password123"}
	_, err := service.Register(ctx, regParams)
	require.NoError(t, err)

	loginParams := LoginParams{Identifier: "loginuser", Password: "wrongpassword"}
	_, err = service.Login(ctx, loginParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestLogin_UserNotFound(t *testing.T) {
	service := setupAuthService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	loginParams := LoginParams{Identifier: "nonexistent", Password: "password"}
	_, err := service.Login(ctx, loginParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
