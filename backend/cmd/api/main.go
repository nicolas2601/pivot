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
	"github.com/google/uuid"

	"github.com/nicolas/finanzas/backend/internal/accounts"
	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/budgets"
	"github.com/nicolas/finanzas/backend/internal/categories"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/db"
	"github.com/nicolas/finanzas/backend/internal/goals"
	"github.com/nicolas/finanzas/backend/internal/middleware"
	"github.com/nicolas/finanzas/backend/internal/recurring"
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

	// DB_SKIP_AUTO_MIGRATE=true disables in-process migration when the
	// database is managed externally (e.g. Supabase CLI pushes migrations).
	// Local docker-compose dev keeps the default (auto-migrate on boot).
	if os.Getenv("DB_SKIP_AUTO_MIGRATE") == "true" {
		log.Println("DB_SKIP_AUTO_MIGRATE=true: skipping in-process migrations")
	} else if err := db.Migrate(cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// --- Repositories ---
	userRepo := auth.NewUserRepository(gormDB)
	sessions := auth.NewSessionRepository(gormDB)
	baseAuthSvc := auth.NewService(userRepo, sessions, cfg)

	catRepo := categories.NewCategoryRepository(gormDB)
	catSvc := categories.NewService(catRepo)

	// Wire the categories service into auth so registration seeds default ES
	// categories for the new user. auth.WithCategorySeeder keeps the dep
	// arrow one-way (auth declares the interface; main wires the concrete).
	authSvc := auth.WithCategorySeeder(baseAuthSvc, catSvc)

	accRepo := accounts.NewAccountRepository(gormDB)
	accSvc := accounts.NewService(accRepo)

	txRepo := transactions.NewRepository(gormDB)

	budgetRepo := budgets.NewRepository(gormDB)

	travelRepo := travel.NewRepository(gormDB)

	goalRepo := goals.NewRepository(gormDB)

	recurringRepo := recurring.NewRepository(gormDB)

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
	goalAccountAdapter := accounts.NewGoalsAccountAdapter(accSvc)
	recurringAccountAdapter := accounts.NewRecurringAccountAdapter(accSvc)
	recurringCategoryAdapter := categories.NewRecurringCategoryAdapter(catSvc)
	recurringUserAdapter := auth.NewRecurringUserResolverAdapter()

	// --- Services ---
	txSvc := transactions.NewService(txRepo, txAccountAdapter, txCategoryAdapter)
	budgetSvc := budgets.NewService(budgetRepo, budgetCategoryAdapter)
	travelSvc := travel.NewService(travelRepo, travelUserAdapter)

	// Lightweight projections for the reports service. The reports package
	// only needs name + color, so we copy from the full model on each call
	// (one query per dashboard load — cheap because gorm caches the query
	// plan and the underlying query is index-only on user_id).
	categoriesListForReports := func(userID uuid.UUID, _ string) ([]reports.CategoryLite, error) {
		cs, err := catSvc.List(userID, "")
		if err != nil {
			return nil, err
		}
		out := make([]reports.CategoryLite, 0, len(cs))
		for _, c := range cs {
			color := ""
			if c.Color != nil {
				color = *c.Color
			}
			out = append(out, reports.CategoryLite{ID: c.ID, Name: c.Name, Color: color})
		}
		return out, nil
	}
	accountsListForReports := func(userID uuid.UUID) ([]reports.AccountLite, error) {
		as, err := accSvc.List(userID)
		if err != nil {
			return nil, err
		}
		out := make([]reports.AccountLite, 0, len(as))
		for _, a := range as {
			out = append(out, reports.AccountLite{ID: a.ID, Name: a.Name})
		}
		return out, nil
	}

	reportSvc := reports.NewService(
		txRepo,
		reportBudgetAdapter,
		reports.CategoriesAdapter(categoriesListForReports),
		reports.AccountsAdapter(accountsListForReports),
	)
	goalSvc := goals.NewService(goalRepo, goalAccountAdapter)
	recurringTxAdapter := transactions.NewRecurringTxCreatorAdapter(txSvc)
	recurringSvc := recurring.NewService(recurringRepo, recurringAccountAdapter, recurringCategoryAdapter, recurringTxAdapter, recurringUserAdapter)

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
	goals.RegisterRoutes(api, goals.NewHandler(goalSvc), requireUserID)
	recurring.RegisterRoutes(api, recurring.NewHandler(recurringSvc), requireUserID)

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