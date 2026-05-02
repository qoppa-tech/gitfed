package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/qoppa-tech/gitfed/internal/config"
	"github.com/qoppa-tech/gitfed/internal/database"
)

const (
	seedUserID = "11111111-1111-7111-8111-111111111111"
	seedOrgID  = "22222222-2222-7222-8222-222222222222"
	seedRepoID = "33333333-3333-7333-8333-333333333333"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("database connect failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	switch os.Args[1] {
	case "seed":
		if err := seed(ctx, db); err != nil {
			slog.Error("seed failed", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("seed completed")
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("usage: go run ./cmd/admin <seed>")
}

func seed(ctx context.Context, db *pgxpool.Pool) error {
	userID, err := uuid.Parse(seedUserID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}
	orgID, err := uuid.Parse(seedOrgID)
	if err != nil {
		return fmt.Errorf("parse organization id: %w", err)
	}
	repoID, err := uuid.Parse(seedRepoID)
	if err != nil {
		return fmt.Errorf("parse repository id: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO users (user_id, name, username, password, email, is_verified)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		ON CONFLICT (username) DO UPDATE SET
			name = EXCLUDED.name,
			password = EXCLUDED.password,
			email = EXCLUDED.email,
			updated_at = now()
	`, userID, "Admin User", "admin", string(hashedPassword), "admin@gitfed.local"); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO organizations (organization_id, organization_name, organization_description)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id) DO UPDATE SET
			organization_name = EXCLUDED.organization_name,
			organization_description = EXCLUDED.organization_description
	`, orgID, "Gitfed Team", "Default seeded organization"); err != nil {
		return fmt.Errorf("upsert organization: %w", err)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO organization_users (organization_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (organization_id, user_id) DO NOTHING
	`, orgID, userID); err != nil {
		return fmt.Errorf("link organization user: %w", err)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO git_repository (id, name, description, is_private, is_deleted, owner_id, default_ref, head)
		VALUES ($1, $2, $3, FALSE, FALSE, $4, 'refs/heads/main', '')
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			owner_id = EXCLUDED.owner_id,
			updated_at = now()
	`, repoID, "hello-gitfed", "Default seeded repository", userID); err != nil {
		return fmt.Errorf("upsert repository: %w", err)
	}

	return nil
}
