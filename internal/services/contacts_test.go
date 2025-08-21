package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/storage/postgres"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRealService initializes the service with a real repository connected to the test DB.
func setupRealService() *ContactsService {
	repo := postgres.NewPostgresContactRepository(testQueries)
	// We need to implement the full repos.ContactRepository, not just the part for contacts.
	// This will require creating a more complete repository implementation.
	// For now, let's assume the contacts part is sufficient.
	// Let's create a more complete repository interface implementation for the test.

	type combinedRepository struct {
		*postgres.PostgresContactRepository
		q *db.Queries
	}

	// This is a bit of a hack. The service expects a single repository.
	// The repository interface I defined earlier is incomplete.
	// Let's fix the service to take the querier directly for now.
	// No, that's bad design. The service should depend on the repo interface.
	// I'll stick with the original design. The PostgresContactRepository implements the full interface.

	return NewContactsService(repo)
}

// createUser is a test helper to insert a user and return their ID.
func createUser(t *testing.T, ctx context.Context, username string) ulid.ULID {
	t.Helper()
	user, err := testQueries.CreateUser(ctx, db.CreateUserParams{
		ID:             ulid.Make().String(),
		Username:       username,
		HashedPassword: "password", // Not used in these tests
	})
	require.NoError(t, err)
	id, err := ulid.Parse(user.ID)
	require.NoError(t, err)
	return id
}

func TestCreateContactRequest_Success_RealDB(t *testing.T) {
	service := setupRealService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	user1ID := createUser(t, ctx, "user1")
	_ = createUser(t, ctx, "user2") // peer
	message := "Hi!"
	var nullMessage = sql.NullString{String: message, Valid: true}

	req, err := service.CreateContactRequest(ctx, user1ID, "user2", &message)

	require.NoError(t, err)
	assert.NotNil(t, req)

	// Verify in DB
	dbReq, err := testQueries.GetContactRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, user1ID.String(), dbReq.FromUserID)
	assert.Equal(t, nullMessage, dbReq.Message)
	assert.Equal(t, db.ContactRequestStatePending, dbReq.State)
}

func TestCreateContactRequest_AddSelf_RealDB(t *testing.T) {
	service := setupRealService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	user1ID := createUser(t, ctx, "user1")

	_, err := service.CreateContactRequest(ctx, user1ID, "user1", nil)

	require.Error(t, err)
	bizErr, ok := err.(*BusinessError)
	require.True(t, ok)
	assert.Equal(t, "SELF_CONTACT_FORBIDDEN", bizErr.Code)
}

func TestAcceptContactRequest_Success_RealDB(t *testing.T) {
	service := setupRealService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	user1ID := createUser(t, ctx, "user1")
	user2ID := createUser(t, ctx, "user2")

	// 1. Create a request from user1 to user2
	req, err := service.CreateContactRequest(ctx, user1ID, "user2", nil)
	require.NoError(t, err)
	reqID, err := ulid.Parse(req.ID)
	require.NoError(t, err)

	// 2. user2 accepts the request
	err = service.AcceptContactRequest(ctx, user2ID, reqID)
	require.NoError(t, err)

	// 3. Verify contacts are created for both users
	contacts1, err := testQueries.ListContacts(ctx, db.ListContactsParams{OwnerID: user1ID.String(), State: db.ContactStateAccepted})
	require.NoError(t, err)
	assert.Len(t, contacts1, 1)
	assert.Equal(t, user2ID.String(), contacts1[0].PeerID)

	contacts2, err := testQueries.ListContacts(ctx, db.ListContactsParams{OwnerID: user2ID.String(), State: db.ContactStateAccepted})
	require.NoError(t, err)
	assert.Len(t, contacts2, 1)
	assert.Equal(t, user1ID.String(), contacts2[0].PeerID)

	// 4. Verify request state is updated
	updatedReq, err := testQueries.GetContactRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ContactRequestStateAccepted, updatedReq.State)
}

func TestBlockPeer_RealDB(t *testing.T) {
	service := setupRealService()
	ctx := context.Background()
	require.NoError(t, truncateTables(ctx, testPool))

	user1ID := createUser(t, ctx, "user1")
	user2ID := createUser(t, ctx, "user2")

	// user1 blocks user2
	err := service.BlockPeer(ctx, user1ID, user2ID)
	require.NoError(t, err)

	// Verify block is created
	isBlocked, err := testQueries.IsBlocked(ctx, db.IsBlockedParams{
		OwnerID:      user1ID.String(),
		TargetUserID: user2ID.String(),
	})
	require.NoError(t, err)
	assert.True(t, isBlocked)

	// Verify user2 cannot create a contact request to user1
	_, err = service.CreateContactRequest(ctx, user2ID, "user1", nil)
	require.Error(t, err)
	bizErr, ok := err.(*BusinessError)
	require.True(t, ok)
	assert.Equal(t, "YOU_ARE_BLOCKED", bizErr.Code)
}
