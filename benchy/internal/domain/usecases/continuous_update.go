package usecases

import (
	"context"
	"fmt"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// ContinuousUpdateUseCase gère les mises à jour continues
type ContinuousUpdateUseCase struct {
	networkRepo ports.NetworkRepository
	feedback    ports.FeedbackService
}

// NewContinuousUpdateUseCase crée une nouvelle instance
func NewContinuousUpdateUseCase(
	networkRepo ports.NetworkRepository,
	feedback ports.FeedbackService,
) *ContinuousUpdateUseCase {
	return &ContinuousUpdateUseCase{
		networkRepo: networkRepo,
		feedback:    feedback,
	}
}

// StartContinuousUpdate démarre la mise à jour continue
func (uc *ContinuousUpdateUseCase) StartContinuousUpdate(
	ctx context.Context,
	updateFunc func(context.Context) error,
	interval time.Duration,
) error {
	uc.feedback.Info(ctx, fmt.Sprintf("🔄 Starting continuous update (interval: %v)", interval))
	uc.feedback.Info(ctx, "💡 Press Ctrl+C to stop")
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	// Première exécution immédiate
	if err := updateFunc(ctx); err != nil {
		uc.feedback.Error(ctx, fmt.Sprintf("Initial update failed: %v", err))
	}
	
	for {
		select {
		case <-ticker.C:
			if err := updateFunc(ctx); err != nil {
				uc.feedback.Error(ctx, fmt.Sprintf("Update failed: %v", err))
			}
		case <-ctx.Done():
			uc.feedback.Info(ctx, "🔄 Stopping continuous update...")
			return ctx.Err()
		}
	}
}

// CreateInfosUpdateFunc crée une fonction de mise à jour pour les infos
func (uc *ContinuousUpdateUseCase) CreateInfosUpdateFunc(
	monitorUseCase *MonitorNetworkUseCase,
) func(context.Context) error {
	return func(ctx context.Context) error {
		// Clear screen
		fmt.Print("\033[2J\033[H")
		
		// Afficher timestamp
		uc.feedback.Info(ctx, fmt.Sprintf("📊 Network Information (Last update: %s)", time.Now().Format("15:04:05")))
		
		// Exécuter le monitoring
		return monitorUseCase.Execute(ctx, 0) // 0 = mode one-shot
	}
}

// CreateScenarioUpdateFunc crée une fonction de mise à jour pour les scénarios
func (uc *ContinuousUpdateUseCase) CreateScenarioUpdateFunc(
	scenarioUseCase *RunScenarioUseCase,
	scenarioType entities.ScenarioType,
) func(context.Context) error {
	return func(ctx context.Context) error {
		return scenarioUseCase.Execute(ctx, scenarioType)
	}
}
