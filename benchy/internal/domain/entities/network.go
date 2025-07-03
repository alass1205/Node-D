package entities

import (
	"math/big"
	"time"
)

// NetworkStatus représente l'état du réseau
type NetworkStatus string

const (
	NetworkStatusStopped  NetworkStatus = "stopped"
	NetworkStatusStarting NetworkStatus = "starting"
	NetworkStatusRunning  NetworkStatus = "running"
	NetworkStatusStopping NetworkStatus = "stopping"
)

// Network représente notre réseau Ethereum privé
type Network struct {
	Name      string        `json:"name"`
	ChainID   *big.Int      `json:"chain_id"`
	Consensus string        `json:"consensus"` // "clique"
	Status    NetworkStatus `json:"status"`
	
	// Configuration
	BlockTime    time.Duration `json:"block_time"`
	EpochLength  uint64        `json:"epoch_length"`
	NetworkID    string        `json:"network_id"`
	
	// Nodes
	Nodes      []*Node `json:"nodes"`
	Validators []*Node `json:"validators"`
	
	// Métriques réseau
	TotalNodes     int     `json:"total_nodes"`
	OnlineNodes    int     `json:"online_nodes"`
	LatestBlock    uint64  `json:"latest_block"`
	TotalTxs       uint64  `json:"total_txs"`
	NetworkHashrate float64 `json:"network_hashrate"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at"`
}

// NewNetwork crée un nouveau réseau avec la configuration par défaut
func NewNetwork(name string, chainID *big.Int) *Network {
	return &Network{
		Name:        name,
		ChainID:     chainID,
		Consensus:   "clique",
		Status:      NetworkStatusStopped,
		BlockTime:   5 * time.Second,
		EpochLength: 30000,
		NetworkID:   "benchy-network",
		Nodes:       make([]*Node, 0),
		Validators:  make([]*Node, 0),
		CreatedAt:   time.Now(),
	}
}

// AddNode ajoute un node au réseau
func (n *Network) AddNode(node *Node) {
	n.Nodes = append(n.Nodes, node)
	if node.IsValidator {
		n.Validators = append(n.Validators, node)
	}
	n.TotalNodes = len(n.Nodes)
}

// GetNodeByName retourne un node par son nom
func (n *Network) GetNodeByName(name string) *Node {
	for _, node := range n.Nodes {
		if node.Name == name {
			return node
		}
	}
	return nil
}

// GetOnlineNodes retourne le nombre de nodes en ligne
func (n *Network) GetOnlineNodes() int {
	count := 0
	for _, node := range n.Nodes {
		if node.IsOnline() {
			count++
		}
	}
	return count
}

// IsHealthy retourne true si le réseau est en bonne santé
func (n *Network) IsHealthy() bool {
	if n.Status != NetworkStatusRunning {
		return false
	}
	
	// Au moins 2 validateurs doivent être en ligne
	onlineValidators := 0
	for _, validator := range n.Validators {
		if validator.IsOnline() {
			onlineValidators++
		}
	}
	
	return onlineValidators >= 2
}
