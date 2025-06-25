package main

import (
	"flag"
	"log"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/migration"
)

func main() {
	// Parse command line flags
	var (
		dryRun     = flag.Bool("dry-run", false, "Run in dry-run mode (no actual API calls)")
		resumeFrom = flag.Int("resume-from", 0, "Resume from specific thread ID")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Initialize configuration
	cfg := config.New()
	cfg.Migration.DryRun = *dryRun
	cfg.Migration.Verbose = *verbose
	cfg.Migration.ResumeFrom = *resumeFrom

	// Run migration
	migrator := migration.NewMigrator(cfg)
	if err := migrator.Run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
