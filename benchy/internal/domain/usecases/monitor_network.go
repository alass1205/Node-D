package usecases

import (
	"context"
	"fmt"
	"time"
	"math/big"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// MonitorNetworkUseCase gère le monitoring du réseau
type MonitorNetworkUseCase struct {
	networkRepo     ports.NetworkRepository
	dockerService   ports.DockerService
	ethService      ports.EthereumService
	monitoringService ports.MonitoringService
	feedback        ports.FeedbackService
}

// NewMonitorNetworkUseCase crée une nouvelle instance
func NewMonitorNetworkUseCase(
	networkRepo ports.NetworkRepository,
	dockerService ports.DockerService,
	ethService ports.EthereumService,
	monitoringService ports.MonitoringService,
	feedback ports.FeedbackService,
) *MonitorNetworkUseCase {
	return &MonitorNetworkUseCase{
		networkRepo:     networkRepo,
		dockerService:   dockerService,
		ethService:      ethService,
		monitoringService: monitoringService,
		feedback:        feedback,
	}
}

// Execute affiche les informations du réseau
func (uc *MonitorNetworkUseCase) Execute(ctx context.Context, updateInterval int) error {
	// Récupérer le réseau
	network, err := uc.networkRepo.GetNetwork(ctx, "benchy-network")
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	
	if updateInterval > 0 {
		// Mode monitoring continu
		return uc.continuousMonitoring(ctx, network, updateInterval)
	}
	
	// Mode one-shot
	return uc.displayNetworkInfo(ctx, network)
}

// continuousMonitoring affiche les infos en continu
func (uc *MonitorNetworkUseCase) continuousMonitoring(ctx context.Context, network *entities.Network, interval int) error {
	uc.feedback.Info(ctx, fmt.Sprintf("📊 Monitoring nodes (updating every %d seconds, press Ctrl+C to stop)", interval))
	
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Clear screen
			fmt.Print("\033[2J\033[H")
			uc.feedback.Info(ctx, fmt.Sprintf("📊 Network Information (Last update: %s)", time.Now().Format("15:04:05")))
			
			if err := uc.displayNetworkInfo(ctx, network); err != nil {
				uc.feedback.Error(ctx, fmt.Sprintf("Error updating info: %v", err))
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// displayNetworkInfo affiche les informations du réseau
func (uc *MonitorNetworkUseCase) displayNetworkInfo(ctx context.Context, network *entities.Network) error {
	// Récupérer les métriques de chaque node
	var tableData [][]string
	headers := []string{"Node", "Status", "Latest Block", "Peers", "CPU/Memory", "ETH Balance", "Mempool"}
	
	for _, node := range network.Nodes {
		nodeInfo, err := uc.getNodeInfo(ctx, node)
		if err != nil {
			// Node offline ou erreur
			tableData = append(tableData, []string{
				node.Name,
				"❌ Offline",
				"N/A",
				"N/A",
				"N/A",
				"N/A",
				"N/A",
			})
			continue
		}
		
		row := []string{
			node.Name,
			nodeInfo.StatusDisplay,
			fmt.Sprintf("%d", nodeInfo.LatestBlock),
			fmt.Sprintf("%d", nodeInfo.PeerCount),
			fmt.Sprintf("%.1f%%/%.0fMB", nodeInfo.CPUUsage, nodeInfo.MemoryUsage),
			fmt.Sprintf("%.2f ETH", nodeInfo.ETHBalance),
			fmt.Sprintf("%d", nodeInfo.PendingTxs),
		}
		
		tableData = append(tableData, row)
	}
	
	// Afficher le tableau
	if err := uc.feedback.DisplayTable(ctx, headers, tableData); err != nil {
		return err
	}
	
	// Afficher les informations réseau
	networkMetrics, err := uc.monitoringService.GetNetworkMetrics(ctx, network.Name)
	if err == nil {
		uc.feedback.Info(ctx, fmt.Sprintf("Total pending transactions: %d", networkMetrics.TotalTxs))
		uc.feedback.Info(ctx, fmt.Sprintf("Network health: %s", uc.getNetworkHealthStatus(network)))
	}
	
	return nil
}

// NodeInfo représente les informations d'un node
type NodeInfo struct {
	Name          string
	StatusDisplay string
	LatestBlock   uint64
	PeerCount     int
	CPUUsage      float64
	MemoryUsage   float64
	ETHBalance    float64
	PendingTxs    int
}

// getNodeInfo récupère les informations d'un node
func (uc *MonitorNetworkUseCase) getNodeInfo(ctx context.Context, node *entities.Node) (*NodeInfo, error) {
	info := &NodeInfo{
		Name: node.Name,
	}
	
	// Vérifier si le container est en cours d'exécution
	if node.ContainerID != "" {
		running, err := uc.dockerService.IsContainerRunning(ctx, node.ContainerID)
		if err != nil || !running {
			info.StatusDisplay = "❌ Offline"
			return info, fmt.Errorf("container not running")
		}
		
		// Récupérer les stats du container
		stats, err := uc.dockerService.GetContainerStats(ctx, node.ContainerID)
		if err == nil {
			info.CPUUsage = stats.CPUUsage
			info.MemoryUsage = float64(stats.MemoryUsage) / 1024 / 1024 // MB
		}
	}
	
	// Se connecter au node Ethereum
	nodeURL := fmt.Sprintf("http://localhost:%d", node.RPCPort)
	if err := uc.ethService.ConnectToNode(ctx, nodeURL); err != nil {
		info.StatusDisplay = "❌ Offline"
		return info, fmt.Errorf("failed to connect to ethereum node: %w", err)
	}
	
	// Récupérer les métriques blockchain
	latestBlock, err := uc.ethService.GetLatestBlockNumber(ctx, nodeURL)
	if err == nil {
		info.LatestBlock = latestBlock
	}
	
	peerCount, err := uc.ethService.GetPeerCount(ctx, nodeURL)
	if err == nil {
		info.PeerCount = peerCount
	}
	
	pendingTxs, err := uc.ethService.GetPendingTransactionCount(ctx, nodeURL)
	if err == nil {
		info.PendingTxs = pendingTxs
	}
	
	// Récupérer la balance ETH
	balance, err := uc.ethService.GetBalance(ctx, nodeURL, node.Address)
	if err == nil {
		// Convertir wei en ETH
		ethBalance := new(big.Float).SetInt(balance)
		ethBalance.Quo(ethBalance, big.NewFloat(1e18))
		info.ETHBalance, _ = ethBalance.Float64()
	}
	
	// Déterminer le status d'affichage
	if info.PeerCount > 0 {
		info.StatusDisplay = "✅ Online"
	} else {
		info.StatusDisplay = "🔄 Syncing"
	}
	
	return info, nil
}

// getNetworkHealthStatus retourne le statut de santé du réseau
func (uc *MonitorNetworkUseCase) getNetworkHealthStatus(network *entities.Network) string {
	if network.IsHealthy() {
		return "✅ Healthy"
	}
	return "⚠️ Unhealthy"
}
