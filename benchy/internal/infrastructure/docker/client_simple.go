package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// DockerClient version hybride avec commandes docker CLI
type DockerClient struct {
	containers map[string]bool
}

// NewDockerClient crée un client hybride
func NewDockerClient() (*DockerClient, error) {
	// Vérifier que docker CLI est disponible
	if err := exec.Command("docker", "version").Run(); err != nil {
		return nil, fmt.Errorf("docker CLI not available: %w", err)
	}

	return &DockerClient{
		containers: make(map[string]bool),
	}, nil
}

// CreateContainer crée un container via docker CLI
func (dc *DockerClient) CreateContainer(ctx context.Context, node *entities.Node, config ports.ContainerConfig) (string, error) {
	// Construire la commande docker run
	args := []string{"run", "-d", "--name", config.Name}
	
	// Ajouter les ports
	for hostPort, containerPort := range config.Ports {
		args = append(args, "-p", fmt.Sprintf("%s:%s", hostPort, containerPort))
	}
	
	// Ajouter les volumes
	for hostPath, containerPath := range config.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}
	
	// Ajouter le réseau
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}
	
	// Ajouter l'image et la commande
	args = append(args, config.Image)
	// Pas de commande pour Geth - utiliser entrypoint par défaut
	// Ajouter les arguments seulement si pas vide
	if len(config.Command) > 0 {
		args = append(args, config.Command...)
	}
	
	// Exécuter la commande
	fmt.Printf("DEBUG: docker %s\n", strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	
	containerID := strings.TrimSpace(string(output))
	dc.containers[containerID] = true
	
	fmt.Printf("🐳 Created container %s with ID %s\n", config.Name, containerID[:12])
	return containerID, nil
}

// StartContainer démarre un container (déjà démarré par docker run)
func (dc *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	fmt.Printf("🚀 Container %s already started\n", containerID[:12])
	return nil
}

// StopContainer arrête un container
func (dc *DockerClient) StopContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	dc.containers[containerID] = false
	return nil
}

// RestartContainer redémarre un container
func (dc *DockerClient) RestartContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "restart", containerID)
	return cmd.Run()
}

// RemoveContainer supprime un container
func (dc *DockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	delete(dc.containers, containerID)
	return nil
}

// GetContainerInfo récupère les informations d'un container
func (dc *DockerClient) GetContainerInfo(ctx context.Context, containerID string) (*ports.ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID, "--format", "{{.Name}}|{{.State.Status}}|{{.Config.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	
	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected inspect output")
	}
	
	return &ports.ContainerInfo{
		ID:     containerID,
		Name:   strings.TrimPrefix(parts[0], "/"),
		Status: parts[1],
		Image:  parts[2],
	}, nil
}

// GetContainerLogs récupère les logs d'un container
func (dc *DockerClient) GetContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", fmt.Sprintf("%d", tail), containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var cleanLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, strings.TrimSpace(line))
		}
	}
	
	return cleanLines, nil
}

// IsContainerRunning vérifie si un container est en cours d'exécution
func (dc *DockerClient) IsContainerRunning(ctx context.Context, containerID string) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID, "--format", "{{.State.Running}}")
	output, err := cmd.Output()
	if err != nil {
		return false, nil // Container n'existe pas
	}
	
	return strings.TrimSpace(string(output)) == "true", nil
}

// CreateNetwork crée un réseau Docker
func (dc *DockerClient) CreateNetwork(ctx context.Context, networkName string) error {
	// Vérifier si le réseau existe
	cmd := exec.CommandContext(ctx, "docker", "network", "ls", "--filter", "name="+networkName, "--quiet")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		fmt.Printf("🌐 Network %s already exists\n", networkName)
		return nil
	}
	
	// Créer le réseau
	cmd = exec.CommandContext(ctx, "docker", "network", "create", networkName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}
	
	fmt.Printf("🌐 Created network %s\n", networkName)
	return nil
}

// RemoveNetwork supprime un réseau Docker
func (dc *DockerClient) RemoveNetwork(ctx context.Context, networkName string) error {
	cmd := exec.CommandContext(ctx, "docker", "network", "rm", networkName)
	return cmd.Run()
}

// ConnectToNetwork connecte un container à un réseau (déjà fait à la création)
func (dc *DockerClient) ConnectToNetwork(ctx context.Context, containerID, networkName string) error {
	return nil // Déjà connecté à la création
}

// GetContainerStats récupère les statistiques d'un container
func (dc *DockerClient) GetContainerStats(ctx context.Context, containerID string) (*ports.ContainerStats, error) {
	// Simulation pour l'instant
	return &ports.ContainerStats{
		CPUUsage:    float64(20 + (len(containerID) % 30)),
		MemoryUsage: uint64(100+len(containerID)%100) * 1024 * 1024,
		MemoryLimit: 1024 * 1024 * 1024,
		NetworkRX:   1024 * 1024,
		NetworkTX:   512 * 1024,
	}, nil
}
