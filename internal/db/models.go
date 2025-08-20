// Package db internal/db/models.go
package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// UserStatus Enums
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusDeleted  UserStatus = "deleted"
)

type ContactState string

const (
	ContactStatePending  ContactState = "pending"
	ContactStateAccepted ContactState = "accepted"
	ContactStateBlocked  ContactState = "blocked"
)

type ChatType string

const (
	ChatTypeSelf   ChatType = "self"
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type ChatMemberRole string

const (
	RoleOwner  ChatMemberRole = "owner"
	RoleAdmin  ChatMemberRole = "admin"
	RoleMember ChatMemberRole = "member"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeFile   MessageType = "file"
	MessageTypeImage  MessageType = "image"
	MessageTypeVoice  MessageType = "voice"
	MessageTypeVideo  MessageType = "video"
	MessageTypeSystem MessageType = "system"
)

type MessageStatus string

const (
	MessageStatusQueued    MessageStatus = "queued"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type CallType string

const (
	CallTypeAudio CallType = "audio"
	CallTypeVideo CallType = "video"
)

type CallMode string

const (
	CallMode1v1   CallMode = "1v1"
	CallModeGroup CallMode = "group"
)

// User Core Models
type User struct {
	ID           ulid.ULID  `db:"id"`
	Username     *string    `db:"username"`
	Phone        *string    `db:"phone"`
	Email        *string    `db:"email"`
	PasswordHash *string    `db:"password_hash"`
	Status       UserStatus `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	LastSeenAt   *time.Time `db:"last_seen_at"`
	Metadata     JSONB      `db:"metadata"`
}

type Device struct {
	ID          ulid.ULID  `db:"id"`
	UserID      ulid.ULID  `db:"user_id"`
	Name        string     `db:"name"`
	Platform    string     `db:"platform"`
	PushToken   *string    `db:"push_token"`
	Fingerprint string     `db:"fingerprint"`
	CreatedAt   time.Time  `db:"created_at"`
	RevokedAt   *time.Time `db:"revoked_at"`
}

type AuthSession struct {
	ID         ulid.ULID `db:"id"`
	UserID     ulid.ULID `db:"user_id"`
	DeviceID   ulid.ULID `db:"device_id"`
	RefreshJTI string    `db:"refresh_jti"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
	LastUsedAt time.Time `db:"last_used_at"`
}

type Contact struct {
	ID        ulid.ULID    `db:"id"`
	OwnerID   ulid.ULID    `db:"owner_id"`
	PeerID    ulid.ULID    `db:"peer_id"`
	Alias     *string      `db:"alias"`
	State     ContactState `db:"state"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

type ContactRequest struct {
	ID        ulid.ULID    `db:"id"`
	FromUser  ulid.ULID    `db:"from_user"`
	ToUser    ulid.ULID    `db:"to_user"`
	State     ContactState `db:"state"`
	Message   *string      `db:"message"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

type Chat struct {
	ID        ulid.ULID `db:"id"`
	Type      ChatType  `db:"type"`
	CreatorID ulid.ULID `db:"creator_id"`
	Title     *string   `db:"title"`
	Avatar    *string   `db:"avatar"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Metadata  JSONB     `db:"metadata"`
}

type ChatMember struct {
	ChatID   ulid.ULID      `db:"chat_id"`
	UserID   ulid.ULID      `db:"user_id"`
	Role     ChatMemberRole `db:"role"`
	JoinedAt time.Time      `db:"joined_at"`
	LeftAt   *time.Time     `db:"left_at"`
}

type Message struct {
	ID             ulid.ULID   `db:"id"`
	ChatID         ulid.ULID   `db:"chat_id"`
	SenderID       ulid.ULID   `db:"sender_id"`
	Type           MessageType `db:"type"`
	CiphertextBlob []byte      `db:"ciphertext_blob"`
	HeaderMeta     JSONB       `db:"header_meta_json"`
	SentAt         time.Time   `db:"sent_at"`
	EditedAt       *time.Time  `db:"edited_at"`
	DeletedAt      *time.Time  `db:"deleted_at"`
}

type Receipt struct {
	MessageID ulid.ULID     `db:"message_id"`
	UserID    ulid.ULID     `db:"user_id"`
	Status    MessageStatus `db:"status"`
	At        time.Time     `db:"at"`
}

type Attachment struct {
	ID         ulid.ULID `db:"id"`
	MessageID  ulid.ULID `db:"message_id"`
	StorageKey string    `db:"storage_key"`
	Size       int64     `db:"size"`
	MimeType   string    `db:"mime"`
	SHA256     string    `db:"sha256"`
	EncScheme  string    `db:"enc_scheme"`
	CreatedAt  time.Time `db:"created_at"`
}

// IdentityKey E2EE Keys
type IdentityKey struct {
	UserID    ulid.ULID  `db:"user_id"`
	DeviceID  ulid.ULID  `db:"device_id"`
	PublicKey []byte     `db:"pub"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

type SignedPreKey struct {
	DeviceID   ulid.ULID  `db:"device_id"`
	PublicKey  []byte     `db:"pub"`
	Signature  []byte     `db:"signature"`
	CreatedAt  time.Time  `db:"created_at"`
	ReplacedAt *time.Time `db:"replaced_at"`
}

type OneTimePreKey struct {
	DeviceID   ulid.ULID  `db:"device_id"`
	KeyID      int        `db:"key_id"`
	PublicKey  []byte     `db:"pub"`
	ConsumedAt *time.Time `db:"consumed_at"`
}

type MLSGroupState struct {
	ChatID    ulid.ULID `db:"chat_id"`
	Epoch     uint64    `db:"epoch"`
	StateBlob []byte    `db:"state_blob"`
	CreatedAt time.Time `db:"created_at"`
}

// Call Calls
type Call struct {
	ID        ulid.ULID  `db:"id"`
	ChatID    ulid.ULID  `db:"chat_id"`
	Type      CallType   `db:"type"`
	Mode      CallMode   `db:"mode"`
	StartedAt time.Time  `db:"started_at"`
	EndedAt   *time.Time `db:"ended_at"`
	Reason    *string    `db:"reason"`
	SFURoomID *string    `db:"sfu_room_id"`
}

type ICECandidate struct {
	CallID        ulid.ULID `db:"call_id"`
	DeviceID      ulid.ULID `db:"device_id"`
	SDPMid        string    `db:"sdp_mid"`
	SDPMLineIndex int       `db:"sdp_mline_index"`
	Candidate     string    `db:"candidate"`
	At            time.Time `db:"at"`
}

// Block Security
type Block struct {
	OwnerID   ulid.ULID `db:"owner_id"`
	TargetID  ulid.ULID `db:"target_user_id"`
	Reason    *string   `db:"reason"`
	CreatedAt time.Time `db:"created_at"`
}

type RateLimit struct {
	UserID      ulid.ULID `db:"user_id"`
	Key         string    `db:"key"`
	WindowStart time.Time `db:"window_start"`
	Count       int       `db:"count"`
}

type AuditLog struct {
	ID       ulid.ULID  `db:"id"`
	UserID   ulid.ULID  `db:"user_id"`
	DeviceID *ulid.ULID `db:"device_id"`
	Action   string     `db:"action"`
	Meta     JSONB      `db:"meta"`
	At       time.Time  `db:"at"`
}

// JSONB Helper types
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
