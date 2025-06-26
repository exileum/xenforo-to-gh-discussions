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
		dryRun         = flag.Bool("dry-run", false, "Run in dry-run mode (no actual API calls)")
		resumeFrom     = flag.Int("resume-from", 0, "Resume from specific thread ID")
		verbose        = flag.Bool("verbose", false, "Enable verbose logging")
		nonInteractive = flag.Bool("non-interactive", false, "Run in non-interactive mode using environment variables")
	)
	flag.Parse()

	// Validate command line flags
	if *resumeFrom < 0 {
		log.Fatalf("resume-from must be a positive value, got: %d", *resumeFrom)
	}

	// Load configuration
	var cfg *config.Config
	if *nonInteractive {
		cfg = config.New()
	} else {
		cfg = config.InteractiveConfig()
	}

	// Apply command line overrides
	cfg.Migration.DryRun = *dryRun
	cfg.Migration.Verbose = *verbose
	cfg.Migration.ResumeFrom = *resumeFrom

	// Run an interactive migration workflow
	runner := migration.NewInteractiveRunner(*nonInteractive)
	if err := runner.Run(cfg); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
