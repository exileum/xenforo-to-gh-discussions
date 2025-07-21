package progress

import (
	"context"
	"encoding/json"
	"log"
	"os"
)

type Persistence struct {
	filePath string
}

func NewPersistence(filePath string) *Persistence {
	return &Persistence{
		filePath: filePath,
	}
}

func (p *Persistence) Load(ctx context.Context) (*MigrationProgress, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	progress := &MigrationProgress{
		CompletedThreads: []int{},
		FailedThreads:    []int{},
	}

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return progress, err
	}

	err = json.Unmarshal(data, progress)
	if err != nil {
		log.Printf("Failed to unmarshal progress data from %s: %v", p.filePath, err)
		log.Printf("Using default progress state instead of corrupted data")
		return &MigrationProgress{
			CompletedThreads: []int{},
			FailedThreads:    []int{},
		}, err
	}

	return progress, nil
}

func (p *Persistence) Save(ctx context.Context, progress *MigrationProgress) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal progress data: %v", err)
		return err
	}

	err = os.WriteFile(p.filePath, data, 0644)
	if err != nil {
		log.Printf("Failed to save progress to %s: %v", p.filePath, err)
		return err
	}

	return nil
}
