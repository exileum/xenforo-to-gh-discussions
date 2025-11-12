package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/progress"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// InteractiveRunner handles the interactive migration flow
type InteractiveRunner struct {
	nonInteractive bool
}

// NewInteractiveRunner creates a new interactive migration runner
func NewInteractiveRunner(nonInteractive bool) *InteractiveRunner {
	return &InteractiveRunner{
		nonInteractive: nonInteractive,
	}
}

// Run executes the complete migration workflow with interactive prompts
func (r *InteractiveRunner) Run(cfg *config.Config) error {
	for {
		r.setProgressFile(cfg)

		if shouldContinue, err := r.handlePreMigrationSteps(cfg); err != nil {
			return err
		} else if !shouldContinue {
			continue
		}

		if err := r.runMigration(cfg); err != nil {
			if r.nonInteractive {
				return fmt.Errorf("migration failed: %w", err)
			}
			// Interactive mode: error handling continues the loop
		}

		if r.nonInteractive {
			break
		}

		if shouldContinue, err := r.handlePostMigrationSteps(cfg); err != nil {
			log.Printf("Error selecting categories: %v", err)
			break
		} else if !shouldContinue {
			break
		}
	}

	return nil
}

func (r *InteractiveRunner) setProgressFile(cfg *config.Config) {
	cfg.Migration.ProgressFile = fmt.Sprintf("migration_progress_node%d.json", cfg.GitHub.XenForoNodeID)
}

func (r *InteractiveRunner) handlePreMigrationSteps(cfg *config.Config) (bool, error) {
	if r.nonInteractive || cfg.Migration.DryRun {
		return true, nil
	}

	if config.PromptBool("Would you like to do a dry run first? (recommended)", true) {
		if err := r.runDryRun(cfg); err != nil {
			log.Printf("Dry run failed: %v", err)
			return false, nil
		}
	}

	return config.PromptBool("Start the actual migration now?", false), nil
}

func (r *InteractiveRunner) runMigration(cfg *config.Config) error {
	fmt.Printf("\nStarting migration of XenForo Node %d to GitHub Category %s...\n",
		cfg.GitHub.XenForoNodeID, cfg.GitHub.GitHubCategoryID)

	migrator := NewMigrator(cfg)
	ctx := context.Background()
	if err := migrator.Run(ctx); err != nil {
		if !r.nonInteractive {
			r.handleMigrationError(err, cfg)
		}
		return err
	}
	return nil
}

func (r *InteractiveRunner) handlePostMigrationSteps(cfg *config.Config) (bool, error) {
	fmt.Println("\nMigration complete!")
	if !config.PromptBool("Migrate another category?", true) {
		return false, nil
	}

	if err := r.selectNewCategories(cfg); err != nil {
		return false, err
	}
	return true, nil
}

// handleMigrationError handles errors during migration with retry/skip/abort options
func (r *InteractiveRunner) handleMigrationError(err error, cfg *config.Config) {
	fmt.Printf("\nError: %v\n\n", err)
	fmt.Println("What would you like to do?")
	fmt.Println("1. Retry now")
	fmt.Println("2. Skip this thread and continue")
	fmt.Println("3. Abort migration (progress saved)")

	choice := config.PromptChoice(1, 3)
	switch choice {
	case 1:
		// Retry will happen in the next loop iteration
		return
	case 2:
		// Skip this thread by incrementing the resume position
		fmt.Println("Skipping current thread...")

		// Get current progress to find last processed thread
		tracker, err := progress.NewTracker(context.Background(), cfg.Migration.ProgressFile, false)
		if err != nil {
			fmt.Printf("Warning: Could not load progress file: %v\n", err)
			return
		}

		// Set resume from next thread (increment by 1)
		progressData := tracker.GetProgress()
		nextThreadID := progressData.LastThreadID + 1
		cfg.Migration.ResumeFrom = nextThreadID

		fmt.Printf("Will resume from thread ID %d on retry\n", nextThreadID)
		return
	case 3:
		fmt.Printf("\nMigration aborted. To resume later, run with:\n")
		fmt.Printf("  --resume-from=%d\n", r.getLastProcessedID(cfg))
		os.Exit(1)
	}
}

// getLastProcessedID reads the progress file to get the last processed thread ID
func (r *InteractiveRunner) getLastProcessedID(cfg *config.Config) int {
	tracker, err := progress.NewTracker(context.Background(), cfg.Migration.ProgressFile, true) // dryRun=true just for reading
	if err != nil {
		return 0
	}

	progressData := tracker.GetProgress()
	return progressData.LastThreadID
}

// selectNewCategories prompts the user to select new source and target categories
func (r *InteractiveRunner) selectNewCategories(cfg *config.Config) error {
	fmt.Println("\n=== Select Next Migration ===")

	// Fetch XenForo categories
	fmt.Print("\nFetching XenForo categories... ")
	categories, err := config.ValidateXenForoAuth(cfg.XenForo.APIURL, cfg.XenForo.APIKey, cfg.XenForo.APIUser)
	if err != nil {
		return fmt.Errorf("failed to fetch XenForo categories: %w", err)
	}
	fmt.Printf("✓ Found %d categories\n\n", len(categories))

	// Mark already migrated category if same as current
	for i, cat := range categories {
		if cat.ID == fmt.Sprintf("%d", cfg.GitHub.XenForoNodeID) {
			categories[i].Name = cat.Name + " ✓ Already migrated"
		}
	}

	selectedCategory, err := config.PromptSelection("Select XenForo category to migrate:", categories)
	if err != nil {
		return err
	}

	nodeID, err := strconv.Atoi(selectedCategory.ID)
	if err != nil {
		return fmt.Errorf("failed to parse category ID '%s': %w", selectedCategory.ID, err)
	}
	cfg.GitHub.XenForoNodeID = nodeID

	// Fetch GitHub categories
	fmt.Print("\nFetching GitHub Discussion categories... ")
	ctx := context.Background()
	ghCategories, err := config.ValidateGitHubAuth(ctx, cfg.GitHub.Token, cfg.GitHub.Repository)
	if err != nil {
		return fmt.Errorf("failed to fetch GitHub categories: %w", err)
	}
	fmt.Printf("✓ Found %d categories\n\n", len(ghCategories))

	selectedGHCategory, err := config.PromptSelection("Select target GitHub Discussion category:", ghCategories)
	if err != nil {
		return err
	}

	cfg.GitHub.GitHubCategoryID = selectedGHCategory.ID

	return nil
}

// runDryRun performs a dry run of the migration
func (r *InteractiveRunner) runDryRun(cfg *config.Config) error {
	fmt.Println("\nRunning dry run...")

	// Create XenForo client
	client := xenforo.NewClient(cfg.XenForo.APIURL, cfg.XenForo.APIKey, cfg.XenForo.APIUser, cfg.Migration.MaxRetries)

	// Get statistics from XenForo API
	threadCount, postCount, attachmentCount, userCount, err := client.GetDryRunStats(context.Background(), cfg.GitHub.XenForoNodeID)
	if err != nil {
		return fmt.Errorf("failed to get dry run statistics: %w", err)
	}

	fmt.Println("\nDry run complete. Migration summary:")
	fmt.Println("┌─────────────┬────────┐")
	fmt.Println("│ Content     │ Count  │")
	fmt.Println("├─────────────┼────────┤")
	fmt.Printf("│ Threads     │ %6d │\n", threadCount)
	fmt.Printf("│ Posts       │ %6d │\n", postCount)
	fmt.Printf("│ Attachments │ %6d │\n", attachmentCount)
	fmt.Printf("│ Users       │ %6d │\n", userCount)
	fmt.Println("└─────────────┴────────┘")

	return nil
}
