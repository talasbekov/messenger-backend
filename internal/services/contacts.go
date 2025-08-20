package services

import (
	"context"

	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/repos"
	"github.com/oklog/ulid/v2"
)

// ContactsService provides business logic for contact management.
type ContactsService struct {
	repo repos.ContactRepository
}

// NewContactsService creates a new ContactsService.
func NewContactsService(repo repos.ContactRepository) *ContactsService {
	return &ContactsService{repo: repo}
}

// CreateContactRequest initiates a new contact request.
func (s *ContactsService) CreateContactRequest(ctx context.Context, fromUserID ulid.ULID, peerIdentifier string, message *string) (*db.ContactRequest, error) {
	// Business rules from ТЗ:
	// - Find user by username/phone/email
	// - Can't add self
	// - Check for blocks

	peer, err := s.repo.FindUserByIdentifier(ctx, peerIdentifier)
	if err != nil {
		// Handle not found error specifically
		return nil, err // e.g., ErrUserNotFound
	}

	if fromUserID == peer.ID {
		return nil, &BusinessError{Code: "SELF_CONTACT_FORBIDDEN", Message: "Cannot add yourself as a contact"}
	}

	// Check if peer has blocked requester
	blocked, err := s.repo.IsBlocked(ctx, peer.ID, fromUserID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, &BusinessError{Code: "YOU_ARE_BLOCKED", Message: "This user has blocked you"}
	}

	// TODO: Check for existing pending request to prevent duplicates (409 CONTACT_REQUEST_EXISTS)

	return s.repo.CreateContactRequest(ctx, fromUserID, peer.ID, message)
}

// AcceptContactRequest accepts a pending contact request.
func (s *ContactsService) AcceptContactRequest(ctx context.Context, userID, requestID ulid.ULID) error {
	req, err := s.repo.GetContactRequest(ctx, requestID)
	if err != nil {
		return err
	}

	// Ensure the user accepting is the recipient of the request
	if req.ToUser != userID {
		return &BusinessError{Code: "FORBIDDEN", Message: "You are not authorized to accept this request"}
	}

	if req.State != db.ContactStatePending {
		return &BusinessError{Code: "REQUEST_NOT_PENDING", Message: "Request is not in a pending state"}
	}

	// Create reciprocal contact entries
	if _, err := s.repo.CreateContact(ctx, req.FromUser, req.ToUser); err != nil {
		return err
	}
	if _, err := s.repo.CreateContact(ctx, req.ToUser, req.FromUser); err != nil {
		// Attempt to roll back the first creation
		_ = s.repo.DeleteContact(ctx, req.FromUser, req.ToUser)
		return err
	}

	return s.repo.UpdateContactRequestState(ctx, requestID, db.ContactStateAccepted)
}

// RejectContactRequest rejects a pending contact request.
func (s *ContactsService) RejectContactRequest(ctx context.Context, userID, requestID ulid.ULID) error {
	req, err := s.repo.GetContactRequest(ctx, requestID)
	if err != nil {
		return err
	}
	// Ensure the user rejecting is the recipient of the request
	if req.ToUser != userID {
		return &BusinessError{Code: "FORBIDDEN", Message: "You are not authorized to reject this request"}
	}
	if req.State != db.ContactStatePending {
		return &BusinessError{Code: "REQUEST_NOT_PENDING", Message: "Request is not in a pending state"}
	}

	// In a real app, we might want a "rejected" state, but for now we just delete/ignore.
	// For simplicity, we'll just update the state.
	return s.repo.UpdateContactRequestState(ctx, requestID, "rejected") // Assuming a 'rejected' state exists
}

// DeleteContact removes a contact from the user's list.
func (s *ContactsService) DeleteContact(ctx context.Context, ownerID, contactID ulid.ULID) error {
	// This should remove the relationship from one side.
	return s.repo.DeleteContact(ctx, ownerID, contactID)
}

// BlockPeer blocks another user.
func (s *ContactsService) BlockPeer(ctx context.Context, ownerID, targetID ulid.ULID) error {
	// Also remove existing contact relationship if any
	_ = s.repo.DeleteContact(ctx, ownerID, targetID)
	_ = s.repo.DeleteContact(ctx, targetID, ownerID)
	return s.repo.CreateBlock(ctx, ownerID, targetID)
}

// UnblockPeer unblocks another user.
func (s *ContactsService) UnblockPeer(ctx context.Context, ownerID, targetID ulid.ULID) error {
	return s.repo.DeleteBlock(ctx, ownerID, targetID)
}

// ListContacts retrieves a user's contacts.
func (s *ContactsService) ListContacts(ctx context.Context, ownerID ulid.ULID, state db.ContactState) ([]db.Contact, error) {
	return s.repo.ListContacts(ctx, ownerID, state)
}

// BusinessError for custom error types
type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}
