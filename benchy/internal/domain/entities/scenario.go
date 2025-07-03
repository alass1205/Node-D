package entities

import (
	"context"
	"time"
)

// ScenarioType représente le type de scénario
type ScenarioType string

const (
	ScenarioInit        ScenarioType = "init"
	ScenarioTransfers   ScenarioType = "transfers"
	ScenarioERC20       ScenarioType = "erc20"
	ScenarioReplacement ScenarioType = "replacement"
)

// ScenarioStatus représente l'état d'un scénario
type ScenarioStatus string

const (
	ScenarioStatusIdle    ScenarioStatus = "idle"
	ScenarioStatusRunning ScenarioStatus = "running"
	ScenarioStatusStopped ScenarioStatus = "stopped"
	ScenarioStatusFailed  ScenarioStatus = "failed"
)

// Scenario représente un scénario de test
type Scenario struct {
	ID          string         `json:"id"`
	Type        ScenarioType   `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      ScenarioStatus `json:"status"`
	
	// Configuration
	Duration    time.Duration `json:"duration"`
	Interval    time.Duration `json:"interval"`
	Repetitions int           `json:"repetitions"`
	
	// Progression
	CurrentStep int      `json:"current_step"`
	TotalSteps  int      `json:"total_steps"`
	Progress    float64  `json:"progress"`
	
	// Résultats
	TransactionHashes []string    `json:"transaction_hashes"`
	Errors           []string    `json:"errors"`
	Metrics          interface{} `json:"metrics"`
	
	// Timestamps
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// NewScenario crée un nouveau scénario
func NewScenario(scenarioType ScenarioType, name, description string) *Scenario {
	return &Scenario{
		ID:          generateScenarioID(),
		Type:        scenarioType,
		Name:        name,
		Description: description,
		Status:      ScenarioStatusIdle,
		TransactionHashes: make([]string, 0),
		Errors:           make([]string, 0),
	}
}

// Start démarre le scénario
func (s *Scenario) Start() {
	s.Status = ScenarioStatusRunning
	s.StartedAt = time.Now()
	s.CurrentStep = 0
	s.Progress = 0.0
}

// Complete termine le scénario
func (s *Scenario) Complete() {
	s.Status = ScenarioStatusStopped
	s.CompletedAt = time.Now()
	s.Progress = 100.0
}

// Fail marque le scénario comme échoué
func (s *Scenario) Fail(err error) {
	s.Status = ScenarioStatusFailed
	s.CompletedAt = time.Now()
	s.Errors = append(s.Errors, err.Error())
}

// AddTransactionHash ajoute un hash de transaction
func (s *Scenario) AddTransactionHash(hash string) {
	s.TransactionHashes = append(s.TransactionHashes, hash)
}

// UpdateProgress met à jour la progression
func (s *Scenario) UpdateProgress(currentStep, totalSteps int) {
	s.CurrentStep = currentStep
	s.TotalSteps = totalSteps
	if totalSteps > 0 {
		s.Progress = float64(currentStep) / float64(totalSteps) * 100
	}
}

// IsRunning retourne true si le scénario est en cours
func (s *Scenario) IsRunning() bool {
	return s.Status == ScenarioStatusRunning
}

// ScenarioRunner interface pour exécuter les scénarios
type ScenarioRunner interface {
	Run(ctx context.Context, scenario *Scenario, network *Network) error
}

// generateScenarioID génère un ID unique pour le scénario
func generateScenarioID() string {
	return "scenario-" + time.Now().Format("20060102-150405")
}
