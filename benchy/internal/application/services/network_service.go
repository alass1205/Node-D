package services

import (
	"context"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"benchy/internal/domain/entities"
	"benchy/internal/infrastructure/docker"
	"benchy/internal/infrastructure/feedback"
	"benchy/internal/infrastructure/monitoring"
)

// NetworkService gère le lancement et la configuration du réseau
type NetworkService struct {
	dockerClient  *docker.DockerClient
	feedback      *feedback.ConsoleFeedback
	monitor       *monitoring.SystemMonitor
	baseDir       string
}

// NewNetworkService crée un nouveau service réseau
func NewNetworkService(baseDir string) (*NetworkService, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &NetworkService{
		dockerClient:  dockerClient,
		feedback:      feedback.NewConsoleFeedback(),
		monitor:       monitoring.NewSystemMonitor(),
		baseDir:       baseDir,
	}, nil
}

// LaunchNetwork lance le réseau Ethereum avec 5 nodes
func (ns *NetworkService) LaunchNetwork(ctx context.Context) error {
	ns.feedback.Info(ctx, "🚀 Launching Ethereum network...")

	// 1. Configuration
	ns.feedback.Info(ctx, "📋 Configuration:")
	ns.feedback.Info(ctx, "   - 5 nodes: Alice, Bob, Cassandra, Driss, Elena")
	ns.feedback.Info(ctx, "   - 3 validators: Alice, Bob, Cassandra")
	ns.feedback.Info(ctx, "   - Clients: Geth + Nethermind")
	ns.feedback.Info(ctx, "   - Consensus: Clique")

	ns.feedback.Success(ctx, "✅ Configuration generated successfully")

	// 2. Créer le réseau Docker
	if err := ns.dockerClient.CreateNetwork(ctx, "benchy-network"); err != nil {
		ns.feedback.Warning(ctx, "🌐 Network benchy-network already exists")
	} else {
		ns.feedback.Success(ctx, "🌐 Created network benchy-network")
	}
	ns.feedback.Success(ctx, "✅ Docker network created")

	// 3. Lancer tous les 5 nodes avec genesis init
	progress, err := ns.feedback.StartProgress(ctx, "Launching nodes", 5)
	if err != nil {
		return err
	}
	defer progress.Close()

	successCount := 0
	
	// Alice (Geth avec Genesis Init)
	if err := ns.launchAliceNodeWithGenesis(ctx); err != nil {
		progress.Update(1, fmt.Sprintf("❌ alice failed: %v", err))
	} else {
		successCount++
		progress.Update(1, "✅ alice launched (Geth+Genesis)")
	}
	time.Sleep(2 * time.Second)

	// Bob (Geth avec Genesis Init)
	if err := ns.launchBobNodeWithGenesis(ctx); err != nil {
		progress.Update(2, fmt.Sprintf("❌ bob failed: %v", err))
	} else {
		successCount++
		progress.Update(2, "✅ bob launched (Geth+Genesis)")
	}
	time.Sleep(2 * time.Second)

	// Cassandra (Nethermind)
	if err := ns.launchCassandraNode(ctx); err != nil {
		progress.Update(3, fmt.Sprintf("❌ cassandra failed: %v", err))
	} else {
		successCount++
		progress.Update(3, "✅ cassandra launched (Nethermind)")
	}
	time.Sleep(1 * time.Second)

	// Driss (Geth avec Genesis)
	if err := ns.launchDrissNodeWithGenesis(ctx); err != nil {
		progress.Update(4, fmt.Sprintf("❌ driss failed: %v", err))
	} else {
		successCount++
		progress.Update(4, "✅ driss launched (Geth+Genesis)")
	}
	time.Sleep(1 * time.Second)

	// Elena (Nethermind)
	if err := ns.launchElenaNode(ctx); err != nil {
		progress.Update(5, fmt.Sprintf("❌ elena failed: %v", err))
	} else {
		successCount++
		progress.Update(5, "✅ elena launched (Nethermind)")
	}

	if successCount == 0 {
		progress.Error("No nodes launched successfully")
		return fmt.Errorf("failed to launch any nodes")
	} else if successCount == 5 {
		progress.Complete("🎉 All 5 nodes launched successfully!")
	} else {
		progress.Complete(fmt.Sprintf("⚠️  %d/5 nodes launched", successCount))
	}

	ns.feedback.Success(ctx, fmt.Sprintf("🎉 Network launched with %d/5 nodes!", successCount))
	ns.feedback.Info(ctx, "💡 Use 'benchy infos' to monitor the network")
	
	return nil
}

// launchAliceNodeWithGenesis lance Alice avec genesis init
func (ns *NetworkService) launchAliceNodeWithGenesis(ctx context.Context) error {
	// Étape 1: Init genesis
	initCmd := []string{
		"docker", "run", "--rm",
		"-v", filepath.Join(ns.baseDir, "nodes/alice/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data", "init", "/genesis.json",
	}
	
	fmt.Printf("DEBUG INIT: %s\n", strings.Join(initCmd[1:], " "))
	execInitCmd := exec.CommandContext(ctx, initCmd[0], initCmd[1:]...)
	if err := execInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to init alice genesis: %w", err)
	}

	// Étape 2: Lancer le node
	cmd := []string{
		"docker", "run", "-d",
		"--name", "benchy-alice",
		"-p", "8545:8545",
		"-p", "30303:30303",
		"-v", filepath.Join(ns.baseDir, "nodes/alice/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data",
		"--networkid", "1337",
		"--port", "30303",
		"--http", "--http.addr", "0.0.0.0", "--http.port", "8545",
		"--http.api", "eth,net,web3,personal,miner,clique",
		"--http.corsdomain", "*",
		"--allow-insecure-unlock",
		"--nodiscover", "--maxpeers", "25",
		"--syncmode", "full", "--verbosity", "3",
		
	}

	fmt.Printf("DEBUG RUN: %s\n", strings.Join(cmd[1:], " "))
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create alice container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Printf("🐳 Created container benchy-alice with ID %s\n", containerID[:12])
	return nil
}

// launchBobNodeWithGenesis lance Bob avec genesis init
func (ns *NetworkService) launchBobNodeWithGenesis(ctx context.Context) error {
	// Étape 1: Init genesis
	initCmd := []string{
		"docker", "run", "--rm",
		"-v", filepath.Join(ns.baseDir, "nodes/bob/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data", "init", "/genesis.json",
	}
	
	fmt.Printf("DEBUG INIT: %s\n", strings.Join(initCmd[1:], " "))
	execInitCmd := exec.CommandContext(ctx, initCmd[0], initCmd[1:]...)
	if err := execInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to init bob genesis: %w", err)
	}

	// Étape 2: Lancer le node
	cmd := []string{
		"docker", "run", "-d",
		"--name", "benchy-bob",
		"-p", "8546:8546",
		"-p", "30304:30304",
		"-v", filepath.Join(ns.baseDir, "nodes/bob/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data",
		"--networkid", "1337",
		"--port", "30304",
		"--http", "--http.addr", "0.0.0.0", "--http.port", "8546",
		"--http.api", "eth,net,web3,personal,miner,clique",
		"--http.corsdomain", "*",
		"--allow-insecure-unlock",
		"--nodiscover", "--maxpeers", "25",
		"--syncmode", "full", "--verbosity", "3",
		
	}

	fmt.Printf("DEBUG RUN: %s\n", strings.Join(cmd[1:], " "))
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create bob container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Printf("🐳 Created container benchy-bob with ID %s\n", containerID[:12])
	return nil
}

// launchDrissNodeWithGenesis lance Driss avec genesis
func (ns *NetworkService) launchDrissNodeWithGenesis(ctx context.Context) error {
	// Init genesis
	initCmd := []string{
		"docker", "run", "--rm",
		"-v", filepath.Join(ns.baseDir, "nodes/driss/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data", "init", "/genesis.json",
	}
	
	execInitCmd := exec.CommandContext(ctx, initCmd[0], initCmd[1:]...)
	if err := execInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to init driss genesis: %w", err)
	}

	// Lancer node
	cmd := []string{
		"docker", "run", "-d",
		"--name", "benchy-driss",
		"-p", "8548:8548",
		"-p", "30306:30306",
		"-v", filepath.Join(ns.baseDir, "nodes/driss/data") + ":/data",
		"-v", filepath.Join(ns.baseDir, "genesis.json") + ":/genesis.json",
		"--network", "benchy-network",
		"ethereum/client-go:v1.13.15",
		"--datadir", "/data",
		"--networkid", "1337",
		"--port", "30306",
		"--http", "--http.addr", "0.0.0.0", "--http.port", "8548",
		"--http.api", "eth,net,web3,personal,clique",
		"--http.corsdomain", "*",
		"--allow-insecure-unlock",
		"--nodiscover", "--maxpeers", "25",
		"--syncmode", "full", "--verbosity", "3",
	}

	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create driss container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Printf("🐳 Created container benchy-driss with ID %s\n", containerID[:12])
	return nil
}

// launchCassandraNode lance Cassandra avec Nethermind
func (ns *NetworkService) launchCassandraNode(ctx context.Context) error {
	cmd := []string{
		"docker", "run", "-d",
		"--name", "benchy-cassandra",
		"-p", "8547:8547",
		"-p", "30305:30305",
		"--network", "benchy-network",
		"nethermind/nethermind:latest",
		"--config", "mainnet",
		"--JsonRpc.Enabled", "true",
		"--JsonRpc.Host", "0.0.0.0",
		"--JsonRpc.Port", "8547",
		"--Network.DiscoveryPort", "30305",
		"--Network.P2PPort", "30305",
	}

	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create cassandra container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Printf("🐳 Created container benchy-cassandra with ID %s\n", containerID[:12])
	return nil
}

// launchElenaNode lance Elena avec Nethermind
func (ns *NetworkService) launchElenaNode(ctx context.Context) error {
	cmd := []string{
		"docker", "run", "-d",
		"--name", "benchy-elena",
		"-p", "8549:8549",
		"-p", "30307:30307",
		"--network", "benchy-network",
		"nethermind/nethermind:latest",
		"--config", "mainnet",
		"--JsonRpc.Enabled", "true",
		"--JsonRpc.Host", "0.0.0.0",
		"--JsonRpc.Port", "8549",
		"--Network.DiscoveryPort", "30307",
		"--Network.P2PPort", "30307",
	}

	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create elena container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	fmt.Printf("🐳 Created container benchy-elena with ID %s\n", containerID[:12])
	return nil
}

// createNetworkEntity crée l'entité Network pour le monitoring  
func (ns *NetworkService) createNetworkEntity() *entities.Network {
	chainID := big.NewInt(1337)
	return entities.NewNetwork("benchy-network", chainID)
}
