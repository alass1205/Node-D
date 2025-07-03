package monitoring

import (
	"context"
	"fmt"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMonitor implémente l'interface MonitoringService
type SystemMonitor struct {
	nodeMetrics map[string]*ports.NodeMetrics
	alerts      map[string][]*ports.Alert
}

// NewSystemMonitor crée un nouveau moniteur système
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		nodeMetrics: make(map[string]*ports.NodeMetrics),
		alerts:      make(map[string][]*ports.Alert),
	}
}

// StartMonitoring démarre le monitoring d'un réseau
func (sm *SystemMonitor) StartMonitoring(ctx context.Context, network *entities.Network) error {
	// Pour l'instant, on initialise juste le monitoring
	// TODO: Implémenter une goroutine de monitoring continu
	
	for _, node := range network.Nodes {
		sm.nodeMetrics[node.Name] = &ports.NodeMetrics{
			NodeName:  node.Name,
			Timestamp: time.Now(),
			IsOnline:  false,
		}
	}

	return nil
}

// StopMonitoring arrête le monitoring d'un réseau
func (sm *SystemMonitor) StopMonitoring(ctx context.Context, networkName string) error {
	// Nettoyer les métriques
	// TODO: Implémenter l'arrêt propre du monitoring
	return nil
}

// GetNodeMetrics récupère les métriques d'un node
func (sm *SystemMonitor) GetNodeMetrics(ctx context.Context, nodeName string) (*ports.NodeMetrics, error) {
	// Récupérer les métriques système actuelles
	cpuUsage, err := sm.getCPUUsage()
	if err != nil {
		cpuUsage = 0.0
	}

	memoryUsage, err := sm.getMemoryUsage()
	if err != nil {
		memoryUsage = 0.0
	}

	diskUsage, err := sm.getDiskUsage()
	if err != nil {
		diskUsage = 0.0
	}

	networkIO, err := sm.getNetworkIO()
	if err != nil {
		networkIO = ports.NetworkIOMetrics{}
	}

	metrics := &ports.NodeMetrics{
		NodeName:    nodeName,
		Timestamp:   time.Now(),
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		DiskUsage:   diskUsage,
		NetworkIO:   networkIO,
		IsOnline:    true,
		LastSeen:    time.Now(),
		
		// Métriques blockchain (seront mises à jour par d'autres services)
		LatestBlock:    0,
		ConnectedPeers: 0,
		PendingTxs:     0,
		SyncStatus:     "unknown",
		
		// Métriques performance
		BlockTime:    5 * time.Second,
		TxThroughput: 0.0,
		ResponseTime: 100 * time.Millisecond,
	}

	// Stocker dans le cache
	sm.nodeMetrics[nodeName] = metrics

	return metrics, nil
}

// GetNetworkMetrics récupère les métriques du réseau
func (sm *SystemMonitor) GetNetworkMetrics(ctx context.Context, networkName string) (*ports.NetworkMetrics, error) {
	// Calculer les métriques agrégées depuis les nodes
	totalNodes := len(sm.nodeMetrics)
	onlineNodes := 0
	var latestBlock uint64 = 0
	var totalTxs uint64 = 0

	for _, metrics := range sm.nodeMetrics {
		if metrics.IsOnline {
			onlineNodes++
		}
		if metrics.LatestBlock > latestBlock {
			latestBlock = metrics.LatestBlock
		}
	}

	networkMetrics := &ports.NetworkMetrics{
		NetworkName:     networkName,
		Timestamp:       time.Now(),
		TotalNodes:      totalNodes,
		OnlineNodes:     onlineNodes,
		ValidatorNodes:  3, // Alice, Bob, Cassandra
		LatestBlock:     latestBlock,
		TotalTxs:        totalTxs,
		AvgBlockTime:    5 * time.Second,
		TxThroughput:    0.0,
		NetworkLatency:  50 * time.Millisecond,
		ConsensusStatus: "healthy",
		MissedBlocks:    0,
		ForkCount:       0,
	}

	return networkMetrics, nil
}

// RegisterAlert enregistre une alerte
func (sm *SystemMonitor) RegisterAlert(ctx context.Context, alert *ports.Alert) error {
	networkKey := "default" // Pour l'instant, on utilise une clé par défaut
	
	if sm.alerts[networkKey] == nil {
		sm.alerts[networkKey] = make([]*ports.Alert, 0)
	}
	
	sm.alerts[networkKey] = append(sm.alerts[networkKey], alert)
	return nil
}

// GetActiveAlerts récupère les alertes actives
func (sm *SystemMonitor) GetActiveAlerts(ctx context.Context, networkName string) ([]*ports.Alert, error) {
	alerts := sm.alerts[networkName]
	if alerts == nil {
		return []*ports.Alert{}, nil
	}

	// Filtrer les alertes non résolues
	var activeAlerts []*ports.Alert
	for _, alert := range alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts, nil
}

// GetMetricsHistory récupère l'historique des métriques
func (sm *SystemMonitor) GetMetricsHistory(ctx context.Context, nodeName string, duration time.Duration) ([]*ports.NodeMetrics, error) {
	// Pour l'instant, on retourne juste les métriques actuelles
	// TODO: Implémenter un vrai système d'historique
	
	currentMetrics, err := sm.GetNodeMetrics(ctx, nodeName)
	if err != nil {
		return nil, err
	}

	return []*ports.NodeMetrics{currentMetrics}, nil
}

// CheckNodeHealth vérifie la santé d'un node
func (sm *SystemMonitor) CheckNodeHealth(ctx context.Context, node *entities.Node) (*ports.HealthStatus, error) {
	checks := make(map[string]bool)
	issues := make([]string, 0)
	score := 100.0

	// Vérifier si le node est en ligne
	checks["online"] = node.IsOnline()
	if !node.IsOnline() {
		issues = append(issues, "Node is offline")
		score -= 50.0
	}

	// Vérifier l'utilisation CPU
	cpuUsage, _ := sm.getCPUUsage()
	checks["cpu_ok"] = cpuUsage < 80.0
	if cpuUsage >= 80.0 {
		issues = append(issues, fmt.Sprintf("High CPU usage: %.1f%%", cpuUsage))
		score -= 20.0
	}

	// Vérifier l'utilisation mémoire
	memUsage, _ := sm.getMemoryUsage()
	checks["memory_ok"] = memUsage < 80.0
	if memUsage >= 80.0 {
		issues = append(issues, fmt.Sprintf("High memory usage: %.1f%%", memUsage))
		score -= 20.0
	}

	// Vérifier la connectivité réseau
	checks["network_ok"] = node.ConnectedPeers > 0
	if node.ConnectedPeers == 0 {
		issues = append(issues, "No connected peers")
		score -= 30.0
	}

	healthStatus := &ports.HealthStatus{
		IsHealthy: len(issues) == 0,
		Score:     score,
		Issues:    issues,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	return healthStatus, nil
}

// CheckNetworkHealth vérifie la santé du réseau
func (sm *SystemMonitor) CheckNetworkHealth(ctx context.Context, network *entities.Network) (*ports.HealthStatus, error) {
	checks := make(map[string]bool)
	issues := make([]string, 0)
	score := 100.0

	// Vérifier le nombre de validateurs en ligne
	onlineValidators := 0
	for _, node := range network.Validators {
		if node.IsOnline() {
			onlineValidators++
		}
	}

	checks["validators_online"] = onlineValidators >= 2
	if onlineValidators < 2 {
		issues = append(issues, fmt.Sprintf("Only %d validators online (minimum: 2)", onlineValidators))
		score -= 40.0
	}

	// Vérifier que le réseau est en cours d'exécution
	checks["network_running"] = network.Status == entities.NetworkStatusRunning
	if network.Status != entities.NetworkStatusRunning {
		issues = append(issues, "Network is not running")
		score -= 60.0
	}

	// Vérifier la synchronisation des blocks
	checks["blocks_synced"] = network.IsHealthy()
	if !network.IsHealthy() {
		issues = append(issues, "Network synchronization issues")
		score -= 30.0
	}

	healthStatus := &ports.HealthStatus{
		IsHealthy: len(issues) == 0,
		Score:     score,
		Issues:    issues,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	return healthStatus, nil
}

// Méthodes utilitaires privées pour récupérer les métriques système

func (sm *SystemMonitor) getCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	
	if len(percentages) > 0 {
		return percentages[0], nil
	}
	
	return 0, nil
}

func (sm *SystemMonitor) getMemoryUsage() (float64, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	
	return vmem.UsedPercent, nil
}

func (sm *SystemMonitor) getDiskUsage() (float64, error) {
	// Pour l'instant, on retourne une valeur fixe
	// TODO: Implémenter la vraie récupération de l'usage disque
	return 45.0, nil
}

func (sm *SystemMonitor) getNetworkIO() (ports.NetworkIOMetrics, error) {
	stats, err := net.IOCounters(false)
	if err != nil {
		return ports.NetworkIOMetrics{}, err
	}
	
	if len(stats) > 0 {
		return ports.NetworkIOMetrics{
			BytesReceived:   stats[0].BytesRecv,
			BytesSent:      stats[0].BytesSent,
			PacketsReceived: stats[0].PacketsRecv,
			PacketsSent:    stats[0].PacketsSent,
		}, nil
	}
	
	return ports.NetworkIOMetrics{}, nil
}
