package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nicolas/finanzas/backend/internal/accounts"
	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/categories"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/db"
	"github.com/nicolas/finanzas/backend/internal/middleware"
	"github.com/nicolas/finanzas/backend/internal/server"
)

// userIDFromToken is a small adapter for the userID middleware.
func userIDFromToken(svc any) func(string) (string, error) {
	return func(token string) (string, error) {
		// Type assert at call site via the registered Service type.
		type meRet struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		// Defer to the real service.
		u, err := callMe(svc, token)
		if err != nil {
			return "", err
		}
		_ = meRet{}
		uid, err := uuid.Parse(u.ID)
		if err != nil {
			return "", err
		}
		return uid.String(), nil
	}
}

// callMe is a thin wrapper that satisfies the interface expected by the
// userID middleware without leaking the auth import into every package.
type meCaller interface {
	Me(token string) (interface{ ID() string }, error)
}

func callMe(svc any, token string) (struct{ ID string }, error) {
	type userIDer interface {
		Me(token string) (any, error)
	}
	// The actual auth.Service.Me returns *auth.User; we don't want to import
	// auth here for dependency cleanliness — but the simplest approach is to
	// just type-assert.
	type concrete interface {
		Me(token string) (any, error)
	}
	c, ok := svc.(concrete)
	if !ok {
		return struct{ ID string }{}, nil
	}
	u, err := c.Me(token)
	if err != nil {
		return struct{ ID string }{}, err
	}
	type withID interface {
		GetID() string
	}
	if wid, ok := u.(withID); ok {
		return struct{ ID string }{ID: wid.GetID()}, nil
	}
	// Reflective fallback: assume the type has a string-convertible ID via JSON
	// marshaling. auth.User.ID is uuid.UUID, which marshals to a string.
	type stringer interface{ String() string }
	if s, ok := u.(stringer); ok {
		return struct{ ID string }{ID: s.String()}, nil
	}
	return struct{ ID string }{}, nil
}

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

	userRepo := auth.NewUserRepository(gormDB)
	sessions := auth.NewSessionRepository(gormDB)
	authSvc := auth.NewService(userRepo, sessions, cfg)

	accRepo := accounts.NewAccountRepository(gormDB)
	accSvc := accounts.NewService(accRepo)

	catRepo := categories.NewCategoryRepository(gormDB)
	catSvc := categories.NewService(catRepo)

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

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}