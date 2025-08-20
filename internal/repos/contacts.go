package repos

import (
	"context"

	"github.com/messenger/backend/internal/db"
	"github.com/oklog/ulid/v2"
)

// ContactRepository defines the interface for database operations on contacts.
// It is defined in a separate package to avoid import cycles.
type ContactRepository interface {
	// User related
	FindUserByIdentifier(ctx context.Context, identifier string) (*db.User, error)

	// Contact requests
	CreateContactRequest(ctx context.Context, from, to ulid.ULID, message *string) (*db.ContactRequest, error)
	GetContactRequest(ctx context.Context, requestID ulid.ULID) (*db.ContactRequest, error)
	UpdateContactRequestState(ctx context.Context, requestID ulid.ULID, state db.ContactState) error

	// Contacts
	CreateContact(ctx context.Context, ownerID, peerID ulid.ULID) (*db.Contact, error)
	ListContacts(ctx context.Context, ownerID ulid.ULID, state db.ContactState) ([]db.Contact, error)
	DeleteContact(ctx context.Context, ownerID, peerID ulid.ULID) error

	// Blocking
	CreateBlock(ctx context.Context, ownerID, targetID ulid.ULID) error
	DeleteBlock(ctx context.Context, ownerID, targetID ulid.ULID) error
	IsBlocked(ctx context.Context, ownerID, targetID ulid.ULID) (bool, error)
}
