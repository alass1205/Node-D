package entities

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionStatus représente l'état d'une transaction
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
	TxStatusReplaced  TransactionStatus = "replaced"
)

// TransactionType représente le type de transaction
type TransactionType string

const (
	TxTypeTransfer    TransactionType = "transfer"
	TxTypeContract    TransactionType = "contract"
	TxTypeERC20       TransactionType = "erc20"
	TxTypeReplacement TransactionType = "replacement"
)

// Transaction représente une transaction Ethereum
type Transaction struct {
	Hash     common.Hash       `json:"hash"`
	Type     TransactionType   `json:"type"`
	Status   TransactionStatus `json:"status"`
	
	// Données de la transaction
	From     common.Address `json:"from"`
	To       common.Address `json:"to"`
	Value    *big.Int       `json:"value"`
	Gas      uint64         `json:"gas"`
	GasPrice *big.Int       `json:"gas_price"`
	Nonce    uint64         `json:"nonce"`
	Data     []byte         `json:"data"`
	
	// Informations de bloc
	BlockNumber uint64      `json:"block_number"`
	BlockHash   common.Hash `json:"block_hash"`
	TxIndex     uint        `json:"tx_index"`
	
	// Métriques
	GasUsed          uint64    `json:"gas_used"`
	ConfirmationTime time.Duration `json:"confirmation_time"`
	
	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	ConfirmedAt time.Time `json:"confirmed_at"`
	
	// Référence vers la transaction Go-Ethereum
	EthTx *types.Transaction `json:"-"`
}

// NewTransaction crée une nouvelle transaction
func NewTransaction(from, to common.Address, value *big.Int, txType TransactionType) *Transaction {
	return &Transaction{
		Type:      txType,
		Status:    TxStatusPending,
		From:      from,
		To:        to,
		Value:     value,
		CreatedAt: time.Now(),
	}
}

// UpdateStatus met à jour le statut de la transaction
func (t *Transaction) UpdateStatus(status TransactionStatus) {
	t.Status = status
	if status == TxStatusConfirmed {
		t.ConfirmedAt = time.Now()
		t.ConfirmationTime = t.ConfirmedAt.Sub(t.CreatedAt)
	}
}

// IsPending retourne true si la transaction est en attente
func (t *Transaction) IsPending() bool {
	return t.Status == TxStatusPending
}

// IsConfirmed retourne true si la transaction est confirmée
func (t *Transaction) IsConfirmed() bool {
	return t.Status == TxStatusConfirmed
}

// GetValueETH retourne la valeur en ETH (float64)
func (t *Transaction) GetValueETH() float64 {
	if t.Value == nil {
		return 0.0
	}
	
	// Convertir wei en ETH
	valueETH := new(big.Float).SetInt(t.Value)
	valueETH.Quo(valueETH, big.NewFloat(1e18))
	result, _ := valueETH.Float64()
	return result
}

// GetGasPriceGwei retourne le prix du gas en Gwei
func (t *Transaction) GetGasPriceGwei() float64 {
	if t.GasPrice == nil {
		return 0.0
	}
	
	// Convertir wei en Gwei
	priceGwei := new(big.Float).SetInt(t.GasPrice)
	priceGwei.Quo(priceGwei, big.NewFloat(1e9))
	result, _ := priceGwei.Float64()
	return result
}
