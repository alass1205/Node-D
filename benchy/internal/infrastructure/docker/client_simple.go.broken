package docker

import (
	"context"
	"fmt"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// DockerClient version simplifiée sans dépendances problématiques
type DockerClient struct {
	containers map[string]bool
}

// NewDockerClient crée un nouveau client Docker simplifié
func NewDockerClient() (*DockerClient, error) {
	return &DockerClient{
		containers: make(map[string]bool),
	}, nil
}

// CreateContainer simule la création d'un container
func (dc *DockerClient) CreateContainer(ctx context.Context, node *entities.Node, config ports.ContainerConfig) (string, error) {
	containerID := fmt.Sprintf("benchy-%s-%s", node.Name, "abc123")
	dc.containers[containerID] = false
	return containerID, nil
}

// StartContainer simule le démarrage
func (dc *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	dc.containers[containerID] = true
	return nil
}

// StopContainer simule l'arrêt
func (dc *DockerClient) StopContainer(ctx context.Context, containerID string) error {
	dc.containers[containerID] = false
	return nil
}

// RestartContainer simule le redémarrage
func (dc *DockerClient) RestartContainer(ctx context.Context, containerID string) error {
	return nil
}

// RemoveContainer simule la suppression
func (dc *DockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	delete(dc.containers, containerID)
	return nil
}

// GetContainerInfo simule la récupération d'infos
func (dc *DockerClient) GetContainerInfo(ctx context.Context, containerID string) (*ports.ContainerInfo, error) {
	return &ports.ContainerInfo{
		ID:     containerID,
		Name:   "benchy-test",
		Status: "running",
	}, nil
}

// GetContainerLogs simule la récupération de logs
func (dc *DockerClient) GetContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error) {
	return []string{"Container log line 1", "Container log line 2"}, nil
}

// IsContainerRunning simule la vérification
func (dc *DockerClient) IsContainerRunning(ctx context.Context, containerID string) (bool, error) {
	return dc.containers[containerID], nil
}

// GetContainerStats simule les statistiques
func (dc *DockerClient) GetContainerStats(ctx context.Context, containerID string) (*ports.ContainerStats, error) {
	return &ports.ContainerStats{
		CPUUsage:    25.5,
		MemoryUsage: 128 * 1024 * 1024, // 128MB
	}, nil
}

// CreateNetwork simule la création de réseau
func (dc *DockerClient) CreateNetwork(ctx context.Context, networkName string) error {
	return nil
}

// RemoveNetwork simule la suppression de réseau
func (dc *DockerClient) RemoveNetwork(ctx context.Context, networkName string) error {
	return nil
}

// ConnectToNetwork simule la connexion au réseau
func (dc *DockerClient) ConnectToNetwork(ctx context.Context, containerID, networkName string) error {
	return nil
}
