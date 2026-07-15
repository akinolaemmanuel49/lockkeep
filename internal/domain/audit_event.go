package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditAction string

const (
	AuditActionRead   AuditAction = "read"
	AuditActionWrite  AuditAction = "write"
	AuditActionDelete AuditAction = "delete"
	AuditActionShare  AuditAction = "share"
	AuditActionRotate AuditAction = "rotate"
	AuditActionAccess AuditAction = "access" // Generic access event
	AuditActionExport AuditAction = "export"
	AuditActionCopy   AuditAction = "copy"
)

type AuditResourceType string

const (
	ResourceTypeSharedSecret AuditResourceType = "shared_secret"
	ResourceTypeOrganization AuditResourceType = "organization"
	ResourceTypeTeam         AuditResourceType = "team"
	ResourceTypeMembership   AuditResourceType = "membership"
	ResourceTypeVaultItem    AuditResourceType = "vault_item" // For personal vault audit (client-reported)
)

// AuditEvent — immutable log of who accessed what, when, and how
type AuditEvent struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id"`

	// Actor
	UserID bson.ObjectID `bson:"user_id" json:"userId"`

	// Target
	ResourceType AuditResourceType `bson:"resource_type" json:"resourceType"`
	ResourceID   bson.ObjectID     `bson:"resource_id" json:"resourceId"`

	// Context
	OrganizationID *bson.ObjectID `bson:"organization_id,omitempty" json:"organizationId,omitempty"`
	TeamID         *bson.ObjectID `bson:"team_id,omitempty" json:"teamId,omitempty"`

	// Action details
	Action      AuditAction `bson:"action" json:"action"`
	Success     bool        `bson:"success" json:"success"`
	ErrorReason *string     `bson:"error_reason,omitempty" json:"errorReason,omitempty"`

	// Request context
	IPAddress string `bson:"ip_address" json:"ipAddress"`
	UserAgent string `bson:"user_agent" json:"userAgent"`
	RequestID string `bson:"request_id" json:"requestId"`
	SessionID string `bson:"session_id" json:"sessionId"`

	// For secret access: was it decrypted? (always true for shared secrets)
	Decrypted bool `bson:"decrypted" json:"decrypted"`

	// Timestamps
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}
