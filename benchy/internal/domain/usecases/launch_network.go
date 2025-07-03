package usecases

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
)

// LaunchNetworkUseCase gère le lancement du réseau
type LaunchNetworkUseCase struct {
	networkRepo   ports.NetworkRepository
	dockerService ports.DockerService
	ethService    ports.EthereumService
	feedback      ports.FeedbackService
}

// NewLaunchNetworkUseCase crée une nouvelle instance
func NewLaunchNetworkUseCase(
	networkRepo ports.NetworkRepository,
	dockerService ports.DockerService,
	ethService ports.EthereumService,
	feedback ports.FeedbackService,
) *LaunchNetworkUseCase {
	return &LaunchNetworkUseCase{
		networkRepo:   networkRepo,
		dockerService: dockerService,
		ethService:    ethService,
		feedback:      feedback,
	}
}

// Execute lance le réseau Ethereum
func (uc *LaunchNetworkUseCase) Execute(ctx context.Context) error {
	// 1. Créer le réseau
	network := entities.NewNetwork("benchy-network", big.NewInt(1337))
	
	// 2. Ajouter les 5 nodes
	if err := uc.createNodes(network); err != nil {
		return fmt.Errorf("failed to create nodes: %w", err)
	}
	
	// 3. Feedback utilisateur
	uc.feedback.Info(ctx, "🚀 Launching Ethereum network...")
	uc.feedback.Info(ctx, "📋 Configuration:")
	uc.feedback.Info(ctx, "   - 5 nodes: Alice, Bob, Cassandra, Driss, Elena")
	uc.feedback.Info(ctx, "   - 3 validators: Alice, Bob, Cassandra")
	uc.feedback.Info(ctx, "   - Clients: Geth + Nethermind")
	uc.feedback.Info(ctx, "   - Consensus: Clique")
	
	// 4. Créer le réseau Docker
	if err := uc.dockerService.CreateNetwork(ctx, "benchy-network"); err != nil {
		return fmt.Errorf("failed to create docker network: %w", err)
	}
	
	// 5. Lancer chaque node
	progress, err := uc.feedback.StartProgress(ctx, "Launching nodes", len(network.Nodes))
	if err != nil {
		return err
	}
	defer progress.Close()
	
	for i, node := range network.Nodes {
		if err := uc.launchNode(ctx, node); err != nil {
			progress.Error(fmt.Sprintf("Failed to launch %s: %v", node.Name, err))
			return fmt.Errorf("failed to launch node %s: %w", node.Name, err)
		}
		progress.Update(i+1, fmt.Sprintf("✅ %s launched", node.Name))
	}
	
	// 6. Attendre que les nodes se connectent
	if err := uc.waitForNetworkReady(ctx, network); err != nil {
		return fmt.Errorf("network failed to become ready: %w", err)
	}
	
	// 7. Sauvegarder la configuration
	if err := uc.networkRepo.CreateNetwork(ctx, network); err != nil {
		return fmt.Errorf("failed to save network configuration: %w", err)
	}
	
	uc.feedback.Success(ctx, "✅ Network launched successfully!")
	uc.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the network")
	
	return nil
}

// createNodes crée les 5 nodes avec leur configuration
func (uc *LaunchNetworkUseCase) createNodes(network *entities.Network) error {
	nodes := []struct {
		name        string
		isValidator bool
		client      entities.ClientType
		port        int
		rpcPort     int
	}{
		{"alice", true, entities.ClientGeth, 30303, 8545},
		{"bob", true, entities.ClientGeth, 30304, 8546},
		{"cassandra", true, entities.ClientNethermind, 30305, 8547},
		{"driss", false, entities.ClientGeth, 30306, 8548},
		{"elena", false, entities.ClientNethermind, 30307, 8549},
	}
	
	for _, nodeConfig := range nodes {
		node := entities.NewNode(
			nodeConfig.name,
			nodeConfig.isValidator,
			nodeConfig.client,
			nodeConfig.port,
			nodeConfig.rpcPort,
		)
		
		// Générer une clé privée pour ce node
		// TODO: Implémenter la génération de clé
		
		network.AddNode(node)
	}
	
	return nil
}

// launchNode lance un node individuel
func (uc *LaunchNetworkUseCase) launchNode(ctx context.Context, node *entities.Node) error {
	// Configuration du container
	config := ports.ContainerConfig{
		Name:        fmt.Sprintf("benchy-%s", node.Name),
		Ports:       map[string]string{
			fmt.Sprintf("%d", node.Port):    fmt.Sprintf("%d", node.Port),
			fmt.Sprintf("%d", node.RPCPort): fmt.Sprintf("%d", node.RPCPort),
		},
		NetworkMode: "benchy-network",
		Labels: map[string]string{
			"benchy.node.name":        node.Name,
			"benchy.node.validator":   fmt.Sprintf("%t", node.IsValidator),
			"benchy.node.client":      string(node.Client),
		},
	}
	
	// Choisir l'image Docker selon le client
	switch node.Client {
	case entities.ClientGeth:
		config.Image = "ethereum/client-go:latest"
		config.Command = uc.getGethCommand(node)
	case entities.ClientNethermind:
		config.Image = "nethermind/nethermind:latest"
		config.Command = uc.getNethermindCommand(node)
	}
	
	// Créer et démarrer le container
	containerID, err := uc.dockerService.CreateContainer(ctx, node, config)
	if err != nil {
		return err
	}
	
	node.ContainerID = containerID
	node.Status = entities.StatusStarting
	
	if err := uc.dockerService.StartContainer(ctx, containerID); err != nil {
		return err
	}
	
	// Attendre que le node soit prêt
	if err := uc.waitForNodeReady(ctx, node); err != nil {
		return err
	}
	
	node.Status = entities.StatusOnline
	return nil
}

// getGethCommand retourne la commande pour Geth
func (uc *LaunchNetworkUseCase) getGethCommand(node *entities.Node) []string {
	cmd := []string{
		"geth",
		"--networkid", "1337",
		"--datadir", "/data",
		"--keystore", "/keystore",
		"--port", fmt.Sprintf("%d", node.Port),
		"--http",
		"--http.addr", "0.0.0.0",
		"--http.port", fmt.Sprintf("%d", node.RPCPort),
		"--http.api", "eth,net,web3,personal,miner",
		"--ws",
		"--ws.addr", "0.0.0.0",
		"--ws.port", fmt.Sprintf("%d", node.RPCPort+1000),
		"--ws.api", "eth,net,web3,personal,miner",
		"--allow-insecure-unlock",
		"--nodiscover",
		"--syncmode", "full",
	}
	
	if node.IsValidator {
		cmd = append(cmd, "--mine", "--miner.threads", "1")
	}
	
	return cmd
}

// getNethermindCommand retourne la commande pour Nethermind
func (uc *LaunchNetworkUseCase) getNethermindCommand(node *entities.Node) []string {
	return []string{
		"./Nethermind.Runner",
		"--config", "mainnet",
		"--datadir", "/data",
		"--Network.DiscoveryPort", fmt.Sprintf("%d", node.Port),
		"--Network.P2PPort", fmt.Sprintf("%d", node.Port),
		"--JsonRpc.Enabled", "true",
		"--JsonRpc.Host", "0.0.0.0",
		"--JsonRpc.Port", fmt.Sprintf("%d", node.RPCPort),
	}
}

// waitForNodeReady attend que le node soit prêt
func (uc *LaunchNetworkUseCase) waitForNodeReady(ctx context.Context, node *entities.Node) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	nodeURL := fmt.Sprintf("http://localhost:%d", node.RPCPort)
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("node %s failed to start within timeout", node.Name)
		case <-ticker.C:
			if err := uc.ethService.ConnectToNode(ctx, nodeURL); err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// waitForNetworkReady attend que le réseau soit prêt
func (uc *LaunchNetworkUseCase) waitForNetworkReady(ctx context.Context, network *entities.Network) error {
	uc.feedback.Info(ctx, "⏳ Waiting for network to stabilize...")
	
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("network failed to stabilize within timeout")
		case <-ticker.C:
			if network.IsHealthy() {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
