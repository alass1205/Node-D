package ports

import (
	"context"
	"math/big"
	"benchy/internal/domain/entities"
	"github.com/ethereum/go-ethereum/common"
)

// EthereumService définit les opérations Ethereum
type EthereumService interface {
	// Connexion aux nodes
	ConnectToNode(ctx context.Context, nodeURL string) error
	DisconnectFromNode(ctx context.Context, nodeURL string) error
	IsNodeConnected(ctx context.Context, nodeURL string) (bool, error)
	
	// Informations blockchain
	GetLatestBlockNumber(ctx context.Context, nodeURL string) (uint64, error)
	GetBlockByNumber(ctx context.Context, nodeURL string, blockNumber uint64) (*BlockInfo, error)
	GetPeerCount(ctx context.Context, nodeURL string) (int, error)
	GetPendingTransactionCount(ctx context.Context, nodeURL string) (int, error)
	
	// Gestion des comptes
	GetBalance(ctx context.Context, nodeURL string, address common.Address) (*big.Int, error)
	GetNonce(ctx context.Context, nodeURL string, address common.Address) (uint64, error)
	
	// Transactions
	SendTransaction(ctx context.Context, nodeURL string, tx *entities.Transaction) (common.Hash, error)
	GetTransactionStatus(ctx context.Context, nodeURL string, txHash common.Hash) (entities.TransactionStatus, error)
	GetTransactionReceipt(ctx context.Context, nodeURL string, txHash common.Hash) (*TransactionReceipt, error)
	
	// Smart contracts
	DeployContract(ctx context.Context, nodeURL string, contractCode []byte, from common.Address) (common.Address, common.Hash, error)
	CallContract(ctx context.Context, nodeURL string, contractAddress common.Address, data []byte) ([]byte, error)
	
	// ERC20 tokens
	GetTokenBalance(ctx context.Context, nodeURL string, tokenAddress, holderAddress common.Address) (*big.Int, error)
	TransferToken(ctx context.Context, nodeURL string, tokenAddress, from, to common.Address, amount *big.Int) (common.Hash, error)
}

// BlockInfo représente les informations d'un bloc
type BlockInfo struct {
	Number       uint64
	Hash         common.Hash
	ParentHash   common.Hash
	Timestamp    uint64
	Difficulty   *big.Int
	GasLimit     uint64
	GasUsed      uint64
	Transactions []common.Hash
	Miner        common.Address
}

// TransactionReceipt représente le reçu d'une transaction
type TransactionReceipt struct {
	TransactionHash common.Hash
	BlockNumber     uint64
	BlockHash       common.Hash
	TransactionIndex uint
	From            common.Address
	To              common.Address
	GasUsed         uint64
	Status          uint64
	ContractAddress common.Address
	Logs            []LogEntry
}

// LogEntry représente un log d'événement
type LogEntry struct {
	Address common.Address
	Topics  []common.Hash
	Data    []byte
}
