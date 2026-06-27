package main

import (
	"context"
	"log"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/service"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	// Connect to MongoDB
	ctx := context.Background()
	var client *mongo.Client
	const maxRetries = 30

	clientOpts := options.Client().ApplyURI(cfg.MongoURI).SetServerSelectionTimeout(5 * time.Second)

	for i := 0; i < maxRetries; i++ {
		client, err = mongo.Connect(clientOpts)

		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

			err = client.Ping(pingCtx, readpref.Primary())
			cancel()

			if err == nil {
				break
			}

			_ = client.Disconnect(ctx)
		}

		log.Printf("Waiting for MongoDB (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("unable to connect to MongoDB:", err)
	}

	defer client.Disconnect(ctx)

	log.Println("Connected to MongoDB")

	db := client.Database("lockkeep")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	vaultRepo := repository.NewVaultRepository(db)
	cryptoPolicyRepo := repository.NewCryptoPolicyRepository(db)

	// Ensure indexes
	if err := userRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}
	if err := vaultRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}

	// JWT
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	// Services — NEW LAYER between repos and handlers
	authService := service.NewAuthService(cfg, userRepo, cryptoPolicyRepo, jwtManager)
	vaultService := service.NewVaultService(client, userRepo, cryptoPolicyRepo, vaultRepo)
	cryptoPolicyService := service.NewCryptoPolicyService(cryptoPolicyRepo)

	// Handlers — now depend on services, not repos directly
	authHandler := handlers.NewAuthHandler(authService)
	vaultHandler := handlers.NewVaultHandler(vaultService)
	cryptoPolicyHandler := handlers.NewCryptoPolicyHandler(cryptoPolicyService)

	authMiddleware := middleware.Auth(jwtManager)
	tenantMiddleware := middleware.TenantIsolation()

	// Router
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	api := r.Group("/api/v1")
	{
		api.POST("/auth/oauth", authHandler.OAuth)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)
		api.POST("/auth/logout", authHandler.Logout)
	}

	// Protected routes
	auth := api.Group("/auth")
	auth.Use(authMiddleware, tenantMiddleware)
	{
		auth.POST("/update/email", authHandler.UpdateEmail)
		auth.POST("/update/password", authHandler.UpdateAccountPassword)
		auth.POST("/vault/create", authHandler.SetVerificationHash)
		auth.PUT("/vault/update", authHandler.SetVerificationHash)
		auth.GET("/vault/kdfparams", authHandler.GetKDFParams)
	}

	vault := api.Group("/vault")
	vault.Use(authMiddleware, tenantMiddleware)
	{
		vault.POST("/verify", vaultHandler.VerifyVaultPassword)
		vault.GET("/items", vaultHandler.GetVaultItems)
		vault.POST("/item", vaultHandler.CreateVaultItem)
		vault.PUT("/item/:id", vaultHandler.UpdateVaultItem)
		vault.DELETE("/item/:id", vaultHandler.DeleteVaultItem)
		vault.POST("/migrate", vaultHandler.MigrateVault)
	}

	// Crypto policy routes (admin-only in production)
	policy := api.Group("/crypto-policy")
	policy.Use(authMiddleware, tenantMiddleware)
	{
		policy.GET("/current", cryptoPolicyHandler.GetCurrentPolicy)
		policy.POST("/current", cryptoPolicyHandler.SetCurrentPolicy)
	}

	port := cfg.Port
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
