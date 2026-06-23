package main

import (
	"context"
	"log"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database("lockkeep")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	credRepo := repository.NewCredentialRepository(db)

	// Ensure indexes
	if err := userRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}
	if err := credRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}

	// JWT
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	// Handlers
	authHandler := handlers.NewAuthHandler(cfg, userRepo, jwtManager)

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
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	protected.Use(middleware.TenantIsolation())
	{
		// TODO: vault and credential handlers
	}

	port := cfg.Port
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
