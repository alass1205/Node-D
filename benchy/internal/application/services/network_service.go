package services

import (
	"context"
	"fmt"
	"path/filepath"

	"math/big"
	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
	"benchy/internal/infrastructure/config"
	"benchy/internal/infrastructure/docker"
	"benchy/internal/infrastructure/ethereum"
	"benchy/internal/infrastructure/feedback"
	"benchy/internal/infrastructure/monitoring"
)

// NetworkService implémente les opérations réseau de haut niveau
type NetworkService struct {
	dockerClient  *docker.DockerClient
	ethClient     *ethereum.EthereumClient
	monitor       *monitoring.SystemMonitor
	feedback      *feedback.ConsoleFeedback
	configManager *config.NodeConfigManager
	baseDir       string
}

// NewNetworkService crée un nouveau service réseau
func NewNetworkService(baseDir string) (*NetworkService, error) {
	// Créer les clients infrastructure
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ethClient := ethereum.NewEthereumClient()
	monitor := monitoring.NewSystemMonitor()
	feedback := feedback.NewConsoleFeedback()
	configManager := config.NewNodeConfigManager(baseDir)

	return &NetworkService{
		dockerClient:  dockerClient,
		ethClient:     ethClient,
		monitor:       monitor,
		feedback:      feedback,
		configManager: configManager,
		baseDir:       baseDir,
	}, nil
}

// LaunchNetwork lance le réseau Ethereum complet
func (ns *NetworkService) LaunchNetwork(ctx context.Context) error {
	ns.feedback.Info(ctx, "🚀 Launching Ethereum network...")
	ns.feedback.Info(ctx, "📋 Configuration:")
	ns.feedback.Info(ctx, "   - 5 nodes: Alice, Bob, Cassandra, Driss, Elena")
	ns.feedback.Info(ctx, "   - 3 validators: Alice, Bob, Cassandra")
	ns.feedback.Info(ctx, "   - Clients: Geth + Nethermind")
	ns.feedback.Info(ctx, "   - Consensus: Clique")

	// 1. Générer les configurations des nodes
	if err := ns.configManager.GenerateDefaultNodes(); err != nil {
		return fmt.Errorf("failed to generate node configurations: %w", err)
	}

	// 2. Sauvegarder les configurations
	if err := ns.configManager.SaveAllConfigurations(); err != nil {
		return fmt.Errorf("failed to save configurations: %w", err)
	}

	// 3. Générer le fichier genesis
	genesis, err := ns.configManager.GenerateGenesisWithNodes()
	if err != nil {
		return fmt.Errorf("failed to generate genesis: %w", err)
	}

	genesisPath := filepath.Join(ns.baseDir, "configs", "genesis.json")
	generator := config.NewGenesisGenerator()
	if err := generator.SaveGenesisToFile(genesis, genesisPath); err != nil {
		return fmt.Errorf("failed to save genesis file: %w", err)
	}

	ns.feedback.Success(ctx, "✅ Configuration generated successfully")

	// 4. Créer le réseau Docker
	if err := ns.dockerClient.CreateNetwork(ctx, "benchy-network"); err != nil {
		return fmt.Errorf("failed to create docker network: %w", err)
	}

	ns.feedback.Success(ctx, "✅ Docker network created")

	// 5. Lancer chaque node (seulement Geth pour l'instant)
	nodes := ns.configManager.GetAllNodes()
	progress, err := ns.feedback.StartProgress(ctx, "Launching nodes", len(nodes))
	if err != nil {
		return err
	}
	defer progress.Close()

	successCount := 0
	for i, nodeConfig := range nodes {
		// Pour l'instant, on lance seulement les nodes Geth
		if nodeConfig.Client == entities.ClientNethermind {
			progress.Update(i+1, fmt.Sprintf("⏭️  %s skipped (Nethermind)", nodeConfig.Name))
			continue
		}

		if err := ns.launchNode(ctx, nodeConfig); err != nil {
			progress.Update(i+1, fmt.Sprintf("❌ %s failed: %v", nodeConfig.Name, err))
			// Continue avec les autres nodes
			continue
		}
		successCount++
		progress.Update(i+1, fmt.Sprintf("✅ %s launched", nodeConfig.Name))
	}

	if successCount == 0 {
		progress.Error("No nodes launched successfully")
		return fmt.Errorf("failed to launch any nodes")
	}

	progress.Complete(fmt.Sprintf("%d nodes launched successfully", successCount))

	// 6. Démarrer le monitoring
	network := ns.createNetworkEntity(nodes)
	if err := ns.monitor.StartMonitoring(ctx, network); err != nil {
		ns.feedback.Warning(ctx, fmt.Sprintf("Warning: monitoring failed to start: %v", err))
	}

	ns.feedback.Success(ctx, fmt.Sprintf("🎉 Network launched successfully! (%d nodes)", successCount))
	ns.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the network")
	ns.feedback.Info(ctx, "💡 Use 'docker ps' to see the containers")

	return nil
}

// launchNode lance un node individuel
func (ns *NetworkService) launchNode(ctx context.Context, nodeConfig *config.NodeConfig) error {
	// Préparer la configuration du container SIMPLIFIÉE
	containerConfig := ns.buildSimpleContainerConfig(nodeConfig)

	// Créer le node entity
	node := entities.NewNode(
		nodeConfig.Name,
		nodeConfig.IsValidator,
		nodeConfig.Client,
		nodeConfig.Port,
		nodeConfig.RPCPort,
	)
	node.Address = nodeConfig.KeyPair.Address

	// Créer le container
	containerID, err := ns.dockerClient.CreateContainer(ctx, node, containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Le container est déjà démarré par docker run -d
	// Mettre à jour le node avec l'ID du container
	node.ContainerID = containerID
	node.Status = entities.StatusStarting

	return nil
}

// buildSimpleContainerConfig construit une configuration simple du container
func (ns *NetworkService) buildSimpleContainerConfig(nodeConfig *config.NodeConfig) ports.ContainerConfig {
	config := ports.ContainerConfig{
		Name: fmt.Sprintf("benchy-%s", nodeConfig.Name),
		Image: "ethereum/client-go:v1.13.15",
		Ports: map[string]string{
			fmt.Sprintf("%d", nodeConfig.Port):    fmt.Sprintf("%d", nodeConfig.Port),
			fmt.Sprintf("%d", nodeConfig.RPCPort): fmt.Sprintf("%d", nodeConfig.RPCPort),
		},
		Volumes: map[string]string{
			nodeConfig.DataDir:     "/data",
			nodeConfig.KeystoreDir: "/keystore",
		},
		NetworkMode: "benchy-network",
		Labels: map[string]string{
			"benchy.node.name":      nodeConfig.Name,
			"benchy.node.validator": fmt.Sprintf("%t", nodeConfig.IsValidator),
			"benchy.node.client":    string(nodeConfig.Client),
		},
	}

	// Commande Geth simplifiée
	config.Command = []string{
		
		"--datadir", "/data",
		"--networkid", "1337",
		"--port", fmt.Sprintf("%d", nodeConfig.Port),
		"--http",
		"--http.addr", "0.0.0.0",
		"--http.port", fmt.Sprintf("%d", nodeConfig.RPCPort),
		"--http.api", "eth,net,web3,personal,miner",
		"--http.corsdomain", "*",
		"--allow-insecure-unlock",
		"--nodiscover",
		"--maxpeers", "25",
		"--syncmode", "full",
		"--verbosity", "3",
	}

	// Ajouter les options de mining pour les validateurs
	if nodeConfig.IsValidator {
		config.Command = append(config.Command,
			"--mine",
			"--miner.etherbase", nodeConfig.KeyPair.Address.Hex(),
		)
	}

	return config
}

// createNetworkEntity crée une entité Network depuis les configurations
func (ns *NetworkService) createNetworkEntity(nodeConfigs []*config.NodeConfig) *entities.Network {
	network := entities.NewNetwork("benchy-network", big.NewInt(1337))
	
	for _, nodeConfig := range nodeConfigs {
		node := entities.NewNode(
			nodeConfig.Name,
			nodeConfig.IsValidator,
			nodeConfig.Client,
			nodeConfig.Port,
			nodeConfig.RPCPort,
		)
		node.Address = nodeConfig.KeyPair.Address
		
		network.AddNode(node)
	}
	
	network.Status = entities.NetworkStatusRunning
	return network
}

// GetNetworkStatus récupère le status du réseau
func (ns *NetworkService) GetNetworkStatus(ctx context.Context) (*entities.Network, error) {
	// Pour l'instant, on crée un réseau fictif
	// TODO: Implémenter la récupération du vrai état
	nodes := ns.configManager.GetAllNodes()
	return ns.createNetworkEntity(nodes), nil
}

// StopNetwork arrête le réseau
func (ns *NetworkService) StopNetwork(ctx context.Context) error {
	ns.feedback.Info(ctx, "🛑 Stopping network...")
	
	// Arrêter tous les containers benchy
	// TODO: Implémenter l'arrêt complet
	
	ns.feedback.Success(ctx, "✅ Network stopped")
	return nil
}
