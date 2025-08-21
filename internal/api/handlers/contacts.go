package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/services"
	"github.com/oklog/ulid/v2"
)

// ContactsService defines the interface for contact-related business logic.
type ContactsService interface {
	CreateContactRequest(ctx context.Context, fromUserID ulid.ULID, peerIdentifier string, message *string) (*db.ContactRequest, error)
	AcceptContactRequest(ctx context.Context, userID, requestID ulid.ULID) error
	RejectContactRequest(ctx context.Context, userID, requestID ulid.ULID) error
	DeleteContact(ctx context.Context, ownerID, contactID ulid.ULID) error
	BlockPeer(ctx context.Context, ownerID, targetID ulid.ULID) error
	UnblockPeer(ctx context.Context, ownerID, targetID ulid.ULID) error
	ListContacts(ctx context.Context, ownerID ulid.ULID, state db.ContactState) ([]db.Contact, error)
}

// ContactsHandler handles API requests related to contacts.
type ContactsHandler struct {
	service ContactsService
}

// NewContactsHandler creates a new ContactsHandler.
func NewContactsHandler(service ContactsService) *ContactsHandler {
	return &ContactsHandler{service: service}
}

// RegisterContactRoutes registers all contact-related routes with the Gin router.
func (h *ContactsHandler) RegisterContactRoutes(router *gin.RouterGroup) {
	contacts := router.Group("/contacts")
	{
		contacts.GET("", h.ListContacts)
		contacts.DELETE("/:contact_id", h.DeleteContact)

		requests := contacts.Group("/requests")
		{
			requests.POST("", h.CreateContactRequest)
			requests.POST("/:id/accept", h.AcceptContactRequest)
			requests.POST("/:id/reject", h.RejectContactRequest)
		}

		actions := contacts.Group("/actions")
		{
			actions.POST("/block/:peer_id", h.BlockPeer)
			actions.DELETE("/unblock/:peer_id", h.UnblockPeer)
		}
	}
}

// API request/response structs
type CreateContactRequestPayload struct {
	PeerIdentifier string  `json:"peer_identifier" binding:"required"`
	Message        *string `json:"message"`
}

type ErrorResponse struct {
	ErrorCode string      `json:"error_code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// getUserID is a helper to extract user ID from the context.
func getUserID(c *gin.Context) (ulid.ULID, bool) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		return ulid.ULID{}, false
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		return ulid.ULID{}, false
	}

	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		return ulid.ULID{}, false
	}

	return userID, true
}

func (h *ContactsHandler) CreateContactRequest(c *gin.Context) {
	var payload CreateContactRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: err.Error()})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	req, err := h.service.CreateContactRequest(c.Request.Context(), userID, payload.PeerIdentifier, payload.Message)
	if err != nil {
		if bizErr, ok := err.(*services.BusinessError); ok {
			// Map business errors to specific HTTP status codes
			switch bizErr.Code {
			case "SELF_CONTACT_FORBIDDEN":
				c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: bizErr.Code, Message: bizErr.Message})
			case "YOU_ARE_BLOCKED":
				c.JSON(http.StatusForbidden, ErrorResponse{ErrorCode: bizErr.Code, Message: bizErr.Message})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *ContactsHandler) AcceptContactRequest(c *gin.Context) {
	requestID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: "Invalid request ID format"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	if err := h.service.AcceptContactRequest(c.Request.Context(), userID, requestID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) RejectContactRequest(c *gin.Context) {
	requestID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: "Invalid request ID format"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	if err := h.service.RejectContactRequest(c.Request.Context(), userID, requestID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) ListContacts(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	// For simplicity, only filtering by 'accepted'. A real implementation would parse this from query params.
	state := db.ContactStateAccepted

	contacts, err := h.service.ListContacts(c.Request.Context(), userID, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": contacts})
}

func (h *ContactsHandler) DeleteContact(c *gin.Context) {
	contactID, err := ulid.Parse(c.Param("contact_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: "Invalid contact ID format"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	if err := h.service.DeleteContact(c.Request.Context(), userID, contactID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) BlockPeer(c *gin.Context) {
	peerID, err := ulid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: "Invalid peer ID format"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	if err := h.service.BlockPeer(c.Request.Context(), userID, peerID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) UnblockPeer(c *gin.Context) {
	peerID, err := ulid.Parse(c.Param("peer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{ErrorCode: "VALIDATION_ERROR", Message: "Invalid peer ID format"})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{ErrorCode: "UNAUTHORIZED", Message: "User ID not found in context"})
		return
	}

	if err := h.service.UnblockPeer(c.Request.Context(), userID, peerID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{ErrorCode: "INTERNAL_ERROR", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
