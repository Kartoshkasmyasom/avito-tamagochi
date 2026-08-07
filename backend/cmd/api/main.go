package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/accelolabs/avito-tamagochi/backend/internal/app/registration"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/handler"
	authmiddleware "github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/middleware"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/leaderboard"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/rewards"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/summary"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/tasks"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	authRepo := repository.NewPgRepository(db)
	authService := service.NewAuthService(authRepo)
	petService := pet.NewService(db)
	registrationService := registration.NewService(authService, petService)
	authHandler := handler.NewAuthHandler(authService, registrationService)
	tasksService := tasks.NewService(db)
	rewardsService := rewards.NewService(db)
	leaderboardService := leaderboard.NewService(db)
	summaryService := summary.NewService(db)
	gameHandler := game.NewHandler(petService, tasksService, rewardsService, leaderboardService, summaryService)
	requireSession := authmiddleware.RequireSession(authRepo)

	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		protected := v1.Group("", requireSession)
		{
			protected.GET("/pet", gameHandler.GetPet)
			protected.POST("/pet/actions", gameHandler.PetAction)
			protected.GET("/tasks", gameHandler.ListTasks)
			protected.POST("/demo/activities", gameHandler.Activity)
			protected.GET("/rewards", gameHandler.ListRewards)
			protected.POST("/rewards/:rewardId/claim", gameHandler.ClaimReward)
			protected.GET("/leaderboard", gameHandler.Leaderboard)
			protected.GET("/summary/daily", gameHandler.Summary)
		}
	}

	log.Printf("server started on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
