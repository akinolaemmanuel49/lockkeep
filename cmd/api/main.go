package main

import (
	"context"
	"log"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/service"
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
	userRepo := repository.NewUserRepository(rawDB)
	identityRepo := repository.NewIdentityRepository(rawDB)
	orgRepo := repository.NewOrganizationRepository(rawDB)
	membershipRepo := repository.NewMembershipRepository(rawDB)
	// teamRepo := repository.NewTeamRepository(rawDB)
	// vaultItemRepo := repository.NewVaultItemRepository(rawDB)
	// sharedSecretRepo := repository.NewSharedSecretRepository(rawDB)
	// auditRepo := repository.NewAuditRepository(rawDB)
	// cryptoPolicyRepo := repository.NewCryptoPolicyRepository(rawDB)

	// Ensure indexes
	indexRepos := []interface{ EnsureIndexes(context.Context) error }{
		userRepo,
		identityRepo,
		orgRepo,
		membershipRepo,
	}
	for _, repo := range indexRepos {
		if err := repo.EnsureIndexes(ctx); err != nil {
			log.Fatalf("ensure indexes: %v", err)
		}
	}

	// JWT
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	// Services
	authService := service.NewAuthService(cfg, userRepo, identityRepo, membershipRepo, uow, jwtManager)
	// orgService := service.NewOrganizationService(orgRepo, membershipRepo)
	// teamService := service.NewTeamService(teamRepo, membershipRepo)
	// vaultService := service.NewVaultService(vaultItemRepo)
	// sharedSecretService := service.NewSharedSecretService(sharedSecretRepo, auditRepo)
	// auditService := service.NewAuditService(auditRepo)
	// cryptoPolicyService := service.NewCryptoPolicyService(cryptoPolicyRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	// orgHandler := handlers.NewOrganizationHandler(orgService)
	// ...

	// Middleware
	authMiddleware := middleware.Auth(jwtManager)
	// requireOrgAccess := middleware.RequireOrgAccess(membershipRepo, orgRepo)

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

	// Public
	v2.POST("/auth/register", authHandler.Register)
	v2.POST("/auth/login", authHandler.Login)
	v2.POST("/auth/oauth", authHandler.OAuth)
	v2.POST("/auth/refresh", authHandler.Refresh)

	// Protected
	me := v2.Group("/me")
	me.Use(authMiddleware)
	{
		me.GET("", authHandler.GetMe)
		// me.GET("/vault", vaultHandler.GetVault)
		// me.POST("/vault/items", vaultHandler.CreateItem)
	}

	// Organizations
	// orgs := v2.Group("/organizations")
	// orgs.Use(authMiddleware)
	// {
	//     orgs.GET("", orgHandler.List)
	//     orgs.POST("", orgHandler.Create)
	//     org := orgs.Group("/:orgSlug")
	//     org.Use(requireOrgAccess)
	//     {
	//         org.GET("", orgHandler.Get)
	//         org.GET("/members", orgHandler.ListMembers)
	//         org.POST("/teams", teamHandler.Create)
	//     }
	// }

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
