package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/db"
	"github.com/nicolas/finanzas/backend/internal/middleware"
	"github.com/nicolas/finanzas/backend/internal/server"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	gormDB, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Migrate(cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	repo := auth.NewUserRepository(gormDB)
	sessions := auth.NewSessionRepository(gormDB)
	svc := auth.NewService(repo, sessions, cfg)

	r := server.New(gormDB)
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, svc, cfg)

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}