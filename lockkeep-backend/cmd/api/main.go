package main

import (
	"context"
	"log"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/crypto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repositories"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	ctx := context.Background()

	// Connect to MongoDB
	db, err := mongo.Connect(ctx, cfg.MongoURI, "lockkeep")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Disconnect(ctx)
	log.Println("Connected to MongoDB")

	rawDB := db.Raw()
	uow := mongo.NewUnitOfWork(rawDB.Client())

	// Repositories
	userRepo := repositories.NewUserRepository(rawDB)
	identityRepo := repositories.NewIdentityRepository(rawDB)
	orgRepo := repositories.NewOrganizationRepository(rawDB)
	membershipRepo := repositories.NewMembershipRepository(rawDB)
	teamRepo := repositories.NewTeamRepository(rawDB)
	vaultRepo := repositories.NewVaultRepository(rawDB)
	vaultItemRepo := repositories.NewVaultItemRepository(rawDB)
	// sharedSecretRepo := repository.NewSharedSecretRepository(rawDB)
	auditRepo := repositories.NewAuditEventRepository(rawDB)
	cryptoPolicyRepo := repositories.NewCryptoPolicyRepository(rawDB)

	// Shared secrets (workspace) repositories
	appRepo := repositories.NewApplicationRepository(rawDB)
	envRepo := repositories.NewEnvironmentRepository(rawDB)
	envSecretRepo := repositories.NewEnvSecretRepository(rawDB)
	serviceAccountRepo := repositories.NewServiceAccountRepository(rawDB)
	apiKeyRepo := repositories.NewApiKeyRepository(rawDB)
	keyMgmtRepo := repositories.NewKeyManagementRepository(rawDB)

	// Ensure indexes
	indexRepos := []interface{ EnsureIndexes(context.Context) error }{
		userRepo,
		identityRepo,
		orgRepo,
		membershipRepo,
		vaultRepo,
		vaultItemRepo,
		auditRepo,
		appRepo,
		envRepo,
		envSecretRepo,
		serviceAccountRepo,
		apiKeyRepo,
		keyMgmtRepo,
	}
	for _, repo := range indexRepos {
		if err := repo.EnsureIndexes(ctx); err != nil {
			log.Fatalf("ensure indexes: %v", err)
		}
	}

	// JWT
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	// Services
	authService := services.NewAuthService(cfg, userRepo, identityRepo, membershipRepo, uow, jwtManager)
	userService := services.NewUserService(cfg, userRepo)
	orgService := services.NewOrganizationService(orgRepo, membershipRepo)
	teamService := services.NewTeamService(teamRepo, membershipRepo)
	vaultService := services.NewVaultService(cfg, vaultRepo, userRepo)
	vaultItemService := services.NewVaultItemService(vaultItemRepo)
	// sharedSecretService := service.NewSharedSecretService(sharedSecretRepo, auditRepo)
	// auditService := service.NewAuditService(auditRepo)
	cryptoPolicyService := services.NewCryptoPolicyService(cryptoPolicyRepo)

	// Shared secrets (workspace) services
	keyProvider, err := crypto.NewLocalKeyProvider([]byte(cfg.KeyWrappingKey))
	if err != nil {
		log.Fatalf("initialise key provider: %v", err)
	}
	if cfg.KeyWrappingKey == "" {
		log.Println("WARNING: LOCKKEEP_KEK is not set; using an ephemeral key-wrapping key (wrapped keys become unreadable on restart)")
	}
	environmentService := services.NewEnvironmentService(appRepo, envRepo, envSecretRepo, keyMgmtRepo, membershipRepo, keyProvider)
	envSecretService := services.NewEnvSecretService(envSecretRepo, envRepo, keyMgmtRepo, auditRepo, membershipRepo, keyProvider)
	serviceAccountService := services.NewServiceAccountService(serviceAccountRepo, apiKeyRepo, envRepo, membershipRepo, cfg.APIKeyHashSecret)

	// Seed the default crypto policy so clients always have parameters to
	// derive keys against (no-op when a policy already exists).
	if err := seedDefaultCryptoPolicy(ctx, cryptoPolicyService); err != nil {
		log.Fatalf("seed default crypto policy: %v", err)
	}

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	teamHandler := handlers.NewTeamHandler(teamService)
	vaultHandler := handlers.NewVaultHandler(vaultService, vaultItemService)
	cryptoPolicyHandler := handlers.NewCryptoPolicyHandler(cryptoPolicyService)
	applicationHandler := handlers.NewApplicationHandler(environmentService)
	environmentHandler := handlers.NewEnvironmentHandler(environmentService)
	envSecretHandler := handlers.NewEnvSecretHandler(environmentService, envSecretService)
	serviceAccountHandler := handlers.NewServiceAccountHandler(environmentService, serviceAccountService)
	machineHandler := handlers.NewMachineHandler(envSecretService)

	// Middleware
	authMiddleware := middleware.Auth(jwtManager)
	requireOrgAccess := middleware.RequireOrgAccess(membershipRepo, orgRepo)
	machineAuth := middleware.MachineAuth(serviceAccountService)

	// Router
	r := gin.Default()

	// CORS
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}
	r.Use(corsMiddleware(allowedOrigins))

	// API v2
	v2 := r.Group("/api/v2")

	// Public routes (register, login, refresh, oauth)
	v2.POST("/auth/register", authHandler.Register)
	v2.POST("/auth/login", authHandler.Login)
	v2.POST("/auth/refresh", authHandler.Refresh)
	v2.POST("/auth/oauth", authHandler.OAuth)

	// User routes (profile, requires auth). Personal vault lives under /me/vault.
	me := v2.Group("/me")
	me.Use(authMiddleware)
	{
		me.GET("", userHandler.GetMe)
		me.PATCH("", userHandler.UpdateProfile)

		me.POST("/email", authHandler.UpdateEmail)
		me.POST("/password", authHandler.UpdateAccountPassword)

		me.GET("/vault", vaultHandler.GetVault)
		me.PATCH("/vault", vaultHandler.UpdateVault)
		me.POST("/vault/setup", authHandler.SetVerificationHash)
		me.PUT("/vault/setup", authHandler.SetVerificationHash)
		me.GET("/vault/kdfparams", authHandler.GetKDFParams)
		me.POST("/vault/verify", vaultHandler.VerifyVaultPassword)

		me.GET("/vault/items", vaultHandler.ListVaultItems)
		me.POST("/vault/items", vaultHandler.CreateVaultItem)
		me.GET("/vault/items/:itemID", vaultHandler.GetVaultItem)
		me.PUT("/vault/items/:itemID", vaultHandler.UpdateVaultItem)
		me.DELETE("/vault/items/:itemID", vaultHandler.DeleteVaultItem)
	}

	orgs := v2.Group("/organizations")
	orgs.Use(authMiddleware)
	{
		orgs.GET("", orgHandler.List)
		orgs.POST("", orgHandler.Create)

		org := orgs.Group("/:orgSlug")
		org.Use(requireOrgAccess)
		{
			org.GET("", orgHandler.Get)
			org.PATCH("", orgHandler.Update)
			org.DELETE("", orgHandler.Delete)
			org.GET("/teams", teamHandler.List)
			org.POST("/teams", middleware.RequirePermission(domain.PermTeamManage), teamHandler.Create)

			team := org.Group("/teams")
			{
				team.GET("/:teamSlug", teamHandler.Get)
				team.DELETE("/:teamID", middleware.RequirePermission(domain.PermTeamManage), teamHandler.Delete)
			}

			// Shared secrets: applications → environments → secrets
			apps := org.Group("/applications")
			{
				apps.GET("", applicationHandler.List)
				apps.POST("", middleware.RequirePermission(domain.PermSecretWrite), applicationHandler.Create)
				apps.GET("/:appSlug", applicationHandler.Get)
				apps.PATCH("/:appSlug", middleware.RequirePermission(domain.PermSecretWrite), applicationHandler.Update)
				apps.DELETE("/:appSlug", middleware.RequirePermission(domain.PermSecretWrite), applicationHandler.Delete)

				envs := apps.Group("/:appSlug/environments")
				{
					envs.GET("", environmentHandler.List)
					envs.POST("", middleware.RequirePermission(domain.PermSecretWrite), environmentHandler.Create)
					envs.GET("/:envSlug", environmentHandler.Get)
					envs.PATCH("/:envSlug", middleware.RequirePermission(domain.PermSecretWrite), environmentHandler.Update)
					envs.DELETE("/:envSlug", middleware.RequirePermission(domain.PermSecretWrite), environmentHandler.Delete)

					secrets := envs.Group("/:envSlug/secrets")
					{
						secrets.GET("", envSecretHandler.List)
						secrets.POST("", middleware.RequirePermission(domain.PermSecretWrite), envSecretHandler.Upsert)
						secrets.GET("/:key", envSecretHandler.Get)
						secrets.PUT("/:key", middleware.RequirePermission(domain.PermSecretWrite), envSecretHandler.Upsert)
						secrets.DELETE("/:key", middleware.RequirePermission(domain.PermSecretWrite), envSecretHandler.Delete)
					}
				}
			}

			// Service accounts + environment-scoped API keys
			accounts := org.Group("/service-accounts", middleware.RequirePermission(domain.PermSecretWrite))
			{
				accounts.GET("", serviceAccountHandler.List)
				accounts.POST("", serviceAccountHandler.Create)
				accounts.DELETE("/:accountID", serviceAccountHandler.Delete)
				accounts.POST("/:accountID/api-keys", serviceAccountHandler.MintApiKey)
				accounts.GET("/:accountID/api-keys", serviceAccountHandler.ListApiKeys)
				accounts.DELETE("/:accountID/api-keys/:keyID", serviceAccountHandler.RevokeApiKey)
			}
		}
	}

	// Machine routes: authenticated by an environment-scoped API key only
	// (no session JWT). The key carries the environment and scope.
	machine := v2.Group("/machine")
	machine.Use(machineAuth)
	{
		machine.GET("/secrets", machineHandler.Export)
	}

	// Crypto policy routes (system-wide KDF parameters). Writes require the
	// system admin role.
	cryptoPolicy := v2.Group("/crypto-policy")
	cryptoPolicy.Use(authMiddleware)
	{
		cryptoPolicy.GET("/current", cryptoPolicyHandler.GetCurrentPolicy)
		cryptoPolicy.POST("/current", cryptoPolicyHandler.SetCurrentPolicy)
	}

	port := cfg.Port
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(allowedOrigins map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// seedDefaultCryptoPolicy writes the initial crypto policy (argon2id defaults)
// on first boot so clients always have KDF parameters to derive keys against.
func seedDefaultCryptoPolicy(ctx context.Context, p ports.CryptoPolicyService) error {
	current, err := p.GetCurrentPolicy(ctx)
	if err != nil {
		return err
	}
	if current != nil {
		log.Printf("Crypto policy already seeded (version %d)", current.Version)
		return nil
	}

	defaultPolicy := &dto.SetCurrentPolicy{}
	defaultPolicy.KDFParams.Algorithm = "argon2id"
	defaultPolicy.KDFParams.Memory = 65536
	defaultPolicy.KDFParams.Iterations = 3
	defaultPolicy.KDFParams.Parallelism = 4

	if err := p.SetCurrentPolicy(ctx, defaultPolicy); err != nil {
		return err
	}

	log.Println("Seeded default crypto policy (argon2id v1)")
	return nil
}
