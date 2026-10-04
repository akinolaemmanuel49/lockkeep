package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/crypto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EnvSecretService struct {
	secretRepo     ports.EnvSecretRepository
	envRepo        ports.EnvironmentRepository
	keyRepo        ports.KeyManagementRepository
	auditRepo      ports.AuditRepository
	membershipRepo ports.MembershipRepository
	keyProvider    crypto.KeyProvider
}

func NewEnvSecretService(
	secretRepo ports.EnvSecretRepository,
	envRepo ports.EnvironmentRepository,
	keyRepo ports.KeyManagementRepository,
	auditRepo ports.AuditRepository,
	membershipRepo ports.MembershipRepository,
	keyProvider crypto.KeyProvider,
) *EnvSecretService {
	return &EnvSecretService{
		secretRepo:     secretRepo,
		envRepo:        envRepo,
		keyRepo:        keyRepo,
		auditRepo:      auditRepo,
		membershipRepo: membershipRepo,
		keyProvider:    keyProvider,
	}
}

func (s *EnvSecretService) requireMember(ctx context.Context, userID, orgID bson.ObjectID, perm domain.Permission) error {
	membership, err := s.membershipRepo.FindByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrEntityHasNoRoles
	}
	if !utils.HasPermission(membership.RoleID, perm) {
		return ErrEntityHasInvalidRoles
	}
	return nil
}

func (s *EnvSecretService) loadEnvDEK(ctx context.Context, orgID, envID bson.ObjectID) (*domain.Environment, []byte, error) {
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return nil, nil, err
	}
	if env == nil || env.OrganizationID != orgID {
		return nil, nil, ErrEnvironmentNotFound
	}
	wrapped, err := s.keyRepo.FindByScope(ctx, crypto.ScopeEnvironment, envID)
	if err != nil {
		return nil, nil, err
	}
	if wrapped == nil {
		return nil, nil, ErrEnvironmentKeyNotFound
	}
	dek, err := s.keyProvider.UnwrapDataKey(ctx, wrapped)
	if err != nil {
		return nil, nil, err
	}
	return env, dek, nil
}

func (s *EnvSecretService) logAudit(ctx context.Context, userID, orgID, resourceID bson.ObjectID, action domain.AuditAction, success bool) {
	_ = s.auditRepo.Log(ctx, &domain.AuditEvent{
		ID:             bson.NewObjectID(),
		UserID:         userID,
		ResourceType:   domain.ResourceTypeSharedSecret,
		ResourceID:     resourceID,
		OrganizationID: &orgID,
		Action:         action,
		Success:        success,
		Decrypted:      true,
		Timestamp:      time.Now(),
	})
}

func validKind(kind domain.SecretKind) bool {
	switch kind {
	case domain.SecretKindEnv, domain.SecretKindNote, domain.SecretKindFile, domain.SecretKindJSON:
		return true
	}
	return false
}

func (s *EnvSecretService) UpsertSecret(ctx context.Context, actor, orgID, envID bson.ObjectID, secret dto.SecretInput) (*dto.EnvSecretDTO, error) {
	if err := s.requireMember(ctx, actor, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}
	if !validKind(secret.Kind) {
		return nil, ErrInvalidSecretKind
	}

	_, dek, err := s.loadEnvDEK(ctx, orgID, envID)
	if err != nil {
		return nil, err
	}

	existing, err := s.secretRepo.FindByKey(ctx, envID, secret.Key)
	if err != nil {
		return nil, err
	}

	version := 1
	secretID := bson.NewObjectID()
	if existing != nil {
		version = existing.Version + 1
		secretID = existing.ID
	}

	ciphertext, nonce, err := crypto.Seal(dek, []byte(secret.Value), crypto.EnvSecretAAD(envID, secret.Key, version))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var updatedBy *bson.ObjectID
	if existing != nil {
		updatedBy = &actor
	}

	record := &domain.EnvSecret{
		ID:             secretID,
		OrganizationID: orgID,
		EnvironmentID:  envID,
		Key:            secret.Key,
		Kind:           secret.Kind,
		Version:        version,
		EncryptedData:  ciphertext,
		Nonce:          nonce,
		CreatedBy:      actor,
		CreatedAt:      now,
		UpdatedBy:      updatedBy,
		UpdatedAt:      now,
	}

	if existing == nil {
		if err := s.secretRepo.Create(ctx, record); err != nil {
			return nil, err
		}
	} else {
		// Preserve the original creation timestamps.
		record.CreatedAt = existing.CreatedAt
		if err := s.secretRepo.Update(ctx, secretID, *record); err != nil {
			return nil, err
		}
	}

	s.logAudit(ctx, actor, orgID, envID, domain.AuditActionWrite, true)
	return s.toDTO(record).WithValue(secret.Value), nil
}

func (s *EnvSecretService) ListSecrets(ctx context.Context, actor, envID bson.ObjectID) ([]dto.EnvSecretDTO, error) {
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	wrapped, err := s.keyRepo.FindByScope(ctx, crypto.ScopeEnvironment, envID)
	if err != nil {
		return nil, err
	}
	if wrapped == nil {
		return nil, ErrEnvironmentKeyNotFound
	}
	dek, err := s.keyProvider.UnwrapDataKey(ctx, wrapped)
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretRepo.FindByEnvironment(ctx, envID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.EnvSecretDTO, 0, len(secrets))
	for i := range secrets {
		plaintext, err := crypto.Open(dek, secrets[i].EncryptedData, secrets[i].Nonce, crypto.EnvSecretAAD(envID, secrets[i].Key, secrets[i].Version))
		if err != nil {
			return nil, err
		}
		out = append(out, *s.toDTO(&secrets[i]).WithValue(string(plaintext)))
	}
	return out, nil
}

func (s *EnvSecretService) GetSecret(ctx context.Context, actor, envID bson.ObjectID, key string) (*dto.EnvSecretDTO, error) {
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	wrapped, err := s.keyRepo.FindByScope(ctx, crypto.ScopeEnvironment, envID)
	if err != nil {
		return nil, err
	}
	if wrapped == nil {
		return nil, ErrEnvironmentKeyNotFound
	}
	dek, err := s.keyProvider.UnwrapDataKey(ctx, wrapped)
	if err != nil {
		return nil, err
	}

	secret, err := s.secretRepo.FindByKey(ctx, envID, key)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, ErrSecretNotFound
	}

	plaintext, err := crypto.Open(dek, secret.EncryptedData, secret.Nonce, crypto.EnvSecretAAD(envID, secret.Key, secret.Version))
	if err != nil {
		return nil, err
	}
	return s.toDTO(secret).WithValue(string(plaintext)), nil
}

func (s *EnvSecretService) DeleteSecret(ctx context.Context, actor, orgID, envID bson.ObjectID, key string) error {
	if err := s.requireMember(ctx, actor, orgID, domain.PermSecretWrite); err != nil {
		return err
	}
	secret, err := s.secretRepo.FindByKey(ctx, envID, key)
	if err != nil {
		return err
	}
	if secret == nil {
		return ErrSecretNotFound
	}
	if err := s.secretRepo.Delete(ctx, secret.ID); err != nil {
		return err
	}
	s.logAudit(ctx, actor, orgID, envID, domain.AuditActionDelete, true)
	return nil
}

// ExportEnvelope re-wraps the environment DEK under a key derived from the
// caller's bearer secret and returns all secret ciphertexts. Invoked on the
// machine path where the caller has already been authenticated by API key, so
// it deliberately makes no membership assumptions.
func (s *EnvSecretService) ExportEnvelope(ctx context.Context, actorID, envID bson.ObjectID, bearerSecret []byte) (*dto.ExportEnvelope, error) {
	env, err := s.envRepo.FindByID(ctx, envID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	wrapped, err := s.keyRepo.FindByScope(ctx, crypto.ScopeEnvironment, envID)
	if err != nil {
		return nil, err
	}
	if wrapped == nil {
		return nil, ErrEnvironmentKeyNotFound
	}
	dek, err := s.keyProvider.UnwrapDataKey(ctx, wrapped)
	if err != nil {
		return nil, err
	}

	// Client wrapping key: HKDF(bearerSecret, salt=envID hex, info=ExportInfo).
	clientKEK, err := crypto.DeriveKey(bearerSecret, crypto.EnvWrapAAD(envID), crypto.ExportInfo, 32)
	if err != nil {
		return nil, err
	}
	wrappedDEKCiphertext, wrappedDEKNonce, err := crypto.Seal(clientKEK, dek, crypto.EnvWrapAAD(envID))
	if err != nil {
		return nil, err
	}

	secrets, err := s.secretRepo.FindByEnvironment(ctx, envID)
	if err != nil {
		return nil, err
	}

	envelope := &dto.ExportEnvelope{
		EnvironmentID: envID.Hex(),
		KDF: dto.KDFSpec{
			Algorithm: "hkdf-sha256",
			Salt:      envID.Hex(),
			Info:      crypto.ExportInfo,
			Length:    32,
		},
		WrappedDEK: dto.WrappedKey{
			KeyID:      wrapped.ID.Hex(),
			Algorithm:  "aes-256-gcm-v1",
			Ciphertext: wrappedDEKCiphertext,
			Nonce:      wrappedDEKNonce,
			AAD:        envID.Hex(),
		},
		Secrets: make([]dto.ExportSecret, 0, len(secrets)),
	}
	for _, secret := range secrets {
		envelope.Secrets = append(envelope.Secrets, dto.ExportSecret{
			Key:        secret.Key,
			Kind:       secret.Kind,
			Version:    secret.Version,
			Ciphertext: secret.EncryptedData,
			Nonce:      secret.Nonce,
		})
	}

	s.logAudit(ctx, actorID, env.OrganizationID, envID, domain.AuditActionExport, true)
	return envelope, nil
}

func (s *EnvSecretService) toDTO(secret *domain.EnvSecret) *dto.EnvSecretDTO {
	return &dto.EnvSecretDTO{
		Key:       secret.Key,
		Kind:      secret.Kind,
		Version:   secret.Version,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}
}