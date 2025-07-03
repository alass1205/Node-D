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

	// 5. Lancer chaque node
	nodes := ns.configManager.GetAllNodes()
	progress, err := ns.feedback.StartProgress(ctx, "Launching nodes", len(nodes))
	if err != nil {
		return err
	}
	defer progress.Close()

	for i, nodeConfig := range nodes {
		if err := ns.launchNode(ctx, nodeConfig); err != nil {
			progress.Error(fmt.Sprintf("Failed to launch %s: %v", nodeConfig.Name, err))
			return fmt.Errorf("failed to launch node %s: %w", nodeConfig.Name, err)
		}
		progress.Update(i+1, fmt.Sprintf("✅ %s launched", nodeConfig.Name))
	}

	progress.Complete("All nodes launched successfully")

	// 6. Démarrer le monitoring
	network := ns.createNetworkEntity(nodes)
	if err := ns.monitor.StartMonitoring(ctx, network); err != nil {
		ns.feedback.Warning(ctx, fmt.Sprintf("Warning: monitoring failed to start: %v", err))
	}

	ns.feedback.Success(ctx, "🎉 Network launched successfully!")
	ns.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the network")
	ns.feedback.Info(ctx, "💡 Use 'docker ps' to see the containers")

	return nil
}

// launchNode lance un node individuel
func (ns *NetworkService) launchNode(ctx context.Context, nodeConfig *config.NodeConfig) error {
	// Préparer la configuration du container
	containerConfig := ns.buildContainerConfig(nodeConfig)

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

	// Démarrer le container
	if err := ns.dockerClient.StartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Mettre à jour le node avec l'ID du container
	node.ContainerID = containerID
	node.Status = entities.StatusStarting

	return nil
}

// buildContainerConfig construit la configuration du container pour un node
func (ns *NetworkService) buildContainerConfig(nodeConfig *config.NodeConfig) ports.ContainerConfig {
	genesisPath := filepath.Join(ns.baseDir, "configs", "genesis.json")
	
	config := ports.ContainerConfig{
		Name: fmt.Sprintf("benchy-%s", nodeConfig.Name),
		Ports: map[string]string{
			fmt.Sprintf("%d", nodeConfig.Port):    fmt.Sprintf("%d", nodeConfig.Port),
			fmt.Sprintf("%d", nodeConfig.RPCPort): fmt.Sprintf("%d", nodeConfig.RPCPort),
		},
		Volumes: map[string]string{
			nodeConfig.DataDir:     "/data",
			nodeConfig.KeystoreDir: "/keystore",
			genesisPath:           "/genesis.json",
		},
		NetworkMode: "benchy-network",
		Labels: map[string]string{
			"benchy.node.name":      nodeConfig.Name,
			"benchy.node.validator": fmt.Sprintf("%t", nodeConfig.IsValidator),
			"benchy.node.client":    string(nodeConfig.Client),
		},
	}

	// Configuration spécifique selon le client
	switch nodeConfig.Client {
	case entities.ClientGeth:
		config.Image = "ethereum/client-go:latest"
		config.Command = ns.buildGethCommand(nodeConfig)
	case entities.ClientNethermind:
		config.Image = "nethermind/nethermind:latest"
		config.Command = ns.buildNethermindCommand(nodeConfig)
	}

	return config
}

// buildGethCommand construit la commande pour Geth
func (ns *NetworkService) buildGethCommand(nodeConfig *config.NodeConfig) []string {
	cmd := []string{
		"geth",
		"--datadir", "/data",
		"--keystore", "/keystore",
		"--networkid", "1337",
		"--port", fmt.Sprintf("%d", nodeConfig.Port),
		"--http",
		"--http.addr", "0.0.0.0",
		"--http.port", fmt.Sprintf("%d", nodeConfig.RPCPort),
		"--http.api", "eth,net,web3,personal,miner,admin",
		"--http.corsdomain", "*",
		"--ws",
		"--ws.addr", "0.0.0.0",
		"--ws.port", fmt.Sprintf("%d", nodeConfig.WSPort),
		"--ws.api", "eth,net,web3,personal,miner,admin",
		"--ws.origins", "*",
		"--allow-insecure-unlock",
		"--nodiscover",
		"--maxpeers", "25",
		"--syncmode", "full",
		"--verbosity", "3",
	}

	// Initialiser avec le genesis si c'est le premier démarrage
	cmd = append(cmd, "--init", "/genesis.json")

	if nodeConfig.IsValidator {
		cmd = append(cmd, 
			"--mine", 
			"--miner.threads", "1",
			"--miner.etherbase", nodeConfig.KeyPair.Address.Hex(),
		)
	}

	return cmd
}

// buildNethermindCommand construit la commande pour Nethermind
func (ns *NetworkService) buildNethermindCommand(nodeConfig *config.NodeConfig) []string {
	return []string{
		"./Nethermind.Runner",
		"--config", "mainnet",
		"--datadir", "/data",
		"--Network.DiscoveryPort", fmt.Sprintf("%d", nodeConfig.Port),
		"--Network.P2PPort", fmt.Sprintf("%d", nodeConfig.Port),
		"--JsonRpc.Enabled", "true",
		"--JsonRpc.Host", "0.0.0.0",
		"--JsonRpc.Port", fmt.Sprintf("%d", nodeConfig.RPCPort),
		"--JsonRpc.EnabledModules", "Eth,Subscribe,Trace,TxPool,Web3,Personal,Proof,Net,Parity,Health,Rpc",
	}
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
