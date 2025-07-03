package ports

import (
	"context"
	"benchy/internal/domain/entities"
)

// NetworkRepository définit les opérations sur le réseau
type NetworkRepository interface {
	// Gestion du réseau
	CreateNetwork(ctx context.Context, network *entities.Network) error
	GetNetwork(ctx context.Context, name string) (*entities.Network, error)
	UpdateNetwork(ctx context.Context, network *entities.Network) error
	DeleteNetwork(ctx context.Context, name string) error
	
	// Gestion des nodes
	AddNode(ctx context.Context, networkName string, node *entities.Node) error
	GetNode(ctx context.Context, networkName, nodeName string) (*entities.Node, error)
	UpdateNode(ctx context.Context, networkName string, node *entities.Node) error
	RemoveNode(ctx context.Context, networkName, nodeName string) error
	GetAllNodes(ctx context.Context, networkName string) ([]*entities.Node, error)
	
	// Status du réseau
	IsNetworkRunning(ctx context.Context, networkName string) (bool, error)
	GetNetworkStatus(ctx context.Context, networkName string) (entities.NetworkStatus, error)
}
