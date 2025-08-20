package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/repos"
	"github.com/oklog/ulid/v2"
)

// InMemoryContactRepository is a mock implementation of the ContactRepository for testing.
type InMemoryContactRepository struct {
	mu              sync.RWMutex
	users           map[ulid.ULID]*db.User
	contactRequests map[ulid.ULID]*db.ContactRequest
	contacts        map[ulid.ULID]map[ulid.ULID]*db.Contact
	blocks          map[ulid.ULID]map[ulid.ULID]bool
}

// NewInMemoryContactRepository creates a new in-memory repository.
func NewInMemoryContactRepository() *InMemoryContactRepository {
	// Seed with some data for testing
	user1 := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
	user2 := ulid.MustParse("01H8XGJWBXBAQ1JBS9M6S3S2A2")
	user3 := ulid.MustParse("01H8XGJWBZBAQ1JBS9M6S3S2A3")
	username1 := "user1"
	username2 := "user2"
	username3 := "user3"

	return &InMemoryContactRepository{
		users: map[ulid.ULID]*db.User{
			user1: {ID: user1, Username: &username1, Status: db.UserStatusActive},
			user2: {ID: user2, Username: &username2, Status: db.UserStatusActive},
			user3: {ID: user3, Username: &username3, Status: db.UserStatusActive},
		},
		contactRequests: make(map[ulid.ULID]*db.ContactRequest),
		contacts:        make(map[ulid.ULID]map[ulid.ULID]*db.Contact),
		blocks:          make(map[ulid.ULID]map[ulid.ULID]bool),
	}
}

var _ repos.ContactRepository = (*InMemoryContactRepository)(nil)

func (r *InMemoryContactRepository) FindUserByIdentifier(ctx context.Context, identifier string) (*db.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Username != nil && *user.Username == identifier {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user with identifier '%s' not found", identifier)
}

func (r *InMemoryContactRepository) CreateContactRequest(ctx context.Context, from, to ulid.ULID, message *string) (*db.ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	req := &db.ContactRequest{
		ID:        ulid.MustNew(ulid.MaxTime(), nil),
		FromUser:  from,
		ToUser:    to,
		State:     db.ContactStatePending,
		Message:   message,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.contactRequests[req.ID] = req
	return req, nil
}

func (r *InMemoryContactRepository) GetContactRequest(ctx context.Context, requestID ulid.ULID) (*db.ContactRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	req, ok := r.contactRequests[requestID]
	if !ok {
		return nil, fmt.Errorf("contact request not found")
	}
	return req, nil
}

func (r *InMemoryContactRepository) UpdateContactRequestState(ctx context.Context, requestID ulid.ULID, state db.ContactState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	req, ok := r.contactRequests[requestID]
	if !ok {
		return fmt.Errorf("contact request not found")
	}
	req.State = state
	req.UpdatedAt = time.Now()
	return nil
}

func (r *InMemoryContactRepository) CreateContact(ctx context.Context, ownerID, peerID ulid.ULID) (*db.Contact, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.contacts[ownerID]; !ok {
		r.contacts[ownerID] = make(map[ulid.ULID]*db.Contact)
	}
	contact := &db.Contact{
		ID:        ulid.MustNew(ulid.MaxTime(), nil),
		OwnerID:   ownerID,
		PeerID:    peerID,
		State:     db.ContactStateAccepted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.contacts[ownerID][peerID] = contact
	return contact, nil
}

func (r *InMemoryContactRepository) ListContacts(ctx context.Context, ownerID ulid.ULID, state db.ContactState) ([]db.Contact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []db.Contact
	if userContacts, ok := r.contacts[ownerID]; ok {
		for _, contact := range userContacts {
			if contact.State == state {
				result = append(result, *contact)
			}
		}
	}
	return result, nil
}

func (r *InMemoryContactRepository) DeleteContact(ctx context.Context, ownerID, peerID ulid.ULID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userContacts, ok := r.contacts[ownerID]; ok {
		delete(userContacts, peerID)
	}
	return nil
}

func (r *InMemoryContactRepository) CreateBlock(ctx context.Context, ownerID, targetID ulid.ULID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.blocks[ownerID]; !ok {
		r.blocks[ownerID] = make(map[ulid.ULID]bool)
	}
	r.blocks[ownerID][targetID] = true
	return nil
}

func (r *InMemoryContactRepository) DeleteBlock(ctx context.Context, ownerID, targetID ulid.ULID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userBlocks, ok := r.blocks[ownerID]; ok {
		delete(userBlocks, targetID)
	}
	return nil
}

func (r *InMemoryContactRepository) IsBlocked(ctx context.Context, ownerID, targetID ulid.ULID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if userBlocks, ok := r.blocks[ownerID]; ok {
		if blocked, ok := userBlocks[targetID]; ok && blocked {
			return true, nil
		}
	}
	return false, nil
}
