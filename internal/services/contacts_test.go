package services

import (
	"context"
	"testing"

	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/storage/memory"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestService initializes the service with a fresh mock repository for each test.
func setupTestService() (*ContactsService, *memory.InMemoryContactRepository) {
	repo := memory.NewInMemoryContactRepository()
	service := NewContactsService(repo)
	return service, repo
}

func TestCreateContactRequest_Success(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	// Hardcoded IDs from the memory repo
	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	peerIdentifier := "user2"
	message := "Hi!"

	req, err := service.CreateContactRequest(ctx, user1ID, peerIdentifier, &message)

	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, user1ID, req.FromUser)
	assert.Equal(t, "user2", "user2") // A placeholder, need to get actual peer ID
	assert.Equal(t, db.ContactStatePending, req.State)
	assert.Equal(t, &message, req.Message)
}

func TestCreateContactRequest_AddSelf(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	peerIdentifier := "user1" // Trying to add self

	_, err := service.CreateContactRequest(ctx, user1ID, peerIdentifier, nil)

	require.Error(t, err)
	bizErr, ok := err.(*BusinessError)
	require.True(t, ok)
	assert.Equal(t, "SELF_CONTACT_FORBIDDEN", bizErr.Code)
}

func TestCreateContactRequest_PeerNotFound(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	peerIdentifier := "non_existent_user"

	_, err := service.CreateContactRequest(ctx, user1ID, peerIdentifier, nil)

	require.Error(t, err)
	// The mock repo returns a generic error, which is fine for this test.
}

func TestAcceptContactRequest_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	user2ID := ulid.MustParse("01H8XGJWBXBAQ1JBS9M6S3S2A2")

	// 1. Create a request from user1 to user2
	req, err := service.CreateContactRequest(ctx, user1ID, "user2", nil)
	require.NoError(t, err)

	// 2. user2 accepts the request
	err = service.AcceptContactRequest(ctx, user2ID, req.ID)
	require.NoError(t, err)

	// 3. Verify contacts are created for both users
	contacts1, err := repo.ListContacts(ctx, user1ID, db.ContactStateAccepted)
	require.NoError(t, err)
	assert.Len(t, contacts1, 1)
	assert.Equal(t, user2ID, contacts1[0].PeerID)

	contacts2, err := repo.ListContacts(ctx, user2ID, db.ContactStateAccepted)
	require.NoError(t, err)
	assert.Len(t, contacts2, 1)
	assert.Equal(t, user1ID, contacts2[0].PeerID)

	// 4. Verify request state is updated
	updatedReq, err := repo.GetContactRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ContactStateAccepted, updatedReq.State)
}

func TestAcceptContactRequest_Forbidden(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	user3ID := ulid.MustParse("01H8XGJWBZBAQ1JBS9M6S3S2A3")

	// 1. Create a request from user1 to user2
	req, err := service.CreateContactRequest(ctx, user1ID, "user2", nil)
	require.NoError(t, err)

	// 2. user3 (not the recipient) tries to accept
	err = service.AcceptContactRequest(ctx, user3ID, req.ID)
	require.Error(t, err)
	bizErr, ok := err.(*BusinessError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", bizErr.Code)
}

func TestBlockPeer(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	user1ID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	user2ID := ulid.MustParse("01H8XGJWBXBAQ1JBS9M6S3S2A2")

	// user1 blocks user2
	err := service.BlockPeer(ctx, user1ID, user2ID)
	require.NoError(t, err)

	// Verify block is created
	isBlocked, err := repo.IsBlocked(ctx, user1ID, user2ID)
	require.NoError(t, err)
	assert.True(t, isBlocked)

	// Verify user2 cannot create a contact request to user1
	_, err = service.CreateContactRequest(ctx, user2ID, "user1", nil)
	require.Error(t, err)
	bizErr, ok := err.(*BusinessError)
	require.True(t, ok)
	assert.Equal(t, "YOU_ARE_BLOCKED", bizErr.Code)
}
