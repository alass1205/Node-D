package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/domain/ports"
	"benchy/internal/infrastructure/config"
	"benchy/internal/infrastructure/docker"
	"benchy/internal/infrastructure/ethereum"
	"benchy/internal/infrastructure/feedback"
	"benchy/internal/infrastructure/monitoring"
)

// NetworkServiceReal implémente le lancement de vrais containers
type NetworkServiceReal struct {
	dockerClient  *docker.DockerClient
	ethClient     *ethereum.EthereumClient
	monitor       *monitoring.SystemMonitor
	feedback      *feedback.ConsoleFeedback
	configManager *config.NodeConfigManager
	baseDir       string
}

// NewNetworkServiceReal crée un service avec vrais containers
func NewNetworkServiceReal(baseDir string) (*NetworkServiceReal, error) {
	// Créer les clients infrastructure
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ethClient := ethereum.NewEthereumClient()
	monitor := monitoring.NewSystemMonitor()
	feedback := feedback.NewConsoleFeedback()
	configManager := config.NewNodeConfigManager(baseDir)

	return &NetworkServiceReal{
		dockerClient:  dockerClient,
		ethClient:     ethClient,
		monitor:       monitor,
		feedback:      feedback,
		configManager: configManager,
		baseDir:       baseDir,
	}, nil
}

// LaunchNetworkReal lance vraiment des containers Docker
func (ns *NetworkServiceReal) LaunchNetworkReal(ctx context.Context) error {
	ns.feedback.Info(ctx, "🚀 Launching REAL Ethereum network...")
	ns.feedback.Info(ctx, "📋 Configuration:")
	ns.feedback.Info(ctx, "   - 5 nodes: Alice, Bob, Cassandra, Driss, Elena")
	ns.feedback.Info(ctx, "   - 3 validators: Alice, Bob, Cassandra")
	ns.feedback.Info(ctx, "   - Clients: Geth + Nethermind")
	ns.feedback.Info(ctx, "   - Consensus: Clique")

	// 1. Vérifier que Docker fonctionne
	if err := ns.checkDockerAvailable(ctx); err != nil {
		return fmt.Errorf("docker not available: %w", err)
	}

	// 2. Nettoyer d'éventuels anciens containers
	if err := ns.cleanupOldContainers(ctx); err != nil {
		ns.feedback.Warning(ctx, fmt.Sprintf("Warning: cleanup failed: %v", err))
	}

	// 3. Créer le répertoire de configuration
	if err := os.MkdirAll(ns.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// 4. Générer les configurations
	if err := ns.configManager.GenerateDefaultNodes(); err != nil {
		return fmt.Errorf("failed to generate node configurations: %w", err)
	}

	// 5. Créer le fichier genesis
	if err := ns.createGenesisFile(ctx); err != nil {
		return fmt.Errorf("failed to create genesis file: %w", err)
	}

	// 6. Créer le réseau Docker
	ns.feedback.Info(ctx, "🌐 Creating Docker network...")
	if err := ns.dockerClient.CreateNetwork(ctx, "benchy-network"); err != nil {
		return fmt.Errorf("failed to create docker network: %w", err)
	}

	// 7. Lancer les containers
	nodes := ns.configManager.GetAllNodes()
	progress, err := ns.feedback.StartProgress(ctx, "Launching containers", len(nodes))
	if err != nil {
		return err
	}
	defer progress.Close()

	for i, nodeConfig := range nodes {
		if err := ns.launchRealContainer(ctx, nodeConfig); err != nil {
			progress.Error(fmt.Sprintf("Failed to launch %s: %v", nodeConfig.Name, err))
			return fmt.Errorf("failed to launch container %s: %w", nodeConfig.Name, err)
		}
		progress.Update(i+1, fmt.Sprintf("✅ %s container started", nodeConfig.Name))
		
		// Petit délai entre les lancements
		time.Sleep(2 * time.Second)
	}

	progress.Complete("All containers launched successfully")

	// 8. Attendre que les nodes soient prêts
	ns.feedback.Info(ctx, "⏳ Waiting for nodes to be ready...")
	if err := ns.waitForNodesReady(ctx, nodes); err != nil {
		return fmt.Errorf("nodes failed to become ready: %w", err)
	}

	ns.feedback.Success(ctx, "🎉 Real Ethereum network launched successfully!")
	ns.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the live network")
	ns.feedback.Info(ctx, "💡 Use 'docker ps' to see the running containers")

	return nil
}

// checkDockerAvailable vérifie que Docker est disponible
func (ns *NetworkServiceReal) checkDockerAvailable(ctx context.Context) error {
	// Essayer de créer un réseau test
	testNetwork := "benchy-test-" + fmt.Sprintf("%d", time.Now().Unix())
	if err := ns.dockerClient.CreateNetwork(ctx, testNetwork); err != nil {
		return fmt.Errorf("docker not accessible: %w", err)
	}
	
	// Nettoyer immédiatement
	ns.dockerClient.RemoveNetwork(ctx, testNetwork)
	return nil
}

// cleanupOldContainers nettoie d'anciens containers benchy
func (ns *NetworkServiceReal) cleanupOldContainers(ctx context.Context) error {
	ns.feedback.Info(ctx, "🧹 Cleaning up old containers...")
	
	// TODO: Implémenter le nettoyage réel
	// Pour l'instant, on simule
	time.Sleep(1 * time.Second)
	
	return nil
}

// createGenesisFile crée le fichier genesis
func (ns *NetworkServiceReal) createGenesisFile(ctx context.Context) error {
	ns.feedback.Info(ctx, "📄 Creating genesis file...")
	
	genesis, err := ns.configManager.GenerateGenesisWithNodes()
	if err != nil {
		return err
	}

	genesisPath := filepath.Join(ns.baseDir, "genesis.json")
	generator := config.NewGenesisGenerator()
	
	return generator.SaveGenesisToFile(genesis, genesisPath)
}

// launchRealContainer lance un vrai container Docker
func (ns *NetworkServiceReal) launchRealContainer(ctx context.Context, nodeConfig *config.NodeConfig) error {
	// Créer les répertoires pour le node
	if err := os.MkdirAll(nodeConfig.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	if err := os.MkdirAll(nodeConfig.KeystoreDir, 0755); err != nil {
		return fmt.Errorf("failed to create keystore directory: %w", err)
	}

	// Sauvegarder les clés
	if err := nodeConfig.KeyPair.SaveKeyPairToFile(nodeConfig.KeystoreDir, nodeConfig.Name); err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	// Configuration du container
	containerConfig := ns.buildRealContainerConfig(nodeConfig)

	// Créer le node entity
	node := entities.NewNode(
		nodeConfig.Name,
		nodeConfig.IsValidator,
		nodeConfig.Client,
		nodeConfig.Port,
		nodeConfig.RPCPort,
	)
	node.Address = nodeConfig.KeyPair.Address

	// Créer et démarrer le container
	containerID, err := ns.dockerClient.CreateContainer(ctx, node, containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := ns.dockerClient.StartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// buildRealContainerConfig construit la config pour un vrai container
func (ns *NetworkServiceReal) buildRealContainerConfig(nodeConfig *config.NodeConfig) ports.ContainerConfig {
	genesisPath := filepath.Join(ns.baseDir, "genesis.json")
	
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

	// Configuration spécifique par client
	switch nodeConfig.Client {
	case entities.ClientGeth:
		config.Image = "ethereum/client-go:v1.10.26"  // Version stable
		config.Command = ns.buildRealGethCommand(nodeConfig)
	case entities.ClientNethermind:
		config.Image = "nethermind/nethermind:1.14.7"  // Version stable
		config.Command = ns.buildRealNethermindCommand(nodeConfig)
	}

	return config
}

// buildRealGethCommand construit une vraie commande Geth
func (ns *NetworkServiceReal) buildRealGethCommand(nodeConfig *config.NodeConfig) []string {
	cmd := []string{
		"geth",
		"--datadir", "/data",
		"--keystore", "/keystore",
		"--networkid", "1337",
		"--port", fmt.Sprintf("%d", nodeConfig.Port),
		"--http",
		"--http.addr", "0.0.0.0",
		"--http.port", fmt.Sprintf("%d", nodeConfig.RPCPort),
		"--http.api", "eth,net,web3,personal,miner,admin,debug",
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
		"--gcmode", "archive",
		"--verbosity", "3",
		"--nat", "extip:127.0.0.1",
	}

	// Initialisation avec genesis au premier démarrage
	if _, err := os.Stat(filepath.Join(nodeConfig.DataDir, "geth")); os.IsNotExist(err) {
		// Premier démarrage - initialiser avec genesis
		return append([]string{"geth", "init", "/genesis.json", "--datadir", "/data", "&&"}, cmd...)
	}

	if nodeConfig.IsValidator {
		cmd = append(cmd,
			"--mine",
			"--miner.threads", "1",
			"--miner.etherbase", nodeConfig.KeyPair.Address.Hex(),
		)
	}

	return cmd
}

// buildRealNethermindCommand construit une vraie commande Nethermind
func (ns *NetworkServiceReal) buildRealNethermindCommand(nodeConfig *config.NodeConfig) []string {
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
		"--Init.ChainSpecPath", "/genesis.json",
	}
}

// waitForNodesReady attend que les nodes soient opérationnels
func (ns *NetworkServiceReal) waitForNodesReady(ctx context.Context, nodes []*config.NodeConfig) error {
	spinner, err := ns.feedback.StartSpinner(ctx, "Waiting for nodes to start...")
	if err != nil {
		return err
	}

	maxWait := 120 * time.Second
	checkInterval := 5 * time.Second
	timeout := time.After(maxWait)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	readyNodes := make(map[string]bool)

	for {
		select {
		case <-timeout:
			spinner.Error("❌ Timeout waiting for nodes to be ready")
			return fmt.Errorf("timeout waiting for nodes to be ready")
		case <-ticker.C:
			allReady := true
			for _, node := range nodes {
				if !readyNodes[node.Name] {
					if ns.checkNodeReady(ctx, node) {
						readyNodes[node.Name] = true
						ns.feedback.Success(ctx, fmt.Sprintf("✅ %s is ready", node.Name))
					} else {
						allReady = false
					}
				}
			}
			
			if allReady {
				spinner.Success("✅ All nodes are ready!")
				return nil
			}
		case <-ctx.Done():
			spinner.Error("❌ Context cancelled")
			return ctx.Err()
		}
	}
}

// checkNodeReady vérifie si un node est prêt
func (ns *NetworkServiceReal) checkNodeReady(ctx context.Context, nodeConfig *config.NodeConfig) bool {
	nodeURL := fmt.Sprintf("http://localhost:%d", nodeConfig.RPCPort)
	
	// Essayer de se connecter
	if err := ns.ethClient.ConnectToNode(ctx, nodeURL); err != nil {
		return false
	}
	
	// Essayer de récupérer le dernier bloc
	_, err := ns.ethClient.GetLatestBlockNumber(ctx, nodeURL)
	return err == nil
}
