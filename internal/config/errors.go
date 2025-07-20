package config

import (
	"errors"
	"fmt"
)

// Error classification for better error handling throughout the application.

// ConfigurationError represents errors during configuration validation or setup.
type ConfigurationError struct {
	Field   string // The configuration field that caused the error
	Message string // Human-readable error message
	Cause   error  // Underlying error cause (optional)
}

func (e *ConfigurationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("configuration error in field '%s': %s (caused by: %v)", e.Field, e.Message, e.Cause)
	}
	return fmt.Sprintf("configuration error in field '%s': %s", e.Field, e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Cause
}

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(field, message string) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Message: message,
	}
}

// NewConfigurationErrorWithCause creates a new configuration error with an underlying cause.
func NewConfigurationErrorWithCause(field, message string, cause error) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Message: message,
		Cause:   cause,
	}
}

// ValidationError represents validation failures for specific fields.
type ValidationError struct {
	Field string // The field that failed validation
	Value string // The invalid value
	Rule  string // The validation rule that was violated
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': value '%s' violates rule '%s'", e.Field, e.Value, e.Rule)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, value, rule string) *ValidationError {
	return &ValidationError{
		Field: field,
		Value: value,
		Rule:  rule,
	}
}

// Sentinel errors for common configuration issues
var (
	// ErrMissingRequiredField indicates a required configuration field is missing
	ErrMissingRequiredField = errors.New("required configuration field is missing")

	// ErrInvalidURL indicates an invalid URL format
	ErrInvalidURL = errors.New("invalid URL format")

	// ErrInvalidToken indicates an invalid authentication token
	ErrInvalidToken = errors.New("invalid authentication token")

	// ErrInvalidRepository indicates an invalid repository format
	ErrInvalidRepository = errors.New("invalid repository format")

	// ErrInvalidCategory indicates an invalid category configuration
	ErrInvalidCategory = errors.New("invalid category configuration")

	// ErrInvalidNodeID indicates an invalid node ID
	ErrInvalidNodeID = errors.New("invalid node ID")

	// ErrInvalidUserMapping indicates an invalid user mapping configuration
	ErrInvalidUserMapping = errors.New("invalid user mapping configuration")

	// ErrInvalidFilesystemPath indicates an invalid filesystem path
	ErrInvalidFilesystemPath = errors.New("invalid filesystem path")

	// ErrInvalidRateLimit indicates an invalid rate limiting configuration
	ErrInvalidRateLimit = errors.New("invalid rate limiting configuration")

	// ErrInvalidRetryConfiguration indicates an invalid retry configuration
	ErrInvalidRetryConfiguration = errors.New("invalid retry configuration")
)

// IsConfigurationError checks if an error is a configuration error.
func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// GetConfigurationField extracts the field name from a configuration error.
func GetConfigurationField(err error) string {
	var configErr *ConfigurationError
	if errors.As(err, &configErr) {
		return configErr.Field
	}
	return ""
}

// GetValidationRule extracts the validation rule from a validation error.
func GetValidationRule(err error) string {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Rule
	}
	return ""
}
