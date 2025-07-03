package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// GenesisGenerator gère la génération de la configuration genesis
type GenesisGenerator struct {
	chainID     *big.Int
	period      uint64 // Block time en secondes
	epoch       uint64 // Epoch length pour Clique
	validators  []common.Address
	allocations map[common.Address]*big.Int
}

// NewGenesisGenerator crée un nouveau générateur de genesis
func NewGenesisGenerator() *GenesisGenerator {
	return &GenesisGenerator{
		chainID:     big.NewInt(1337),
		period:      5,  // 5 secondes par bloc
		epoch:       30000,
		allocations: make(map[common.Address]*big.Int),
	}
}

// AddValidator ajoute un validateur au genesis
func (g *GenesisGenerator) AddValidator(address common.Address) {
	g.validators = append(g.validators, address)
	
	// Donner 1000 ETH à chaque validateur
	initialBalance := new(big.Int)
	initialBalance.SetString("1000000000000000000000", 10) // 1000 ETH en wei
	g.allocations[address] = initialBalance
}

// AddAllocation ajoute une allocation d'ETH pour une adresse
func (g *GenesisGenerator) AddAllocation(address common.Address, balance *big.Int) {
	g.allocations[address] = balance
}

// GenerateGenesis génère la configuration genesis Clique
func (g *GenesisGenerator) GenerateGenesis() (*core.Genesis, error) {
	if len(g.validators) == 0 {
		return nil, fmt.Errorf("au moins un validateur requis")
	}
	
	// Créer la configuration Clique
	config := &params.ChainConfig{
		ChainID:                 g.chainID,
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		Clique: &params.CliqueConfig{
			Period: g.period,
			Epoch:  g.epoch,
		},
	}
	
	// Créer les allocations pour le genesis
	alloc := make(core.GenesisAlloc)
	for address, balance := range g.allocations {
		alloc[address] = core.GenesisAccount{
			Balance: balance,
		}
	}
	
	// Créer l'extraData pour Clique (contient les validateurs)
	extraData := make([]byte, 32) // 32 bytes de padding
	for _, validator := range g.validators {
		extraData = append(extraData, validator.Bytes()...)
	}
	extraData = append(extraData, make([]byte, 65)...) // 65 bytes signature vide
	
	genesis := &core.Genesis{
		Config:     config,
		Nonce:      0,
		Timestamp:  0,
		ExtraData:  extraData,
		GasLimit:   8000000,
		Difficulty: big.NewInt(1),
		Mixhash:    common.Hash{},
		Coinbase:   common.Address{},
		Alloc:      alloc,
	}
	
	return genesis, nil
}

// SaveGenesisToFile sauvegarde le genesis dans un fichier JSON
func (g *GenesisGenerator) SaveGenesisToFile(genesis *core.Genesis, filePath string) error {
	// Créer le répertoire si nécessaire
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Convertir en JSON
	genesisJSON, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis: %w", err)
	}
	
	// Écrire le fichier
	if err := os.WriteFile(filePath, genesisJSON, 0644); err != nil {
		return fmt.Errorf("failed to write genesis file: %w", err)
	}
	
	return nil
}

// KeyPair représente une paire de clé privée/publique
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    common.Address
}

// GenerateKeyPair génère une nouvelle paire de clés
func GenerateKeyPair() (*KeyPair, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	
	publicKey := &privateKey.PublicKey
	address := crypto.PubkeyToAddress(*publicKey)
	
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

// SaveKeyPairToFile sauvegarde la clé privée dans un fichier
func (kp *KeyPair) SaveKeyPairToFile(keyDir string, name string) error {
	// Créer le répertoire si nécessaire
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}
	
	// Sauvegarder la clé privée
	privateKeyBytes := crypto.FromECDSA(kp.PrivateKey)
	privateKeyPath := filepath.Join(keyDir, fmt.Sprintf("%s-private.key", name))
	
	if err := os.WriteFile(privateKeyPath, privateKeyBytes, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	
	// Sauvegarder l'adresse
	addressPath := filepath.Join(keyDir, fmt.Sprintf("%s-address.txt", name))
	if err := os.WriteFile(addressPath, []byte(kp.Address.Hex()), 0644); err != nil {
		return fmt.Errorf("failed to save address: %w", err)
	}
	
	return nil
}

// LoadKeyPairFromFile charge une paire de clés depuis un fichier
func LoadKeyPairFromFile(keyDir string, name string) (*KeyPair, error) {
	privateKeyPath := filepath.Join(keyDir, fmt.Sprintf("%s-private.key", name))
	
	// Lire la clé privée
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	
	// Convertir en clé ECDSA
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	
	publicKey := &privateKey.PublicKey
	address := crypto.PubkeyToAddress(*publicKey)
	
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}
