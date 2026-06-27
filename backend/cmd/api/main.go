package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nicolas/finanzas/backend/internal/accounts"
	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/budgets"
	"github.com/nicolas/finanzas/backend/internal/categories"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/db"
	"github.com/nicolas/finanzas/backend/internal/middleware"
	"github.com/nicolas/finanzas/backend/internal/reports"
	"github.com/nicolas/finanzas/backend/internal/server"
	"github.com/nicolas/finanzas/backend/internal/transactions"
	"github.com/nicolas/finanzas/backend/internal/travel"
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

	// --- Repositories ---
	userRepo := auth.NewUserRepository(gormDB)
	sessions := auth.NewSessionRepository(gormDB)
	authSvc := auth.NewService(userRepo, sessions, cfg)

	accRepo := accounts.NewAccountRepository(gormDB)
	accSvc := accounts.NewService(accRepo)

	catRepo := categories.NewCategoryRepository(gormDB)
	catSvc := categories.NewService(catRepo)

	txRepo := transactions.NewRepository(gormDB)

	budgetRepo := budgets.NewRepository(gormDB)

	travelRepo := travel.NewRepository(gormDB)

	// --- Adapters ---
	// Each adapter translates one service into the small contract another
	// package expects. Keeping these in the *producing* package (the one
	// that owns the data) avoids cross-package cycles and keeps the
	// consumer free of concrete imports.
	txAccountAdapter := accounts.NewTransactionAccountAdapter(accSvc)
	txCategoryAdapter := categories.NewTransactionCategoryAdapter(catSvc)
	budgetCategoryAdapter := categories.NewBudgetCategoryAdapter(catSvc)
	travelUserAdapter := auth.NewTravelUserAdapter(userRepo)
	reportBudgetAdapter := budgets.NewReportBudgetAdapter(budgetRepo)

	// --- Services ---
	txSvc := transactions.NewService(txRepo, txAccountAdapter, txCategoryAdapter)
	budgetSvc := budgets.NewService(budgetRepo, budgetCategoryAdapter)
	travelSvc := travel.NewService(travelRepo, travelUserAdapter)
	reportSvc := reports.NewService(txRepo, reportBudgetAdapter)

	// --- HTTP ---
	r := server.New(gormDB)
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")

	// userID resolver closure — closes over authSvc (no cycle since main imports both)
	userIDResolver := func(token string) (string, error) {
		user, err := authSvc.Me(token)
		if err != nil {
			return "", err
		}
		return user.ID.String(), nil
	}
	requireUserID := middleware.RequireUserID(userIDResolver)

	auth.RegisterRoutes(api, authSvc, cfg)
	accounts.RegisterRoutes(api, accounts.NewHandler(accSvc), requireUserID)
	categories.RegisterRoutes(api, categories.NewHandler(catSvc), requireUserID)
	transactions.RegisterRoutes(api, transactions.NewHandler(txSvc), requireUserID)
	budgets.RegisterRoutes(api, budgets.NewHandler(budgetSvc), requireUserID)
	travel.RegisterRoutes(api, travel.NewHandler(travelSvc), requireUserID)
	reports.RegisterRoutes(api, reports.NewHandler(reportSvc), requireUserID)

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)

	// Graceful shutdown: explicit http.Server with timeouts (gin.Engine.Run uses
	// http.Server internally but with no ReadHeaderTimeout and no Shutdown hook).
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	case sig := <-quit:
		log.Printf("Received %s, shutting down gracefully...", sig)
	}

	// Stop accepting new connections; let in-flight requests finish (up to 15s).
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close DB pool so in-flight queries get a clean stop and file descriptors
	// are released before the process exits.
	if sqlDB, err := gormDB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("Server exited")
}