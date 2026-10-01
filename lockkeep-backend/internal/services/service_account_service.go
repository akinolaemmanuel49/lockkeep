package services

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"io"
	"strings"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/crypto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	apiKeyPrefix    = "lk_live_"
	apiKeyRawBytes  = 24
	apiKeyPrefixLen = 12
)

type ServiceAccountService struct {
	saRepo         ports.ServiceAccountRepository
	apiKeyRepo     ports.ApiKeyRepository
	envRepo        ports.EnvironmentRepository
	membershipRepo ports.MembershipRepository
	hashSecret     []byte
}

func NewServiceAccountService(
	saRepo ports.ServiceAccountRepository,
	apiKeyRepo ports.ApiKeyRepository,
	envRepo ports.EnvironmentRepository,
	membershipRepo ports.MembershipRepository,
	hashSecret string,
) *ServiceAccountService {
	return &ServiceAccountService{
		saRepo:         saRepo,
		apiKeyRepo:     apiKeyRepo,
		envRepo:        envRepo,
		membershipRepo: membershipRepo,
		hashSecret:     []byte(hashSecret),
	}
}

func (s *ServiceAccountService) requireMember(ctx context.Context, userID, orgID bson.ObjectID, perm domain.Permission) error {
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

func (s *ServiceAccountService) CreateServiceAccount(ctx context.Context, userID, orgID bson.ObjectID, name string, description *string) (*domain.ServiceAccount, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, err
	}
	account := &domain.ServiceAccount{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.saRepo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *ServiceAccountService) ListServiceAccounts(ctx context.Context, orgID bson.ObjectID) ([]domain.ServiceAccount, error) {
	return s.saRepo.FindByOrganization(ctx, orgID)
}

func (s *ServiceAccountService) DeleteServiceAccount(ctx context.Context, userID, orgID, serviceAccountID bson.ObjectID) error {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return err
	}
	account, err := s.saRepo.FindByID(ctx, serviceAccountID)
	if err != nil {
		return err
	}
	if account == nil || account.OrganizationID != orgID {
		return ErrServiceAccountNotFound
	}

	now := time.Now()
	account.DisabledAt = &now
	account.UpdatedAt = now
	if err := s.saRepo.Update(ctx, serviceAccountID, *account); err != nil {
		return err
	}

	keys, err := s.apiKeyRepo.FindByServiceAccount(ctx, serviceAccountID)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if key.RevokedAt != nil {
			continue
		}
		key.RevokedAt = &now
		key.UpdatedAt = now
		if err := s.apiKeyRepo.Update(ctx, key.ID, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceAccountService) MintApiKey(ctx context.Context, userID, orgID, serviceAccountID, environmentID bson.ObjectID, scopes []domain.ApiKeyScope, expiresAt *time.Time) (*domain.ApiKey, string, error) {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return nil, "", err
	}
	if len(scopes) == 0 {
		return nil, "", ErrApiKeyScopesInvalid
	}
	for _, scope := range scopes {
		if scope != domain.ApiKeyScopeRead && scope != domain.ApiKeyScopeWrite {
			return nil, "", ErrApiKeyScopesInvalid
		}
	}

	account, err := s.saRepo.FindByID(ctx, serviceAccountID)
	if err != nil {
		return nil, "", err
	}
	if account == nil || account.OrganizationID != orgID {
		return nil, "", ErrServiceAccountNotFound
	}

	env, err := s.envRepo.FindByID(ctx, environmentID)
	if err != nil {
		return nil, "", err
	}
	if env == nil || env.OrganizationID != orgID {
		return nil, "", ErrEnvironmentNotFound
	}

	raw, prefix, err := s.newRawKey()
	if err != nil {
		return nil, "", err
	}

	key := &domain.ApiKey{
		ID:               bson.NewObjectID(),
		ServiceAccountID: serviceAccountID,
		EnvironmentID:    environmentID,
		KeyPrefix:        prefix,
		KeyHash:          crypto.HMACSHA256(s.hashSecret, []byte(raw)),
		Scopes:           scopes,
		CreatedBy:        userID,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := s.apiKeyRepo.Create(ctx, key); err != nil {
		return nil, "", err
	}
	return key, raw, nil
}

func (s *ServiceAccountService) ListApiKeys(ctx context.Context, orgID, serviceAccountID bson.ObjectID) ([]domain.ApiKey, error) {
	account, err := s.saRepo.FindByID(ctx, serviceAccountID)
	if err != nil {
		return nil, err
	}
	if account == nil || account.OrganizationID != orgID {
		return nil, ErrServiceAccountNotFound
	}
	return s.apiKeyRepo.FindByServiceAccount(ctx, serviceAccountID)
}

func (s *ServiceAccountService) RevokeApiKey(ctx context.Context, userID, orgID, keyID bson.ObjectID) error {
	if err := s.requireMember(ctx, userID, orgID, domain.PermSecretWrite); err != nil {
		return err
	}
	key, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return err
	}
	account, err := s.saRepo.FindByID(ctx, key.ServiceAccountID)
	if err != nil {
		return err
	}
	if account == nil || account.OrganizationID != orgID {
		return ErrApiKeyNotFound
	}

	now := time.Now()
	key.RevokedAt = &now
	key.UpdatedAt = now
	return s.apiKeyRepo.Update(ctx, keyID, *key)
}

func (s *ServiceAccountService) VerifyApiKey(ctx context.Context, rawKey string) (*domain.ApiKey, *domain.ServiceAccount, error) {
	rawKey = strings.TrimSpace(rawKey)
	if !strings.HasPrefix(rawKey, apiKeyPrefix) {
		return nil, nil, ErrApiKeyNotFound
	}
	hexPart := strings.TrimPrefix(rawKey, apiKeyPrefix)
	if len(hexPart) < apiKeyPrefixLen {
		return nil, nil, ErrApiKeyNotFound
	}

	key, err := s.apiKeyRepo.FindByPrefix(ctx, apiKeyPrefix+hexPart[:apiKeyPrefixLen])
	if err != nil {
		return nil, nil, err
	}
	if key == nil {
		return nil, nil, ErrApiKeyNotFound
	}

	if !crypto.ConstantTimeEqual(key.KeyHash, crypto.HMACSHA256(s.hashSecret, []byte(rawKey))) {
		return nil, nil, ErrApiKeyNotFound
	}
	if key.RevokedAt != nil {
		return nil, nil, ErrApiKeyRevoked
	}
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, nil, ErrApiKeyExpired
	}

	account, err := s.saRepo.FindByID(ctx, key.ServiceAccountID)
	if err != nil {
		return nil, nil, err
	}
	if account == nil || account.DisabledAt != nil {
		return nil, nil, ErrApiKeyDisabled
	}

	now := time.Now()
	key.LastUsedAt = &now
	_ = s.apiKeyRepo.Update(ctx, key.ID, *key)

	return key, account, nil
}

func (s *ServiceAccountService) newRawKey() (raw, prefix string, err error) {
	buf := make([]byte, apiKeyRawBytes)
	if _, err := io.ReadFull(cryptorand.Reader, buf); err != nil {
		return "", "", err
	}
	raw = apiKeyPrefix + hex.EncodeToString(buf)
	prefix = apiKeyPrefix + hex.EncodeToString(buf[:6])
	return raw, prefix, nil
}