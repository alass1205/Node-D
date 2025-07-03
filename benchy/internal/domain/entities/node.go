package entities

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// NodeStatus représente l'état d'un node
type NodeStatus string

const (
	StatusOnline    NodeStatus = "online"
	StatusOffline   NodeStatus = "offline"
	StatusSyncing   NodeStatus = "syncing"
	StatusStarting  NodeStatus = "starting"
	StatusStopping  NodeStatus = "stopping"
)

// ClientType représente le type de client Ethereum
type ClientType string

const (
	ClientGeth       ClientType = "geth"
	ClientNethermind ClientType = "nethermind"
)

// Node représente un node Ethereum dans notre réseau
type Node struct {
	Name        string              `json:"name"`
	IsValidator bool                `json:"is_validator"`
	Client      ClientType          `json:"client"`
	PrivateKey  *ecdsa.PrivateKey   `json:"-"` // Ne pas sérialiser
	Address     common.Address      `json:"address"`
	
	// Configuration réseau
	Port        int    `json:"port"`
	RPCPort     int    `json:"rpc_port"`
	ContainerID string `json:"container_id"`
	
	// Status en temps réel
	Status        NodeStatus `json:"status"`
	LatestBlock   uint64     `json:"latest_block"`
	ConnectedPeers int       `json:"connected_peers"`
	PendingTxs    int        `json:"pending_txs"`
	
	// Métriques système
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	
	// Balances
	ETHBalance   *big.Int           `json:"eth_balance"`
	TokenBalance map[string]*big.Int `json:"token_balance"`
	
	// Timestamps
	LastSeen time.Time `json:"last_seen"`
	StartedAt time.Time `json:"started_at"`
}

// NewNode crée un nouveau node avec les paramètres de base
func NewNode(name string, isValidator bool, client ClientType, port, rpcPort int) *Node {
	return &Node{
		Name:         name,
		IsValidator:  isValidator,
		Client:       client,
		Port:         port,
		RPCPort:      rpcPort,
		Status:       StatusOffline,
		ETHBalance:   big.NewInt(0),
		TokenBalance: make(map[string]*big.Int),
		LastSeen:     time.Now(),
	}
}

// IsOnline retourne true si le node est en ligne
func (n *Node) IsOnline() bool {
	return n.Status == StatusOnline || n.Status == StatusSyncing
}

// GetDisplayName retourne le nom formaté pour l'affichage
func (n *Node) GetDisplayName() string {
	if n.IsValidator {
		return n.Name + " (validator)"
	}
	return n.Name
}

// GetStatusEmoji retourne l'emoji correspondant au status
func (n *Node) GetStatusEmoji() string {
	switch n.Status {
	case StatusOnline:
		return "✅"
	case StatusOffline:
		return "❌"
	case StatusSyncing:
		return "🔄"
	case StatusStarting:
		return "🔄"
	case StatusStopping:
		return "⏹️"
	default:
		return "❓"
	}
}
