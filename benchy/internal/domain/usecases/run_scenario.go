package usecases

import (
	"context"
	"fmt"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// RunScenarioUseCase gère l'exécution des scénarios
type RunScenarioUseCase struct {
	networkRepo ports.NetworkRepository
	ethService  ports.EthereumService
	feedback    ports.FeedbackService
}

// NewRunScenarioUseCase crée une nouvelle instance
func NewRunScenarioUseCase(
	networkRepo ports.NetworkRepository,
	ethService ports.EthereumService,
	feedback ports.FeedbackService,
) *RunScenarioUseCase {
	return &RunScenarioUseCase{
		networkRepo: networkRepo,
		ethService:  ethService,
		feedback:    feedback,
	}
}

// Execute exécute le scénario spécifié
func (uc *RunScenarioUseCase) Execute(ctx context.Context, scenarioType entities.ScenarioType) error {
	return fmt.Errorf("scenario execution not implemented yet")
}
