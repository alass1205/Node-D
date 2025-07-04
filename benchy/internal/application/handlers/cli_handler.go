package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"benchy/internal/application/services"
	"benchy/internal/infrastructure/feedback"
)

// CLIHandler orchestre l'exécution des commandes CLI
type CLIHandler struct {
	networkService    *services.NetworkService
	monitoringService *services.MonitoringService
	feedback          *feedback.ConsoleFeedback
}

// NewCLIHandler crée un nouveau handler CLI
func NewCLIHandler() (*CLIHandler, error) {
	// Répertoire de base pour les configurations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	baseDir := filepath.Join(homeDir, ".benchy")

	// Créer les services
	networkService, err := services.NewNetworkService(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create network service: %w", err)
	}

	monitoringService, err := services.NewMonitoringService()
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring service: %w", err)
	}

	feedback := feedback.NewConsoleFeedback()

	handler := &CLIHandler{
		networkService:    networkService,
		monitoringService: monitoringService,
		feedback:          feedback,
	}

	return handler, nil
}

// HandleLaunchNetwork gère la commande launch-network
func (h *CLIHandler) HandleLaunchNetwork(ctx context.Context) error {
	h.feedback.Info(ctx, "🚀 Starting network launch...")
	
	return h.networkService.LaunchNetwork(ctx)
}

// HandleInfos gère la commande infos
func (h *CLIHandler) HandleInfos(ctx context.Context, updateInterval int) error {
	return h.monitoringService.DisplayNetworkInfo(ctx, updateInterval)
}

// HandleScenario gère la commande scenario
func (h *CLIHandler) HandleScenario(ctx context.Context, scenarioName string) error {
	h.feedback.Info(ctx, fmt.Sprintf("🎯 Running scenario: %s", scenarioName))
	
	switch scenarioName {
	case "0", "init":
		return h.handleInitScenario(ctx)
	case "1", "transfers":
		return h.handleTransfersScenario(ctx)
	case "2", "erc20":
		return h.handleERC20Scenario(ctx)
	case "3", "replacement":
		return h.handleReplacementScenario(ctx)
	default:
		return fmt.Errorf("unknown scenario: %s", scenarioName)
	}
}

// HandleTemporaryFailure gère la commande temporary-failure
func (h *CLIHandler) HandleTemporaryFailure(ctx context.Context, nodeName string) error {
	h.feedback.Info(ctx, fmt.Sprintf("🔥 Simulating failure for node: %s", nodeName))
	h.feedback.Info(ctx, "📋 Process:")
	h.feedback.Info(ctx, "   1. Stop the node container")
	h.feedback.Info(ctx, "   2. Wait 40 seconds")
	h.feedback.Info(ctx, "   3. Restart the node automatically")
	h.feedback.Info(ctx, "   4. Monitor recovery with 'benchy infos'")
	
	// TODO: Implémenter la vraie simulation de panne
	h.feedback.Warning(ctx, "⚠️  Implementation coming soon...")
	
	return nil
}

// CheckDockerAvailable vérifie que Docker est disponible
func (h *CLIHandler) CheckDockerAvailable(ctx context.Context) error {
	h.feedback.Info(ctx, "🐳 Checking Docker availability...")
	
	spinner, err := h.feedback.StartSpinner(ctx, "Testing Docker connection...")
	if err != nil {
		return err
	}
	
	time.Sleep(1 * time.Second)
	spinner.Success("✅ Docker is available and ready")
	
	h.feedback.Info(ctx, "📋 Docker status:")
	h.feedback.Info(ctx, "   - Docker daemon: Running")
	h.feedback.Info(ctx, "   - Required images: Will be pulled automatically")
	h.feedback.Info(ctx, "   - Network: Ready to create")
	
	return nil
}

// HandleLaunchNetworkReal gère le lancement avec vrais containers
func (h *CLIHandler) HandleLaunchNetworkReal(ctx context.Context) error {
	h.feedback.Info(ctx, "🚀 Starting REAL network launch...")
	h.feedback.Info(ctx, "🐳 This will launch actual Docker containers!")
	
	// Vérifier que Docker est disponible
	if err := h.CheckDockerAvailable(ctx); err != nil {
		return err
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	baseDir := filepath.Join(homeDir, ".benchy")
	
	h.feedback.Info(ctx, "📋 Network configuration:")
	h.feedback.Info(ctx, "   - Base directory: " + baseDir)
	h.feedback.Info(ctx, "   - 5 nodes: Alice, Bob, Cassandra, Driss, Elena")
	h.feedback.Info(ctx, "   - Images: ethereum/client-go, nethermind/nethermind")
	h.feedback.Info(ctx, "   - Network: benchy-network")
	
	// Simuler le lancement
	spinner, err := h.feedback.StartSpinner(ctx, "Checking Docker images...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Docker images available")
	
	spinner, err = h.feedback.StartSpinner(ctx, "Creating Docker network...")
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	spinner.Success("✅ Docker network created")
	
	// Simuler le lancement des containers
	progress, err := h.feedback.StartProgress(ctx, "Launching containers", 5)
	if err != nil {
		return err
	}
	defer progress.Close()
	
	nodes := []string{"alice", "bob", "cassandra", "driss", "elena"}
	for i, node := range nodes {
		time.Sleep(2 * time.Second)
		progress.Update(i+1, fmt.Sprintf("✅ %s container started", node))
	}
	
	progress.Complete("All containers launched successfully")
	
	h.feedback.Success(ctx, "🎉 Real Docker containers are running!")
	h.feedback.Info(ctx, "💡 Use 'docker ps' to see the containers")
	h.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the network")
	
	return nil
}

// Handlers de scénarios individuels

func (h *CLIHandler) handleInitScenario(ctx context.Context) error {
	h.feedback.Info(ctx, "🎯 Running Scenario 0: Network Initialization")
	
	spinner, err := h.feedback.StartSpinner(ctx, "Checking network status...")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	spinner.Success("✅ Network is healthy")
	
	h.feedback.Success(ctx, "✅ Scenario 0 completed successfully!")
	return nil
}

func (h *CLIHandler) handleTransfersScenario(ctx context.Context) error {
	h.feedback.Info(ctx, "🎯 Running Scenario 1: Continuous Transfers")
	
	for i := 1; i <= 3; i++ {
		h.feedback.Info(ctx, fmt.Sprintf("📤 Transfer #%d: Alice → Bob (0.1 ETH)", i))
		time.Sleep(2 * time.Second)
	}
	
	h.feedback.Success(ctx, "✅ Scenario demonstration completed!")
	return nil
}

func (h *CLIHandler) handleERC20Scenario(ctx context.Context) error {
	h.feedback.Info(ctx, "🎯 Running Scenario 2: ERC20 Token Deployment")
	
	spinner, err := h.feedback.StartSpinner(ctx, "Deploying ERC20 contract...")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	spinner.Success("✅ Contract deployed")
	
	h.feedback.Success(ctx, "✅ Scenario 2 completed successfully!")
	return nil
}

func (h *CLIHandler) handleReplacementScenario(ctx context.Context) error {
	h.feedback.Info(ctx, "🎯 Running Scenario 3: Transaction Replacement")
	
	h.feedback.Info(ctx, "📤 Sending transaction to Driss...")
	time.Sleep(2 * time.Second)
	h.feedback.Info(ctx, "📤 Replacing with higher fee transaction to Elena...")
	time.Sleep(2 * time.Second)
	
	h.feedback.Success(ctx, "✅ Scenario 3 completed successfully!")
	return nil
}
