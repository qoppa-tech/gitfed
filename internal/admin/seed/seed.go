package seed

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qoppa-tech/gitfed/internal/database/sqlc"
	gitmod "github.com/qoppa-tech/gitfed/internal/modules/git"
	"github.com/qoppa-tech/gitfed/internal/modules/organization"
	"github.com/qoppa-tech/gitfed/internal/modules/user"
)

const (
	defaultUserID = "11111111-1111-7111-8111-111111111111"
	defaultOrgID  = "22222222-2222-7222-8222-222222222222"
	defaultRepoID = "33333333-3333-7333-8333-333333333333"
)

func Run(ctx context.Context, db *pgxpool.Pool) error {
	repoID, err := uuid.Parse(defaultRepoID)
	if err != nil {
		return fmt.Errorf("parse repository id: %w", err)
	}

	q := sqlc.New(db)
	userStore := user.NewStore(q)
	userSvc := user.NewService(userStore)
	orgStore := organization.NewStore(q)
	orgSvc := organization.NewService(orgStore)
	repoStore := gitmod.NewStore(q)

	adminUser, err := userStore.GetByUsername(ctx, "admin")
	if err != nil {
		if err != user.ErrNotFound {
			return fmt.Errorf("get admin user: %w", err)
		}
		adminUser, err = userSvc.Register(ctx, user.RegisterInput{
			Name:     "Admin User",
			Username: "admin",
			Password: "admin123",
			Email:    "admin@gitfed.local",
		})
		if err != nil {
			return fmt.Errorf("register admin user: %w", err)
		}
	}

	orgID, err := ensureOrganization(ctx, orgSvc, adminUser.ID)
	if err != nil {
		return err
	}

	if err := ensureRepository(ctx, repoStore, repoID, adminUser.ID); err != nil {
		return err
	}

	_ = orgID
	return nil
}

func ensureOrganization(ctx context.Context, orgSvc *organization.Service, userID uuid.UUID) (uuid.UUID, error) {
	orgs, err := orgSvc.GetByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("list user organizations: %w", err)
	}

	for _, org := range orgs {
		if org.Name == "Gitfed Team" {
			return org.ID, nil
		}
	}

	org, err := orgSvc.Create(ctx, organization.CreateInput{
		Name:        "Gitfed Team",
		Description: "Default seeded organization",
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create organization: %w", err)
	}

	if err := orgSvc.AddUser(ctx, org.ID, userID); err != nil {
		return uuid.Nil, fmt.Errorf("add user to organization: %w", err)
	}

	return org.ID, nil
}

func ensureRepository(ctx context.Context, repoStore *gitmod.Store, repoID, ownerID uuid.UUID) error {
	_, err := repoStore.GetByName(ctx, ownerID, "hello-gitfed")
	if err == nil {
		return nil
	}
	if err != gitmod.ErrRepoNotFound {
		return fmt.Errorf("get repository: %w", err)
	}

	_, err = repoStore.Create(ctx, gitmod.CreateInput{
		Id:          repoID,
		Name:        "hello-gitfed",
		Description: "Default seeded repository",
		IsPrivate:   false,
		OwnerID:     ownerID,
		DefaultRef:  "refs/heads/main",
	})
	if err != nil {
		return fmt.Errorf("create repository: %w", err)
	}
	return nil
}
