package services

import (
	"context"
	"database/sql"
	"fmt"

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
	nullIdentifier := sql.NullString{String: peerIdentifier, Valid: true}
	peer, err := s.repo.FindUserByIdentifier(ctx, nullIdentifier, nullIdentifier, nullIdentifier)
	if err != nil {
		// Handle not found error specifically
		return nil, &BusinessError{Code: "USER_NOT_FOUND", Message: "User not found"}
	}

	peerID, err := ulid.Parse(peer.ID)
	if err != nil {
		return nil, fmt.Errorf("internal: failed to parse peer ID: %w", err)
	}

	if fromUserID == peerID {
		return nil, &BusinessError{Code: "SELF_CONTACT_FORBIDDEN", Message: "Cannot add yourself as a contact"}
	}

	// Check if peer has blocked requester
	blocked, err := s.repo.IsBlocked(ctx, peerID, fromUserID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, &BusinessError{Code: "YOU_ARE_BLOCKED", Message: "This user has blocked you"}
	}

	var nullMessage sql.NullString
	if message != nil {
		nullMessage = sql.NullString{String: *message, Valid: true}
	}

	return s.repo.CreateContactRequest(ctx, fromUserID, peerID, nullMessage)
}

// AcceptContactRequest accepts a pending contact request.
func (s *ContactsService) AcceptContactRequest(ctx context.Context, userID, requestID ulid.ULID) error {
	req, err := s.repo.GetContactRequest(ctx, requestID)
	if err != nil {
		return err
	}

	toUserID, err := ulid.Parse(req.ToUserID)
	if err != nil {
		return fmt.Errorf("internal: failed to parse to_user_id: %w", err)
	}

	// Ensure the user accepting is the recipient of the request
	if toUserID != userID {
		return &BusinessError{Code: "FORBIDDEN", Message: "You are not authorized to accept this request"}
	}

	if req.State != db.ContactRequestStatePending {
		return &BusinessError{Code: "REQUEST_NOT_PENDING", Message: "Request is not in a pending state"}
	}

	fromUserID, err := ulid.Parse(req.FromUserID)
	if err != nil {
		return fmt.Errorf("internal: failed to parse from_user_id: %w", err)
	}

	// Create reciprocal contact entries
	if _, err := s.repo.CreateContact(ctx, fromUserID, toUserID); err != nil {
		return err
	}
	if _, err := s.repo.CreateContact(ctx, toUserID, fromUserID); err != nil {
		// Attempt to roll back the first creation
		_ = s.repo.DeleteContact(ctx, fromUserID, toUserID)
		return err
	}

	return s.repo.UpdateContactRequestState(ctx, requestID, db.ContactRequestStateAccepted)
}

// RejectContactRequest rejects a pending contact request.
func (s *ContactsService) RejectContactRequest(ctx context.Context, userID, requestID ulid.ULID) error {
	req, err := s.repo.GetContactRequest(ctx, requestID)
	if err != nil {
		return err
	}

	toUserID, err := ulid.Parse(req.ToUserID)
	if err != nil {
		return fmt.Errorf("internal: failed to parse to_user_id: %w", err)
	}

	// Ensure the user rejecting is the recipient of the request
	if toUserID != userID {
		return &BusinessError{Code: "FORBIDDEN", Message: "You are not authorized to reject this request"}
	}
	if req.State != db.ContactRequestStatePending {
		return &BusinessError{Code: "REQUEST_NOT_PENDING", Message: "Request is not in a pending state"}
	}

	return s.repo.UpdateContactRequestState(ctx, requestID, db.ContactRequestStateRejected)
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
