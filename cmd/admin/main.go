package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/qoppa-tech/gitfed/internal/admin/seed"
	"github.com/qoppa-tech/gitfed/internal/config"
	"github.com/qoppa-tech/gitfed/internal/database"
	"golang.org/x/term"
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
		input, err := promptSeedInput()
		if err != nil {
			slog.Error("seed input failed", slog.Any("error", err))
			os.Exit(1)
		}

		if err := seed.Run(ctx, db, input); err != nil {
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

func promptSeedInput() (seed.Input, error) {
	reader := bufio.NewReader(os.Stdin)
	name, err := promptWithDefault(reader, "Admin name", "Admin User")
	if err != nil {
		return seed.Input{}, err
	}
	username, err := promptWithDefault(reader, "Admin username", "admin")
	if err != nil {
		return seed.Input{}, err
	}
	email, err := promptWithDefault(reader, "Admin email", "admin@gitfed.local")
	if err != nil {
		return seed.Input{}, err
	}

	fmt.Print("Admin password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return seed.Input{}, fmt.Errorf("read password: %w", err)
	}
	password := strings.TrimSpace(string(passBytes))
	if password == "" {
		return seed.Input{}, fmt.Errorf("admin password is required")
	}

	return seed.Input{
		AdminName:     name,
		AdminUsername: username,
		AdminEmail:    email,
		AdminPassword: password,
	}, nil
}

func promptWithDefault(reader *bufio.Reader, label, def string) (string, error) {
	fmt.Printf("%s [%s]: ", label, def)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}
	v := strings.TrimSpace(raw)
	if v == "" {
		return def, nil
	}
	return v, nil
}
