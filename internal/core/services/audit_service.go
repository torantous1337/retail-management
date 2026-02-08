package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// AuditService implements the audit logging service with blockchain-like hashing.
type AuditService struct {
	auditRepo ports.AuditLogRepository
	prevHash  *string // optional override for transactional use
	lastHash  string  // tracks last computed hash
}

// NewAuditService creates a new audit service instance.
func NewAuditService(auditRepo ports.AuditLogRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// SetPrevHash sets an explicit previous hash for sequential audit logging
// within a transaction, avoiding the read-compute-write race condition.
func (s *AuditService) SetPrevHash(hash string) {
	s.prevHash = &hash
}

// LastHash returns the hash of the most recently created audit log entry.
func (s *AuditService) LastHash() string {
	return s.lastHash
}

// LogAction creates an audit log entry with tamper-proof hashing.
func (s *AuditService) LogAction(ctx context.Context, action, userID string, payload map[string]interface{}) error {
	var prevHash string
	if s.prevHash != nil {
		prevHash = *s.prevHash
	} else {
		// Get the last log entry to get its hash
		lastLog, err := s.auditRepo.GetLastLog(ctx)
		if err == nil && lastLog != nil {
			prevHash = lastLog.CurrentHash
		}
	}

	// Create the new log entry
	now := time.Now()
	log := &domain.AuditLog{
		Action:    action,
		UserID:    userID,
		Timestamp: now,
		Payload:   payload,
		PrevHash:  prevHash,
	}

	// Calculate the hash: SHA256(payload + timestamp + prev_hash)
	log.CurrentHash = s.calculateHash(payload, now, prevHash)

	// Store the log entry
	s.lastHash = log.CurrentHash
	return s.auditRepo.Create(ctx, log)
}

// VerifyAuditChain verifies the integrity of the audit log chain.
func (s *AuditService) VerifyAuditChain(ctx context.Context) (bool, error) {
	return s.auditRepo.VerifyChain(ctx)
}

// GetAuditLogs retrieves audit logs with pagination.
func (s *AuditService) GetAuditLogs(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditRepo.List(ctx, limit, offset)
}

// calculateHash computes SHA256(payload + timestamp + prev_hash).
func (s *AuditService) calculateHash(payload map[string]interface{}, timestamp time.Time, prevHash string) string {
	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		payloadBytes = []byte("{}")
	}

	// Create the hash input: payload + timestamp + prev_hash
	hashInput := fmt.Sprintf("%s%s%s", string(payloadBytes), timestamp.Format(time.RFC3339Nano), prevHash)

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(hashInput))
	return fmt.Sprintf("%x", hash)
}
