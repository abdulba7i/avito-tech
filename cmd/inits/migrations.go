package inits

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB) error {
	migrationsPath := "migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "../../migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found")
		}
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, filename := range files {
		content, err := os.ReadFile(filepath.Join(migrationsPath, filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		sql := extractUpSection(string(content))
		if sql == "" {
			continue
		}

		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("execute migration %s: %w", filename, err)
		}
	}

	return nil
}

func extractUpSection(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUpSection := false

	for _, line := range lines {
		if strings.Contains(line, "-- +goose Up") {
			inUpSection = true
			continue
		}
		if strings.Contains(line, "-- +goose Down") {
			break
		}
		if inUpSection {
			if strings.Contains(line, "-- +goose") {
				continue
			}
			upLines = append(upLines, line)
		}
	}

	return strings.Join(upLines, "\n")
}
