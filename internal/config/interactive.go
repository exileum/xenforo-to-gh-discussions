package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

// SelectOption represents an option in a selection list
type SelectOption struct {
	ID   string
	Name string
	Info string // e.g., "(234 threads)"
}

// PromptString prompts for a string value with a default
func PromptString(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	// Trim whitespace to handle copy-paste issues
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	return input
}

// PromptPassword prompts for a password/token without showing a default
func PromptPassword(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	// Trim whitespace to handle copy-paste issues
	return strings.TrimSpace(input)
}

// PromptInt prompts for an integer value with a default
func PromptInt(prompt string, defaultValue int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%d]: ", prompt, defaultValue)

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid number. Please try again.\n")
			continue
		}

		if value <= 0 {
			fmt.Printf("Please enter a positive number.\n")
			continue
		}

		return value
	}
}

// PromptDuration prompts for a duration value with a default
func PromptDuration(prompt string, defaultValue time.Duration) time.Duration {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		duration, err := time.ParseDuration(input)
		if err != nil {
			fmt.Printf("Invalid duration. Use format like '500ms', '1s', '2m'. Please try again.\n")
			continue
		}

		return duration
	}
}

// PromptBool prompts for a boolean value with a default
func PromptBool(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		if defaultValue {
			fmt.Printf("%s [Y/n]: ", prompt)
		} else {
			fmt.Printf("%s [y/N]: ", prompt)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultValue
		}

		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Printf("Please enter 'y' or 'n'.\n")
		}
	}
}

// PromptSelection displays a list of options and returns the selected item
func PromptSelection(prompt string, options []SelectOption) (SelectOption, error) {
	if len(options) == 0 {
		return SelectOption{}, fmt.Errorf("no options available")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(prompt)
	for i, option := range options {
		if option.Info != "" {
			fmt.Printf("%d. [%s] %s %s\n", i+1, option.ID, option.Name, option.Info)
		} else {
			fmt.Printf("%d. [%s] %s\n", i+1, option.ID, option.Name)
		}
	}

	for {
		fmt.Print("Enter number: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return SelectOption{}, err
		}

		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Please enter a valid number.\n")
			continue
		}

		if choice < 1 || choice > len(options) {
			fmt.Printf("Please enter a number between 1 and %d.\n", len(options))
			continue
		}

		return options[choice-1], nil
	}
}

// PromptChoice prompts for a numeric choice between min and max
func PromptChoice(min, max int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Choose [%d-%d]: ", min, max)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Please enter a valid number.\n")
			continue
		}

		if choice < min || choice > max {
			fmt.Printf("Please enter a number between %d and %d.\n", min, max)
			continue
		}

		return choice
	}
}

// InteractiveConfig creates a new config by prompting the user
func InteractiveConfig() *Config {
	fmt.Println("=== XenForo to GitHub Discussions Migration Tool ===")
	fmt.Println()

	cfg := &Config{}

	// XenForo Configuration
	fmt.Println("XenForo Configuration:")
	// XenForo credential validation with retry loop
	const maxRetries = 3
	var categories []SelectOption
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt == 1 {
			// First attempt: collect initial credentials
			cfg.XenForo.APIURL = PromptString("API URL", getEnvOrDefault("XENFORO_API_URL", "https://your-forum.com/api"))

			// For API Key, check if environment variable exists
			apiKeyEnv := os.Getenv("XENFORO_API_KEY")
			if apiKeyEnv != "" {
				cfg.XenForo.APIKey = apiKeyEnv
				fmt.Printf("API Key: ********** (from environment)\n")
			} else {
				cfg.XenForo.APIKey = PromptPassword("API Key")
			}

			defaultAPIUser := getEnvIntOrDefault("XENFORO_API_USER", 1)
			cfg.XenForo.APIUser = strconv.Itoa(PromptInt("API User", defaultAPIUser))
		} else {
			// Retry attempts: re-prompt for credentials
			fmt.Printf("\nRetry attempt %d of %d\n", attempt, maxRetries)
			fmt.Println("Please check your credentials and try again:")

			cfg.XenForo.APIURL = PromptString("API URL", cfg.XenForo.APIURL)
			cfg.XenForo.APIKey = PromptPassword("API Key")
			cfg.XenForo.APIUser = strconv.Itoa(PromptInt("API User", 1))
		}

		// Validate XenForo credentials
		fmt.Print("Validating XenForo credentials... ")
		categories, err = ValidateXenForoAuth(cfg.XenForo.APIURL, cfg.XenForo.APIKey, cfg.XenForo.APIUser)
		if err == nil {
			fmt.Println("✓ Connected successfully")
			break
		}

		fmt.Printf("✗ %v\n", err)

		if attempt == maxRetries {
			fmt.Printf("\nMaximum retry attempts (%d) reached. Exiting.\n", maxRetries)
			os.Exit(1)
		}
	}

	// Select XenForo category
	fmt.Printf("\nFetching XenForo categories... ")
	fmt.Printf("✓ Found %d categories\n\n", len(categories))

	selectedCategory, err := PromptSelection("Select XenForo category to migrate:", categories)
	if err != nil {
		fmt.Printf("Error selecting category: %v\n", err)
		os.Exit(1)
	}

	nodeID, _ := strconv.Atoi(selectedCategory.ID)
	cfg.GitHub.XenForoNodeID = nodeID

	// GitHub Configuration
	fmt.Println("\nGitHub Configuration:")

	// For GitHub Token, check if environment variable exists
	tokenEnv := os.Getenv("GITHUB_TOKEN")
	if tokenEnv != "" {
		cfg.GitHub.Token = tokenEnv
		fmt.Printf("Personal Access Token: ********** (from environment)\n")
	} else {
		cfg.GitHub.Token = PromptPassword("Personal Access Token (needs 'repo' and 'discussion' permissions)")
	}

	cfg.GitHub.Repository = PromptString("Repository", getEnvOrDefault("GITHUB_REPO", "your_username/your_repo"))

	// Validate GitHub token immediately
	fmt.Print("Validating GitHub token... ")

	ghCategories, err := ValidateGitHubAuth(cfg.GitHub.Token, cfg.GitHub.Repository)
	if err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Token has required permissions")

	// Select GitHub category
	fmt.Printf("\nFetching GitHub Discussion categories... ")
	fmt.Printf("✓ Found %d categories\n\n", len(ghCategories))

	selectedGHCategory, err := PromptSelection("Select target GitHub Discussion category:", ghCategories)
	if err != nil {
		fmt.Printf("Error selecting category: %v\n", err)
		os.Exit(1)
	}

	cfg.GitHub.GitHubCategoryID = selectedGHCategory.ID

	// Migration Settings
	fmt.Println("\nMigration Settings:")
	cfg.Migration.MaxRetries = PromptInt("Max Retries", getEnvIntOrDefault("MAX_RETRIES", 3))
	cfg.Migration.ProgressFile = fmt.Sprintf("migration_progress_node%d.json", cfg.GitHub.XenForoNodeID)

	// Filesystem Settings
	cfg.Filesystem.AttachmentsDir = PromptString("Attachments Directory", getEnvOrDefault("ATTACHMENTS_DIR", "./attachments"))
	cfg.Filesystem.AttachmentRateLimitDelay = PromptDuration("Attachment Rate Limit Delay", getEnvDurationOrDefault("ATTACHMENT_RATE_LIMIT_DELAY", 500*time.Millisecond))

	// Set other defaults
	cfg.Migration.UserMapping = make(map[int]int)

	return cfg
}

// ValidateXenForoAuth validates XenForo credentials and returns available categories
func ValidateXenForoAuth(apiURL, apiKey string, userID string) ([]SelectOption, error) {
	// Create a temporary client for validation
	client := xenforo.NewClient(apiURL, apiKey, userID, 3)

	// Test connection
	if err := client.TestConnection(); err != nil {
		return nil, err
	}

	// Fetch actual categories from XenForo API
	nodes, err := client.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}

	// Filter nodes to only include forum nodes and convert to SelectOptions
	var categories []SelectOption
	for _, node := range nodes {
		// Only include forum type nodes that are displayed in lists
		if node.NodeTypeID == "Forum" && node.DisplayInList {
			threadInfo := ""
			if node.ThreadCount != nil && *node.ThreadCount > 0 {
				threadInfo = fmt.Sprintf("(%d threads)", *node.ThreadCount)
			}
			categories = append(categories, SelectOption{
				ID:   strconv.Itoa(node.NodeID),
				Name: node.Title,
				Info: threadInfo,
			})
		}
	}

	return categories, nil
}

// ValidateGitHubAuth validates GitHub token and returns available discussion categories
func ValidateGitHubAuth(token, repository string) ([]SelectOption, error) {
	// Create a temporary client for validation
	client, err := github.NewClient(token)
	if err != nil {
		return nil, err
	}

	// Get repository info including categories
	info, err := client.GetRepositoryInfo(repository)
	if err != nil {
		return nil, err
	}

	// Convert to SelectOptions
	options := make([]SelectOption, len(info.DiscussionCategories))
	for i, cat := range info.DiscussionCategories {
		options[i] = SelectOption{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	return options, nil
}
