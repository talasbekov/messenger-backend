package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/repos"
	"github.com/oklog/ulid/v2"
)

// PostgresContactRepository is a PostgreSQL implementation of the ContactRepository.
type PostgresContactRepository struct {
	q *db.Queries
}

// NewPostgresContactRepository creates a new instance of PostgresContactRepository.
func NewPostgresContactRepository(d *db.Queries) *PostgresContactRepository {
	return &PostgresContactRepository{q: d}
}

// Statically check that PostgresContactRepository implements ContactRepository.
var _ repos.ContactRepository = (*PostgresContactRepository)(nil)

func (r *PostgresContactRepository) FindUserByIdentifier(ctx context.Context, username, email, phone sql.NullString) (*db.User, error) {
	user, err := r.q.FindUserByIdentifier(ctx, db.FindUserByIdentifierParams{
		Username: pgtype.Text{String: username.String, Valid: username.Valid},
		Email:    pgtype.Text{String: email.String, Valid: email.Valid},
		Phone:    pgtype.Text{String: phone.String, Valid: phone.Valid},
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresContactRepository) CreateContactRequest(ctx context.Context, from, to ulid.ULID, message sql.NullString) (*db.ContactRequest, error) {
	newID, _ := ulid.New(ulid.Now(), nil)
	req, err := r.q.CreateContactRequest(ctx, db.CreateContactRequestParams{
		ID:         newID.String(),
		FromUserID: from.String(),
		ToUserID:   to.String(),
		Message:    pgtype.Text{String: message.String, Valid: message.Valid},
	})
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *PostgresContactRepository) GetContactRequest(ctx context.Context, requestID ulid.ULID) (*db.ContactRequest, error) {
	req, err := r.q.GetContactRequest(ctx, requestID.String())
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *PostgresContactRepository) UpdateContactRequestState(ctx context.Context, requestID ulid.ULID, state db.ContactRequestState) error {
	return r.q.UpdateContactRequestState(ctx, db.UpdateContactRequestStateParams{
		ID:    requestID.String(),
		State: state,
	})
}

func (r *PostgresContactRepository) CreateContact(ctx context.Context, ownerID, peerID ulid.ULID) (*db.Contact, error) {
	newID, _ := ulid.New(ulid.Now(), nil)
	contact, err := r.q.CreateContact(ctx, db.CreateContactParams{
		ID:      newID.String(),
		OwnerID: ownerID.String(),
		PeerID:  peerID.String(),
	})
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *PostgresContactRepository) ListContacts(ctx context.Context, ownerID ulid.ULID, state db.ContactState) ([]db.Contact, error) {
	return r.q.ListContacts(ctx, db.ListContactsParams{
		OwnerID: ownerID.String(),
		State:   state,
	})
}

func (r *PostgresContactRepository) DeleteContact(ctx context.Context, ownerID, peerID ulid.ULID) error {
	return r.q.DeleteContact(ctx, db.DeleteContactParams{
		OwnerID: ownerID.String(),
		PeerID:  peerID.String(),
	})
}

func (r *PostgresContactRepository) CreateBlock(ctx context.Context, ownerID, targetID ulid.ULID) error {
	return r.q.CreateBlock(ctx, db.CreateBlockParams{
		OwnerID:      ownerID.String(),
		TargetUserID: targetID.String(),
	})
}

func (r *PostgresContactRepository) DeleteBlock(ctx context.Context, ownerID, targetID ulid.ULID) error {
	return r.q.DeleteBlock(ctx, db.DeleteBlockParams{
		OwnerID:      ownerID.String(),
		TargetUserID: targetID.String(),
	})
}

func (r *PostgresContactRepository) IsBlocked(ctx context.Context, ownerID, targetID ulid.ULID) (bool, error) {
	return r.q.IsBlocked(ctx, db.IsBlockedParams{
		OwnerID:      ownerID.String(),
		TargetUserID: targetID.String(),
	})
}
