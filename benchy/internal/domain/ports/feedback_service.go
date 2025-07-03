package ports

import (
	"context"

)

// FeedbackService définit les opérations de feedback utilisateur
type FeedbackService interface {
	// Messages de feedback
	Info(ctx context.Context, message string) error
	Success(ctx context.Context, message string) error
	Warning(ctx context.Context, message string) error
	Error(ctx context.Context, message string) error
	
	// Progress tracking
	StartProgress(ctx context.Context, title string, total int) (ProgressTracker, error)
	
	// Spinners
	StartSpinner(ctx context.Context, message string) (Spinner, error)
	
	// Tables et formatting
	DisplayTable(ctx context.Context, headers []string, rows [][]string) error
	DisplayJSON(ctx context.Context, data interface{}) error
	
	// Interactive
	Confirm(ctx context.Context, message string) (bool, error)
	Input(ctx context.Context, prompt string) (string, error)
}

// ProgressTracker représente un tracker de progression
type ProgressTracker interface {
	Update(current int, message string) error
	Increment(message string) error
	Complete(message string) error
	Error(message string) error
	Close() error
}

// Spinner représente un spinner d'attente
type Spinner interface {
	UpdateMessage(message string) error
	Success(message string) error
	Error(message string) error
	Stop() error
}

// FeedbackLevel représente le niveau de feedback
type FeedbackLevel string

const (
	FeedbackLevelSilent  FeedbackLevel = "silent"
	FeedbackLevelNormal  FeedbackLevel = "normal"
	FeedbackLevelVerbose FeedbackLevel = "verbose"
	FeedbackLevelDebug   FeedbackLevel = "debug"
)

// FeedbackConfig représente la configuration du feedback
type FeedbackConfig struct {
	Level     FeedbackLevel
	Colors    bool
	Timestamp bool
	Output    string // "stdout", "stderr", "file"
}
