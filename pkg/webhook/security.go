package webhook

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	// SignatureHeaderName is the header name for webhook signatures
	SignatureHeaderName = "X-Webhook-Signature-256"
	// TimestampHeaderName is the header name for webhook timestamps
	TimestampHeaderName = "X-Webhook-Timestamp"
	// MaxTimestampAge is the maximum age of a webhook timestamp
	MaxTimestampAge = 5 * time.Minute
)

// SecurityManager handles webhook security operations
type SecurityManager struct{}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{}
}

// GenerateSecret creates a cryptographically secure webhook secret
func (s *SecurityManager) GenerateSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// SignPayload creates an HMAC SHA256 signature for a payload
func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies the signature of a webhook payload
func VerifySignature(secret string, payload []byte, signature string) bool {
	expected := SignPayload(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// SignPayloadWithTimestamp creates a signature including timestamp for replay protection
func (s *SecurityManager) SignPayloadWithTimestamp(secret string, payload []byte, timestamp int64) string {
	// Create payload with timestamp
	timestampStr := strconv.FormatInt(timestamp, 10)
	signedPayload := append([]byte(timestampStr+"."), payload...)
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(signedPayload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignatureWithTimestamp verifies signature with timestamp for replay protection
func (s *SecurityManager) VerifySignatureWithTimestamp(secret string, payload []byte, signature string, timestamp int64) bool {
	// Check timestamp age
	now := time.Now().Unix()
	if now-timestamp > int64(MaxTimestampAge.Seconds()) {
		return false
	}
	
	expected := s.SignPayloadWithTimestamp(secret, payload, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// RotateSecret generates a new secret for a webhook
func (s *SecurityManager) RotateSecret() (string, error) {
	return s.GenerateSecret()
}
