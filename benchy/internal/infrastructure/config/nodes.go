package config

import (
	"fmt"
	"math/big"
	"github.com/ethereum/go-ethereum/core"
	"path/filepath"

	"benchy/internal/domain/entities"
	"github.com/ethereum/go-ethereum/common"
)

// NodeConfigManager gère la configuration des nodes
type NodeConfigManager struct {
	baseDir string
	nodes   []*NodeConfig
}

// NodeConfig représente la configuration complète d'un node
type NodeConfig struct {
	Name        string
	IsValidator bool
	Client      entities.ClientType
	Port        int
	RPCPort     int
	WSPort      int
	KeyPair     *KeyPair
	DataDir     string
	KeystoreDir string
}

// NewNodeConfigManager crée un nouveau gestionnaire de configuration
func NewNodeConfigManager(baseDir string) *NodeConfigManager {
	return &NodeConfigManager{
		baseDir: baseDir,
		nodes:   make([]*NodeConfig, 0),
	}
}

// GenerateDefaultNodes génère la configuration des 5 nodes par défaut
func (ncm *NodeConfigManager) GenerateDefaultNodes() error {
	defaultNodes := []struct {
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

	for _, nodeInfo := range defaultNodes {
		// Générer la paire de clés
		keyPair, err := GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate key pair for %s: %w", nodeInfo.name, err)
		}

		// Créer la configuration du node
		nodeConfig := &NodeConfig{
			Name:        nodeInfo.name,
			IsValidator: nodeInfo.isValidator,
			Client:      nodeInfo.client,
			Port:        nodeInfo.port,
			RPCPort:     nodeInfo.rpcPort,
			WSPort:      nodeInfo.rpcPort + 1000, // WebSocket port = RPC port + 1000
			KeyPair:     keyPair,
			DataDir:     filepath.Join(ncm.baseDir, "nodes", nodeInfo.name, "data"),
			KeystoreDir: filepath.Join(ncm.baseDir, "nodes", nodeInfo.name, "keystore"),
		}

		ncm.nodes = append(ncm.nodes, nodeConfig)
	}

	return nil
}

// SaveAllConfigurations sauvegarde toutes les configurations
func (ncm *NodeConfigManager) SaveAllConfigurations() error {
	for _, node := range ncm.nodes {
		if err := ncm.saveNodeConfiguration(node); err != nil {
			return fmt.Errorf("failed to save configuration for %s: %w", node.Name, err)
		}
	}
	return nil
}

// saveNodeConfiguration sauvegarde la configuration d'un node
func (ncm *NodeConfigManager) saveNodeConfiguration(node *NodeConfig) error {
	
	// Sauvegarder la paire de clés
	if err := node.KeyPair.SaveKeyPairToFile(node.KeystoreDir, node.Name); err != nil {
		return fmt.Errorf("failed to save key pair: %w", err)
	}

	return nil
}

// GetValidators retourne les adresses des validateurs
func (ncm *NodeConfigManager) GetValidators() []common.Address {
	var validators []common.Address
	for _, node := range ncm.nodes {
		if node.IsValidator {
			validators = append(validators, node.KeyPair.Address)
		}
	}
	return validators
}

// GetAllAddresses retourne toutes les adresses
func (ncm *NodeConfigManager) GetAllAddresses() []common.Address {
	var addresses []common.Address
	for _, node := range ncm.nodes {
		addresses = append(addresses, node.KeyPair.Address)
	}
	return addresses
}

// GetNodeByName retourne la configuration d'un node par son nom
func (ncm *NodeConfigManager) GetNodeByName(name string) *NodeConfig {
	for _, node := range ncm.nodes {
		if node.Name == name {
			return node
		}
	}
	return nil
}

// GetAllNodes retourne toutes les configurations de nodes
func (ncm *NodeConfigManager) GetAllNodes() []*NodeConfig {
	return ncm.nodes
}

// GenerateGenesisWithNodes génère le genesis avec les nodes configurés
func (ncm *NodeConfigManager) GenerateGenesisWithNodes() (*core.Genesis, error) {
	generator := NewGenesisGenerator()
	
	// Ajouter tous les validateurs
	for _, node := range ncm.nodes {
		if node.IsValidator {
			generator.AddValidator(node.KeyPair.Address)
		}
	}
	
	// Ajouter une petite allocation pour les nodes non-validateurs
	smallBalance := new(big.Int)
	smallBalance.SetString("10000000000000000000", 10) // 10 ETH en wei
	
	for _, node := range ncm.nodes {
		if !node.IsValidator {
			generator.AddAllocation(node.KeyPair.Address, smallBalance)
		}
	}
	
	return generator.GenerateGenesis()
}

// LoadExistingConfigurations charge les configurations existantes
func (ncm *NodeConfigManager) LoadExistingConfigurations() error {
	// Pour l'instant, on génère toujours de nouvelles configurations
	// TODO: Implémenter le chargement depuis des fichiers existants
	return ncm.GenerateDefaultNodes()
}
