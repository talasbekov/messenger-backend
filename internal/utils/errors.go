// Package utils internal/utils/errors.go
package utils

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	// ErrAuthInvalidToken Auth
	ErrAuthInvalidToken  ErrorCode = "AUTH_INVALID_TOKEN"
	ErrAuthExpired       ErrorCode = "AUTH_EXPIRED"
	ErrAuthDeviceRevoked ErrorCode = "AUTH_DEVICE_REVOKED"
	ErrAuthInvalidCreds  ErrorCode = "AUTH_INVALID_CREDENTIALS"

	// ErrUserNotFound User
	ErrUserNotFound         ErrorCode = "USER_NOT_FOUND"
	ErrUserExists           ErrorCode = "USER_EXISTS"
	ErrSelfContactForbidden ErrorCode = "SELF_CONTACT_FORBIDDEN"

	// ErrContactNotFound Contacts
	ErrContactNotFound      ErrorCode = "CONTACT_NOT_FOUND"
	ErrContactRequestExists ErrorCode = "CONTACT_REQUEST_EXISTS"
	ErrPeerBlocked          ErrorCode = "PEER_BLOCKED"
	ErrYouAreBlocked        ErrorCode = "YOU_ARE_BLOCKED"

	// ErrChatNotFound Chats
	ErrChatNotFound      ErrorCode = "CHAT_NOT_FOUND"
	ErrChatAlreadyExists ErrorCode = "CHAT_ALREADY_EXISTS"
	ErrSelfChatExists    ErrorCode = "SELF_CHAT_EXISTS"
	ErrForbiddenRole     ErrorCode = "FORBIDDEN_ROLE"

	// ErrMessageNotFound Messages
	ErrMessageNotFound   ErrorCode = "MESSAGE_NOT_FOUND"
	ErrInvalidCiphertext ErrorCode = "INVALID_CIPHERTEXT"

	// ErrMemberExists Members
	ErrMemberExists   ErrorCode = "MEMBER_EXISTS"
	ErrMemberNotFound ErrorCode = "MEMBER_NOT_FOUND"

	// ErrMLSStateMismatch Keys / MLS
	ErrMLSStateMismatch ErrorCode = "MLS_STATE_MISMATCH"
	ErrKeysExhausted    ErrorCode = "KEYS_EXHAUSTED"

	// ErrCallNotFound Calls
	ErrCallNotFound ErrorCode = "CALL_NOT_FOUND"
	ErrBusyHere     ErrorCode = "BUSY_HERE"
	ErrTimeout      ErrorCode = "TIMEOUT"
	ErrCallGone     ErrorCode = "CALL_GONE"

	// ErrAttachmentTooLarge Attachments
	ErrAttachmentTooLarge ErrorCode = "ATTACHMENT_TOO_LARGE"
	ErrUnsupportedMime    ErrorCode = "UNSUPPORTED_MIME"

	// ErrRateLimited Rate limit
	ErrRateLimited          ErrorCode = "RATE_LIMITED"
	ErrRoomCapacityExceeded ErrorCode = "ROOM_CAPACITY_EXCEEDED"

	// ErrValidation Validation
	ErrValidation ErrorCode = "VALIDATION_ERROR"

	// ErrInternal Generic
	ErrInternal  ErrorCode = "INTERNAL_ERROR"
	ErrNotFound  ErrorCode = "NOT_FOUND"
	ErrConflict  ErrorCode = "CONFLICT"
	ErrForbidden ErrorCode = "FORBIDDEN"
)

type AppError struct {
	Code       ErrorCode              `json:"error_code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	StatusCode int                    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// ---- Convenience constructors ----

func NewNotFound(message string) *AppError {
	return NewError(ErrNotFound, message, http.StatusNotFound)
}

func ErrValidationFailed(details map[string]interface{}) *AppError {
	err := NewError(ErrValidation, "Validation failed", http.StatusUnprocessableEntity) // 422
	err.Details = details
	return err
}

func ErrUnauthorized(message string) *AppError {
	return NewError(ErrAuthInvalidToken, message, http.StatusUnauthorized)
}

func ErrForbiddenAccess(message string) *AppError {
	return NewError(ErrForbidden, message, http.StatusForbidden)
}

func ErrConflictResource(code ErrorCode, message string) *AppError {
	return NewError(code, message, http.StatusConflict)
}

func ErrTooManyRequests(message string) *AppError {
	return NewError(ErrRateLimited, message, http.StatusTooManyRequests)
}

func ErrInternalServer(message string) *AppError {
	return NewError(ErrInternal, message, http.StatusInternalServerError)
}
