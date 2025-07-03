package services

import (
	"context"
	"fmt"
	"math/big"
	"time"
	"github.com/ethereum/go-ethereum/common"

	"benchy/internal/infrastructure/docker"
	"benchy/internal/infrastructure/ethereum"
	"benchy/internal/infrastructure/feedback"
	"benchy/internal/infrastructure/monitoring"
)

// MonitoringService orchestre le monitoring complet du réseau
type MonitoringService struct {
	dockerClient *docker.DockerClient
	ethClient    *ethereum.EthereumClient
	systemMonitor *monitoring.SystemMonitor
	feedback     *feedback.ConsoleFeedback
}

// NewMonitoringService crée un nouveau service de monitoring
func NewMonitoringService() (*MonitoringService, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &MonitoringService{
		dockerClient:  dockerClient,
		ethClient:     ethereum.NewEthereumClient(),
		systemMonitor: monitoring.NewSystemMonitor(),
		feedback:      feedback.NewConsoleFeedback(),
	}, nil
}

// DisplayNetworkInfo affiche les informations complètes du réseau
func (ms *MonitoringService) DisplayNetworkInfo(ctx context.Context, updateInterval int) error {
	if updateInterval > 0 {
		return ms.continuousMonitoring(ctx, updateInterval)
	}
	
	return ms.displayOneShotInfo(ctx)
}

// continuousMonitoring affiche les infos en continu
func (ms *MonitoringService) continuousMonitoring(ctx context.Context, interval int) error {
	ms.feedback.Info(ctx, fmt.Sprintf("📊 Monitoring nodes (updating every %d seconds, press Ctrl+C to stop)", interval))

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Première exécution immédiate
	if err := ms.displayOneShotInfo(ctx); err != nil {
		ms.feedback.Error(ctx, fmt.Sprintf("Error: %v", err))
	}

	for {
		select {
		case <-ticker.C:
			// Clear screen et afficher timestamp
			fmt.Print("\033[2J\033[H")
			ms.feedback.Info(ctx, fmt.Sprintf("📊 Network Information (Last update: %s)", time.Now().Format("15:04:05")))
			fmt.Println()

			if err := ms.displayOneShotInfo(ctx); err != nil {
				ms.feedback.Error(ctx, fmt.Sprintf("Error updating info: %v", err))
			}
		case <-ctx.Done():
			ms.feedback.Info(ctx, "🔄 Stopping monitoring...")
			return ctx.Err()
		}
	}
}

// displayOneShotInfo affiche les infos une seule fois
func (ms *MonitoringService) displayOneShotInfo(ctx context.Context) error {
	// Récupérer les containers benchy
	containers, err := ms.getBenchyContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get containers: %w", err)
	}

	if len(containers) == 0 {
		ms.feedback.Warning(ctx, "⚠️  No benchy containers found. Did you run 'benchy launch-network'?")
		return nil
	}

	// Préparer les données du tableau
	headers := []string{"Node", "Status", "Latest Block", "Peers", "CPU/Memory", "ETH Balance", "Container"}
	var rows [][]string

	for _, container := range containers {
		nodeInfo, err := ms.getNodeInfo(ctx, container)
		if err != nil {
			// Node offline ou erreur
			rows = append(rows, []string{
				container.NodeName,
				"❌ Offline",
				"N/A",
				"N/A",
				"N/A",
				"N/A",
				container.ID[:12],
			})
			continue
		}

		row := []string{
			nodeInfo.Name,
			nodeInfo.StatusDisplay,
			fmt.Sprintf("%d", nodeInfo.LatestBlock),
			fmt.Sprintf("%d", nodeInfo.PeerCount),
			fmt.Sprintf("%.1f%%/%.0fMB", nodeInfo.CPUUsage, nodeInfo.MemoryUsage),
			fmt.Sprintf("%.2f ETH", nodeInfo.ETHBalance),
			container.ID[:12],
		}

		rows = append(rows, row)
	}

	// Afficher le tableau
	if err := ms.feedback.DisplayTable(ctx, headers, rows); err != nil {
		return fmt.Errorf("failed to display table: %w", err)
	}

	// Afficher les informations réseau supplémentaires
	ms.displayNetworkSummary(ctx, containers)

	return nil
}

// getBenchyContainers récupère tous les containers benchy
func (ms *MonitoringService) getBenchyContainers(ctx context.Context) ([]*ContainerInfo, error) {
	// Pour l'instant, on simule la récupération des containers
	// TODO: Implémenter la vraie récupération via Docker API
	
	// Simuler 5 containers pour la démonstration
	containers := []*ContainerInfo{
		{ID: "abc123456789", NodeName: "alice", Status: "running"},
		{ID: "def123456789", NodeName: "bob", Status: "running"},
		{ID: "ghi123456789", NodeName: "cassandra", Status: "running"},
		{ID: "jkl123456789", NodeName: "driss", Status: "running"},
		{ID: "mno123456789", NodeName: "elena", Status: "running"},
	}

	return containers, nil
}

// ContainerInfo représente les infos d'un container benchy
type ContainerInfo struct {
	ID       string
	NodeName string
	Status   string
	Port     int
	RPCPort  int
}

// NodeInfo représente les informations complètes d'un node
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

// getNodeInfo récupère les informations complètes d'un node
func (ms *MonitoringService) getNodeInfo(ctx context.Context, container *ContainerInfo) (*NodeInfo, error) {
	info := &NodeInfo{
		Name: container.NodeName,
	}

	// 1. Vérifier le status du container
	if container.Status != "running" {
		info.StatusDisplay = "❌ Offline"
		return info, fmt.Errorf("container not running")
	}

	// 2. Récupérer les stats Docker (CPU/RAM)
	stats, err := ms.getContainerStats(ctx, container.ID)
	if err == nil {
		info.CPUUsage = stats.CPUUsage
		info.MemoryUsage = stats.MemoryUsage
	}

	// 3. Essayer de se connecter au node Ethereum
	nodeURL := fmt.Sprintf("http://localhost:%d", ms.getNodeRPCPort(container.NodeName))
	
	if err := ms.ethClient.ConnectToNode(ctx, nodeURL); err != nil {
		info.StatusDisplay = "🔄 Starting"
		// Pas encore prêt, mais container en cours
		return info, nil
	}

	// 4. Récupérer les métriques blockchain
	if latestBlock, err := ms.ethClient.GetLatestBlockNumber(ctx, nodeURL); err == nil {
		info.LatestBlock = latestBlock
	}

	if peerCount, err := ms.ethClient.GetPeerCount(ctx, nodeURL); err == nil {
		info.PeerCount = peerCount
	}

	if pendingTxs, err := ms.ethClient.GetPendingTransactionCount(ctx, nodeURL); err == nil {
		info.PendingTxs = pendingTxs
	}

	// 5. Récupérer la balance ETH
	address := ms.getNodeAddress(container.NodeName)
	if balance, err := ms.ethClient.GetBalance(ctx, nodeURL, address); err == nil {
		ethBalance := new(big.Float).SetInt(balance)
		ethBalance.Quo(ethBalance, big.NewFloat(1e18))
		info.ETHBalance, _ = ethBalance.Float64()
	}

	// 6. Déterminer le status d'affichage final
	if info.PeerCount > 0 {
		info.StatusDisplay = "✅ Online"
	} else if info.LatestBlock > 0 {
		info.StatusDisplay = "🔄 Syncing"
	} else {
		info.StatusDisplay = "⏳ Starting"
	}

	return info, nil
}

// getContainerStats récupère les stats d'un container (simulation)
func (ms *MonitoringService) getContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// Pour l'instant, on simule les stats
	// TODO: Utiliser la vraie API Docker
	return &ContainerStats{
		CPUUsage:    float64(20 + (len(containerID) % 30)), // 20-50%
		MemoryUsage: float64(100 + (len(containerID) % 100)), // 100-200MB
	}, nil
}

// getNodeRPCPort retourne le port RPC d'un node par son nom
func (ms *MonitoringService) getNodeRPCPort(nodeName string) int {
	ports := map[string]int{
		"alice":     8545,
		"bob":       8546,
		"cassandra": 8547,
		"driss":     8548,
		"elena":     8549,
	}
	
	if port, exists := ports[nodeName]; exists {
		return port
	}
	return 8545 // Défaut
}

// getNodeAddress retourne l'adresse Ethereum d'un node (simulation)
func (ms *MonitoringService) getNodeAddress(nodeName string) common.Address {
	// Pour l'instant, on utilise des adresses fictives
	// TODO: Récupérer les vraies adresses depuis la configuration
	addresses := map[string]string{
		"alice":     "0x1111111111111111111111111111111111111111",
		"bob":       "0x2222222222222222222222222222222222222222",
		"cassandra": "0x3333333333333333333333333333333333333333",
		"driss":     "0x4444444444444444444444444444444444444444",
		"elena":     "0x5555555555555555555555555555555555555555",
	}
	
	if addr, exists := addresses[nodeName]; exists {
		return common.HexToAddress(addr)
	}
	return common.Address{} // Adresse vide
}

// displayNetworkSummary affiche un résumé du réseau
func (ms *MonitoringService) displayNetworkSummary(ctx context.Context, containers []*ContainerInfo) {
	fmt.Println()
	
	onlineCount := 0
	for _, container := range containers {
		if container.Status == "running" {
			onlineCount++
		}
	}
	
	ms.feedback.Info(ctx, fmt.Sprintf("📈 Network Summary:"))
	ms.feedback.Info(ctx, fmt.Sprintf("   • Total nodes: %d", len(containers)))
	ms.feedback.Info(ctx, fmt.Sprintf("   • Online nodes: %d", onlineCount))
	ms.feedback.Info(ctx, fmt.Sprintf("   • Validators: 3 (Alice, Bob, Cassandra)"))
	ms.feedback.Info(ctx, fmt.Sprintf("   • Consensus: Clique (5s blocks)"))
	
	if onlineCount < len(containers) {
		ms.feedback.Warning(ctx, fmt.Sprintf("⚠️  %d nodes are offline", len(containers)-onlineCount))
	} else {
		ms.feedback.Success(ctx, "✅ All nodes are online")
	}
}

// ContainerStats représente les stats d'un container
type ContainerStats struct {
	CPUUsage    float64
	MemoryUsage float64
}
