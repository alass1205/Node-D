package ports

import (
	"context"
	"time"
	"benchy/internal/domain/entities"
)

// MonitoringService définit les opérations de monitoring
type MonitoringService interface {
	// Monitoring des nodes
	StartMonitoring(ctx context.Context, network *entities.Network) error
	StopMonitoring(ctx context.Context, networkName string) error
	GetNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error)
	GetNetworkMetrics(ctx context.Context, networkName string) (*NetworkMetrics, error)
	
	// Alertes
	RegisterAlert(ctx context.Context, alert *Alert) error
	GetActiveAlerts(ctx context.Context, networkName string) ([]*Alert, error)
	
	// Historique
	GetMetricsHistory(ctx context.Context, nodeName string, duration time.Duration) ([]*NodeMetrics, error)
	
	// Health checks
	CheckNodeHealth(ctx context.Context, node *entities.Node) (*HealthStatus, error)
	CheckNetworkHealth(ctx context.Context, network *entities.Network) (*HealthStatus, error)
}

// NodeMetrics représente les métriques d'un node
type NodeMetrics struct {
	NodeName    string
	Timestamp   time.Time
	
	// Métriques système
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	NetworkIO   NetworkIOMetrics
	
	// Métriques blockchain
	LatestBlock     uint64
	ConnectedPeers  int
	PendingTxs      int
	SyncStatus      string
	
	// Métriques performance
	BlockTime       time.Duration
	TxThroughput    float64
	ResponseTime    time.Duration
	
	// Status
	IsOnline        bool
	LastSeen        time.Time
}

// NetworkMetrics représente les métriques du réseau
type NetworkMetrics struct {
	NetworkName     string
	Timestamp       time.Time
	
	// Métriques générales
	TotalNodes      int
	OnlineNodes     int
	ValidatorNodes  int
	LatestBlock     uint64
	TotalTxs        uint64
	
	// Performance
	AvgBlockTime    time.Duration
	TxThroughput    float64
	NetworkLatency  time.Duration
	
	// Consensus
	ConsensusStatus string
	MissedBlocks    int
	ForkCount       int
}

// NetworkIOMetrics représente les métriques réseau
type NetworkIOMetrics struct {
	BytesReceived uint64
	BytesSent     uint64
	PacketsReceived uint64
	PacketsSent   uint64
}

// Alert représente une alerte système
type Alert struct {
	ID          string
	Type        AlertType
	Severity    AlertSeverity
	NodeName    string
	Message     string
	Timestamp   time.Time
	Resolved    bool
	ResolvedAt  time.Time
}

// AlertType représente le type d'alerte
type AlertType string

const (
	AlertTypeNodeDown     AlertType = "node_down"
	AlertTypeHighCPU      AlertType = "high_cpu"
	AlertTypeHighMemory   AlertType = "high_memory"
	AlertTypeSyncIssue    AlertType = "sync_issue"
	AlertTypeNetworkSplit AlertType = "network_split"
)

// AlertSeverity représente la sévérité d'une alerte
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// HealthStatus représente le statut de santé
type HealthStatus struct {
	IsHealthy   bool
	Score       float64 // 0-100
	Issues      []string
	Timestamp   time.Time
	Checks      map[string]bool
}
