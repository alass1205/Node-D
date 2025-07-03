package ports

import (
	"context"
	"benchy/internal/domain/entities"
)

// ContainerInfo représente les informations d'un container
type ContainerInfo struct {
	ID       string
	Name     string
	Status   string
	Image    string
	Ports    []string
	Networks []string
	
	// Métriques
	CPUUsage    float64
	MemoryUsage float64
	MemoryLimit uint64
	NetworkRX   uint64
	NetworkTX   uint64
}

// DockerService définit les opérations Docker
type DockerService interface {
	// Gestion des containers
	CreateContainer(ctx context.Context, node *entities.Node, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RestartContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	
	// Informations des containers
	GetContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error)
	GetContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error)
	IsContainerRunning(ctx context.Context, containerID string) (bool, error)
	
	// Métriques
	GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error)
	
	// Gestion du réseau Docker
	CreateNetwork(ctx context.Context, networkName string) error
	RemoveNetwork(ctx context.Context, networkName string) error
	ConnectToNetwork(ctx context.Context, containerID, networkName string) error
}

// ContainerConfig représente la configuration d'un container
type ContainerConfig struct {
	Image       string
	Name        string
	Ports       map[string]string // host:container
	Volumes     map[string]string // host:container
	Environment []string
	Command     []string
	NetworkMode string
	Labels      map[string]string
}

// ContainerStats représente les statistiques d'un container
type ContainerStats struct {
	CPUUsage    float64
	MemoryUsage uint64
	MemoryLimit uint64
	NetworkRX   uint64
	NetworkTX   uint64
	BlockRead   uint64
	BlockWrite  uint64
}
