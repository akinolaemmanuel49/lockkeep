package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/crypto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestKeyProvider(t *testing.T) crypto.KeyProvider {
	t.Helper()
	p, err := crypto.NewLocalKeyProvider([]byte("test-key-wrapping-material"))
	assert.NoError(t, err)
	return p
}

func ownerMembership(orgID, userID bson.ObjectID) *domain.Membership {
	return &domain.Membership{OrganizationID: orgID, UserID: userID, RoleID: domain.RoleOrgOwner}
}

// ─── LocalKeyProvider ───

func TestLocalKeyProvider_RoundTrip(t *testing.T) {
	p := newTestKeyProvider(t)
	ctx := context.Background()
	envID := bson.NewObjectID()

	wrapped, dek, err := p.GenerateDataKey(ctx, "", crypto.ScopeEnvironment, envID)
	assert.NoError(t, err)
	assert.Len(t, dek, 32)
	assert.NotNil(t, wrapped)
	assert.Equal(t, crypto.ScopeEnvironment, wrapped.Scope)
	assert.Equal(t, envID, wrapped.ScopeID)
	assert.Equal(t, crypto.LocalKeyWrappingAlgorithm, wrapped.Algorithm)
	assert.NotEmpty(t, wrapped.Nonce)

	got, err := p.UnwrapDataKey(ctx, wrapped)
	assert.NoError(t, err)
	assert.Equal(t, dek, got)

	// Tamper with the ciphertext and ensure authentication fails.
	wrapped.Ciphertext = append([]byte(nil), wrapped.Ciphertext...)
	wrapped.Ciphertext[0] ^= 0xFF
	_, err = p.UnwrapDataKey(ctx, wrapped)
	assert.ErrorIs(t, err, crypto.ErrTamperedKeyWrap)
}

// ─── EnvironmentService ───

func TestEnvironmentService_CreateEnvironment_ProvisionsDEK(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	envRepo := new(mocks.MockEnvironmentRepository)
	secretRepo := new(mocks.MockEnvSecretRepository)
	keyRepo := new(mocks.MockKeyManagementRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	p := newTestKeyProvider(t)

	svc := services.NewEnvironmentService(appRepo, envRepo, secretRepo, keyRepo, membershipRepo, p)

	userID, orgID, appID := bson.NewObjectID(), bson.NewObjectID(), bson.NewObjectID()

	membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(ownerMembership(orgID, userID), nil)
	appRepo.On("FindByID", mock.Anything, appID).Return(&domain.Application{ID: appID, OrganizationID: orgID}, nil)
	envRepo.On("FindBySlug", mock.Anything, orgID, appID, "dev").Return(nil, nil)
	envRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Environment")).Return(nil).Run(func(args mock.Arguments) {
		env := args.Get(1).(*domain.Environment)
		assert.Equal(t, orgID, env.OrganizationID)
		assert.Equal(t, appID, env.ApplicationID)
		assert.Equal(t, "dev", env.Slug)
	})
	var stored *domain.EncryptedKey
	keyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.EncryptedKey")).Return(nil).Run(func(args mock.Arguments) {
		stored = args.Get(1).(*domain.EncryptedKey)
	})

	env, err := svc.CreateEnvironment(context.Background(), userID, orgID, appID, "Development", "dev", true)

	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, "dev", env.Slug)
	assert.True(t, env.IsProtected)
	assert.NotNil(t, stored)
	assert.Equal(t, crypto.ScopeEnvironment, stored.Scope)
	assert.Equal(t, env.ID, stored.ScopeID)

	// The wrapped DEK must be unwrappable, proving it was generated.
	dek, err := p.UnwrapDataKey(context.Background(), stored)
	assert.NoError(t, err)
	assert.Len(t, dek, 32)

	appRepo.AssertExpectations(t)
	envRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	membershipRepo.AssertExpectations(t)
}

// ─── EnvSecretService ───

func newTestEnvSecretService(t *testing.T) (
	*services.EnvSecretService,
	*mocks.MockEnvSecretRepository,
	*mocks.MockEnvironmentRepository,
	*mocks.MockKeyManagementRepository,
	*mocks.MockAuditRepository,
	*mocks.MockMembershipRepository,
	crypto.KeyProvider,
) {
	t.Helper()
	secretRepo := new(mocks.MockEnvSecretRepository)
	envRepo := new(mocks.MockEnvironmentRepository)
	keyRepo := new(mocks.MockKeyManagementRepository)
	auditRepo := new(mocks.MockAuditRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	p := newTestKeyProvider(t)

	svc := services.NewEnvSecretService(secretRepo, envRepo, keyRepo, auditRepo, membershipRepo, p)
	return svc, secretRepo, envRepo, keyRepo, auditRepo, membershipRepo, p
}

func seedEnvironmentDEK(t *testing.T, p crypto.KeyProvider, envID bson.ObjectID) *domain.EncryptedKey {
	t.Helper()
	wrapped, _, err := p.GenerateDataKey(context.Background(), "", crypto.ScopeEnvironment, envID)
	assert.NoError(t, err)
	return wrapped
}

func TestEnvSecretService_Upsert_And_Get_RoundTrip(t *testing.T) {
	svc, secretRepo, envRepo, keyRepo, auditRepo, membershipRepo, p := newTestEnvSecretService(t)

	userID, orgID := bson.NewObjectID(), bson.NewObjectID()
	envID := bson.NewObjectID()
	wrapped := seedEnvironmentDEK(t, p, envID)

	membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(ownerMembership(orgID, userID), nil)
	envRepo.On("FindByID", mock.Anything, envID).Return(&domain.Environment{ID: envID, OrganizationID: orgID}, nil)
	keyRepo.On("FindByScope", mock.Anything, crypto.ScopeEnvironment, envID).Return(wrapped, nil)
	auditRepo.On("Log", mock.Anything, mock.AnythingOfType("*domain.AuditEvent")).Return(nil)

	secretRepo.On("FindByKey", mock.Anything, envID, "DATABASE_URL").Return(nil, nil).Once()
	var saved *domain.EnvSecret
	secretRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.EnvSecret")).Return(nil).Run(func(args mock.Arguments) {
		saved = args.Get(1).(*domain.EnvSecret)
	})

	secret, err := svc.UpsertSecret(context.Background(), userID, orgID, envID, dto.SecretInput{
		Key:   "DATABASE_URL",
		Kind:  domain.SecretKindEnv,
		Value: "postgres://db/prod",
	})

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "DATABASE_URL", secret.Key)
	assert.Equal(t, "postgres://db/prod", secret.Value)
	assert.Equal(t, 1, secret.Version)
	assert.NotNil(t, saved)
	assert.Equal(t, 1, saved.Version)
	assert.NotEmpty(t, saved.EncryptedData)
	assert.NotEmpty(t, saved.Nonce)

	// Read path decrypts server-side.
	secretRepo.On("FindByKey", mock.Anything, envID, "DATABASE_URL").Return(saved, nil)
	got, err := svc.GetSecret(context.Background(), userID, envID, "DATABASE_URL")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "postgres://db/prod", got.Value)
	assert.Equal(t, 1, got.Version)

	secretRepo.AssertExpectations(t)
	envRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
	membershipRepo.AssertExpectations(t)
}

func TestEnvSecretService_Upsert_BumpsVersion(t *testing.T) {
	svc, secretRepo, envRepo, keyRepo, auditRepo, membershipRepo, p := newTestEnvSecretService(t)

	userID, orgID := bson.NewObjectID(), bson.NewObjectID()
	envID := bson.NewObjectID()
	wrapped := seedEnvironmentDEK(t, p, envID)

	membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(ownerMembership(orgID, userID), nil)
	envRepo.On("FindByID", mock.Anything, envID).Return(&domain.Environment{ID: envID, OrganizationID: orgID}, nil)
	keyRepo.On("FindByScope", mock.Anything, crypto.ScopeEnvironment, envID).Return(wrapped, nil)
	auditRepo.On("Log", mock.Anything, mock.AnythingOfType("*domain.AuditEvent")).Return(nil)

	existing := &domain.EnvSecret{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		EnvironmentID:  envID,
		Key:            "TOKEN",
		Kind:           domain.SecretKindEnv,
		Version:        3,
		CreatedAt:      time.Now().Add(-time.Hour),
		UpdatedAt:      time.Now(),
	}
	secretRepo.On("FindByKey", mock.Anything, envID, "TOKEN").Return(existing, nil)
	var updated *domain.EnvSecret
	secretRepo.On("Update", mock.Anything, existing.ID, mock.AnythingOfType("domain.EnvSecret")).Return(nil).Run(func(args mock.Arguments) {
		got := args.Get(2).(domain.EnvSecret)
		updated = &got
	})

	secret, err := svc.UpsertSecret(context.Background(), userID, orgID, envID, dto.SecretInput{
		Key:   "TOKEN",
		Kind:  domain.SecretKindEnv,
		Value: "new-value",
	})

	assert.NoError(t, err)
	assert.Equal(t, 4, secret.Version)
	assert.Equal(t, "new-value", secret.Value)
	assert.NotNil(t, updated)
	assert.Equal(t, 4, updated.Version)
	assert.Equal(t, existing.CreatedAt, updated.CreatedAt)
	secretRepo.AssertExpectations(t)
}

// ─── Export envelope (machine path + client-side decrypt proof) ───

func TestEnvSecretService_ExportEnvelope_ClientRoundTrip(t *testing.T) {
	svc, secretRepo, envRepo, keyRepo, auditRepo, _, p := newTestEnvSecretService(t)

	orgID := bson.NewObjectID()
	envID := bson.NewObjectID()
	dek := make([]byte, 32)
	for i := range dek {
		dek[i] = byte(i)
	}
	wrapped, err := p.WrapDataKey(context.Background(), "", crypto.ScopeEnvironment, envID, dek)
	assert.NoError(t, err)

	secrets := []domain.EnvSecret{
		{Key: "API_KEY", Kind: domain.SecretKindEnv, Version: 1, EnvironmentID: envID, OrganizationID: orgID},
		{Key: "note", Kind: domain.SecretKindNote, Version: 2, EnvironmentID: envID, OrganizationID: orgID},
	}
	for i := range secrets {
		ct, nonce, err := crypto.Seal(dek, []byte(fmt.Sprintf("plaintext-%d", i)), crypto.EnvSecretAAD(envID, secrets[i].Key, secrets[i].Version))
		assert.NoError(t, err)
		secrets[i].EncryptedData = ct
		secrets[i].Nonce = nonce
	}

	envRepo.On("FindByID", mock.Anything, envID).Return(&domain.Environment{ID: envID, OrganizationID: orgID}, nil)
	keyRepo.On("FindByScope", mock.Anything, crypto.ScopeEnvironment, envID).Return(wrapped, nil)
	secretRepo.On("FindByEnvironment", mock.Anything, envID).Return(secrets, nil)
	auditRepo.On("Log", mock.Anything, mock.AnythingOfType("*domain.AuditEvent")).Return(nil)

	bearerSecret := []byte("lk_live_0123456789abcdef0123456789abcdef0123456789abcdef")
	actorID := bson.NewObjectID()

	envelope, err := svc.ExportEnvelope(context.Background(), actorID, envID, bearerSecret)
	assert.NoError(t, err)
	assert.Equal(t, envID.Hex(), envelope.EnvironmentID)
	assert.Equal(t, "hkdf-sha256", envelope.KDF.Algorithm)
	assert.Equal(t, 32, envelope.KDF.Length)
	assert.Len(t, envelope.Secrets, 2)

	// ── Simulated client: never sees the plaintext DEK ──
	clientKEK, err := crypto.DeriveKey(bearerSecret, []byte(envelope.KDF.Salt), envelope.KDF.Info, envelope.KDF.Length)
	assert.NoError(t, err)
	clientDEK, err := crypto.Open(clientKEK, envelope.WrappedDEK.Ciphertext, envelope.WrappedDEK.Nonce, []byte(envelope.WrappedDEK.AAD))
	assert.NoError(t, err)
	assert.Equal(t, dek, clientDEK)

	for i, s := range envelope.Secrets {
		aad := []byte(fmt.Sprintf("%s|%s|%d", envID.Hex(), s.Key, s.Version))
		plaintext, err := crypto.Open(clientDEK, s.Ciphertext, s.Nonce, aad)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("plaintext-%d", i), string(plaintext))
	}

	secretRepo.AssertExpectations(t)
	envRepo.AssertExpectations(t)
	keyRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

// ─── ServiceAccountService ───

func TestServiceAccountService_MintAndVerify(t *testing.T) {
	saRepo := new(mocks.MockServiceAccountRepository)
	apiKeyRepo := new(mocks.MockApiKeyRepository)
	envRepo := new(mocks.MockEnvironmentRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	svc := services.NewServiceAccountService(saRepo, apiKeyRepo, envRepo, membershipRepo, "test-hash-secret")

	userID, orgID := bson.NewObjectID(), bson.NewObjectID()
	accountID, envID := bson.NewObjectID(), bson.NewObjectID()
	env := &domain.Environment{ID: envID, OrganizationID: orgID}

	membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(ownerMembership(orgID, userID), nil)
	saRepo.On("FindByID", mock.Anything, accountID).Return(&domain.ServiceAccount{ID: accountID, OrganizationID: orgID}, nil)
	envRepo.On("FindByID", mock.Anything, envID).Return(env, nil)

	var stored *domain.ApiKey
	apiKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ApiKey")).Return(nil).Run(func(args mock.Arguments) {
		stored = args.Get(1).(*domain.ApiKey)
	})

	key, raw, err := svc.MintApiKey(context.Background(), userID, orgID, accountID, envID, []domain.ApiKeyScope{domain.ApiKeyScopeRead}, nil)

	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotEmpty(t, raw)
	assert.Equal(t, apiKeyPrefix, raw[:len(apiKeyPrefix)])
	assert.NotEmpty(t, stored.KeyHash)
	assert.Equal(t, stored.KeyPrefix, raw[:len(apiKeyPrefix)+12])

	apiKeyRepo.On("FindByPrefix", mock.Anything, stored.KeyPrefix).Return(stored, nil)
	saRepo.On("FindByID", mock.Anything, accountID).Return(&domain.ServiceAccount{ID: accountID, OrganizationID: orgID}, nil)
	apiKeyRepo.On("Update", mock.Anything, stored.ID, mock.AnythingOfType("domain.ApiKey")).Return(nil)

	verified, account, err := svc.VerifyApiKey(context.Background(), raw)
	assert.NoError(t, err)
	assert.Equal(t, stored.ID, verified.ID)
	assert.Equal(t, accountID, account.ID)

	// Tampered / different key must be rejected without leaking the reason.
	apiKeyRepo.On("FindByPrefix", mock.Anything, stored.KeyPrefix).Return(stored, nil)
	_, _, err = svc.VerifyApiKey(context.Background(), raw[:len(raw)-2]+"zz")
	assert.ErrorIs(t, err, services.ErrApiKeyNotFound)

	// Revoked keys are rejected.
	revokedAt := time.Now()
	stored.RevokedAt = &revokedAt
	apiKeyRepo.On("Update", mock.Anything, stored.ID, mock.AnythingOfType("domain.ApiKey")).Return(nil)
	_, _, err = svc.VerifyApiKey(context.Background(), raw)
	assert.ErrorIs(t, err, services.ErrApiKeyRevoked)

	saRepo.AssertExpectations(t)
	apiKeyRepo.AssertExpectations(t)
	envRepo.AssertExpectations(t)
	membershipRepo.AssertExpectations(t)
}

const apiKeyPrefix = "lk_live_"