package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// AuditHandler handles HTTP requests for audit logs.
type AuditHandler struct {
	auditSvc ports.AuditService
}

// NewAuditHandler creates a new audit handler instance.
func NewAuditHandler(auditSvc ports.AuditService) *AuditHandler {
	return &AuditHandler{
		auditSvc: auditSvc,
	}
}

// AuditLogResponse represents the response body for an audit log.
type AuditLogResponse struct {
	ID          int64                  `json:"id"`
	Action      string                 `json:"action"`
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
	PrevHash    string                 `json:"prev_hash"`
	CurrentHash string                 `json:"current_hash"`
}

// ListAuditLogs handles GET /audit-logs
func (h *AuditHandler) ListAuditLogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	logs, err := h.auditSvc.GetAuditLogs(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list audit logs",
		})
	}

	responses := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, h.toResponse(log))
	}

	return c.JSON(fiber.Map{
		"audit_logs": responses,
		"limit":      limit,
		"offset":     offset,
	})
}

// VerifyAuditChain handles GET /audit-logs/verify
func (h *AuditHandler) VerifyAuditChain(c *fiber.Ctx) error {
	valid, err := h.auditSvc.VerifyAuditChain(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify audit chain",
		})
	}

	return c.JSON(fiber.Map{
		"valid": valid,
	})
}

// toResponse converts a domain audit log to a response DTO.
func (h *AuditHandler) toResponse(log *domain.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:          log.ID,
		Action:      log.Action,
		UserID:      log.UserID,
		Timestamp:   log.Timestamp,
		Payload:     log.Payload,
		PrevHash:    log.PrevHash,
		CurrentHash: log.CurrentHash,
	}
}
