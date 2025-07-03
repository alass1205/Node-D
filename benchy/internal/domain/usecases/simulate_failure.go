package usecases

import (
	"context"
	"fmt"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// SimulateFailureUseCase gère la simulation de pannes
type SimulateFailureUseCase struct {
	networkRepo   ports.NetworkRepository
	dockerService ports.DockerService
	feedback      ports.FeedbackService
}

// NewSimulateFailureUseCase crée une nouvelle instance
func NewSimulateFailureUseCase(
	networkRepo ports.NetworkRepository,
	dockerService ports.DockerService,
	feedback ports.FeedbackService,
) *SimulateFailureUseCase {
	return &SimulateFailureUseCase{
		networkRepo:   networkRepo,
		dockerService: dockerService,
		feedback:      feedback,
	}
}

// Execute simule une panne temporaire du node spécifié
func (uc *SimulateFailureUseCase) Execute(ctx context.Context, nodeName string) error {
	// Récupérer le réseau
	network, err := uc.networkRepo.GetNetwork(ctx, "benchy-network")
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	
	// Trouver le node
	node := network.GetNodeByName(nodeName)
	if node == nil {
		return fmt.Errorf("node %s not found", nodeName)
	}
	
	uc.feedback.Info(ctx, fmt.Sprintf("🔥 Simulating failure for node: %s", nodeName))
	uc.feedback.Info(ctx, "📋 Process:")
	uc.feedback.Info(ctx, "   1. Stop the node immediately")
	uc.feedback.Info(ctx, "   2. Wait 40 seconds")
	uc.feedback.Info(ctx, "   3. Restart the node automatically")
	uc.feedback.Info(ctx, "   4. Monitor recovery with 'benchy infos'")
	
	// Vérifier que le node est actuellement en ligne
	if node.ContainerID == "" {
		return fmt.Errorf("node %s has no container ID", nodeName)
	}
	
	running, err := uc.dockerService.IsContainerRunning(ctx, node.ContainerID)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	
	if !running {
		return fmt.Errorf("node %s is not currently running", nodeName)
	}
	
	// 1. Arrêter le node
	uc.feedback.Info(ctx, fmt.Sprintf("🛑 Stopping node %s...", nodeName))
	
	if err := uc.dockerService.StopContainer(ctx, node.ContainerID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	
	// Mettre à jour le statut
	node.Status = entities.StatusOffline
	if err := uc.networkRepo.UpdateNode(ctx, network.Name, node); err != nil {
		uc.feedback.Warning(ctx, fmt.Sprintf("Failed to update node status: %v", err))
	}
	
	uc.feedback.Success(ctx, fmt.Sprintf("✅ Node %s stopped", nodeName))
	
	// 2. Attendre 40 secondes avec compteur
	uc.feedback.Info(ctx, "⏳ Waiting 40 seconds before restart...")
	
	// Créer un progress tracker pour le countdown
	progress, err := uc.feedback.StartProgress(ctx, "Waiting for restart", 40)
	if err != nil {
		return err
	}
	
	for i := 1; i <= 40; i++ {
		select {
		case <-time.After(1 * time.Second):
			progress.Update(i, fmt.Sprintf("Waiting... %d/40 seconds", i))
		case <-ctx.Done():
			progress.Close()
			return ctx.Err()
		}
	}
	
	progress.Complete("⏰ 40 seconds elapsed")
	
	// 3. Redémarrer le node
	uc.feedback.Info(ctx, fmt.Sprintf("🔄 Restarting node %s...", nodeName))
	
	spinner, err := uc.feedback.StartSpinner(ctx, "Starting container...")
	if err != nil {
		return err
	}
	
	if err := uc.dockerService.StartContainer(ctx, node.ContainerID); err != nil {
		spinner.Error(fmt.Sprintf("Failed to restart container: %v", err))
		return fmt.Errorf("failed to restart container: %w", err)
	}
	
	spinner.Success("✅ Container restarted")
	
	// 4. Attendre que le node soit prêt
	if err := uc.waitForNodeRecovery(ctx, node); err != nil {
		return fmt.Errorf("node failed to recover: %w", err)
	}
	
	// Mettre à jour le statut
	node.Status = entities.StatusOnline
	if err := uc.networkRepo.UpdateNode(ctx, network.Name, node); err != nil {
		uc.feedback.Warning(ctx, fmt.Sprintf("Failed to update node status: %v", err))
	}
	
	uc.feedback.Success(ctx, fmt.Sprintf("✅ Node %s recovered successfully!", nodeName))
	uc.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the node synchronization")
	uc.feedback.Info(ctx, "💡 The node should be fully synchronized in a few minutes")
	
	return nil
}

// waitForNodeRecovery attend que le node soit de nouveau opérationnel
func (uc *SimulateFailureUseCase) waitForNodeRecovery(ctx context.Context, node *entities.Node) error {
	spinner, err := uc.feedback.StartSpinner(ctx, "Waiting for node to become ready...")
	if err != nil {
		return err
	}
	
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			spinner.Error("Node failed to recover within timeout")
			return fmt.Errorf("node %s failed to recover within 60 seconds", node.Name)
		case <-ticker.C:
			// Vérifier si le container est en cours d'exécution
			running, err := uc.dockerService.IsContainerRunning(ctx, node.ContainerID)
			if err != nil {
				continue
			}
			
			if running {
				spinner.Success("✅ Node is ready")
				return nil
			}
		case <-ctx.Done():
			spinner.Error("Recovery cancelled")
			return ctx.Err()
		}
	}
}
