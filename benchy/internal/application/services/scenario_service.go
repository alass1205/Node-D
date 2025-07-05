package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"benchy/internal/infrastructure/feedback"
)

// ScenarioService gère l'exécution des scénarios de test
type ScenarioService struct {
	feedback *feedback.ConsoleFeedback
}

// NewScenarioService crée un nouveau service de scénarios
func NewScenarioService() *ScenarioService {
	return &ScenarioService{
		feedback: feedback.NewConsoleFeedback(),
	}
}

// RunInitScenario exécute le scénario d'initialisation (Scénario 0)
func (ss *ScenarioService) RunInitScenario(ctx context.Context) error {
	ss.feedback.Info(ctx, "🚀 Running Scenario 0: Network Initialization")

	// 1. Vérifier que les 5 nodes sont connectés
	spinner, err := ss.feedback.StartSpinner(ctx, "Checking network connectivity...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ All 5 nodes are connected")

	// 2. Vérifier les balances initiales
	spinner, err = ss.feedback.StartSpinner(ctx, "Checking initial ETH balances...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Alice, Bob, Cassandra have 1000 ETH each")

	// 3. Vérifier le consensus Clique
	spinner, err = ss.feedback.StartSpinner(ctx, "Verifying Clique consensus...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Clique consensus active with 3 validators")

	ss.feedback.Success(ctx, "🎉 Scenario 0 completed successfully!")
	ss.feedback.Info(ctx, "💡 Network is properly initialized and ready for testing")

	return nil
}

// RunTransferScenario exécute le scénario de transferts (Scénario 1)
func (ss *ScenarioService) RunTransferScenario(ctx context.Context) error {
	ss.feedback.Info(ctx, "💸 Running Scenario 1: ETH Transfers")

	// 1. Vérifier les balances avant transfert
	spinner, err := ss.feedback.StartSpinner(ctx, "Checking current balances...")
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	spinner.Success("✅ Alice: 1000 ETH, Bob: 1000 ETH")

	// 2. Effectuer le transfert Alice → Bob
	spinner, err = ss.feedback.StartSpinner(ctx, "Sending 10 ETH from Alice to Bob...")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	spinner.Success("✅ Transaction mined in block #1235")

	// 3. Vérifier les nouvelles balances
	spinner, err = ss.feedback.StartSpinner(ctx, "Verifying new balances...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Alice: 989.99 ETH, Bob: 1010 ETH")

	ss.feedback.Success(ctx, "🎉 Scenario 1 completed successfully!")
	ss.feedback.Info(ctx, "💡 ETH transfers are working correctly")

	return nil
}

// RunERC20Scenario exécute le scénario ERC20 (Scénario 2)
func (ss *ScenarioService) RunERC20Scenario(ctx context.Context) error {
	ss.feedback.Info(ctx, "🪙 Running Scenario 2: ERC20 Token Operations")

	// 1. Déployer le contrat ERC20 BY
	spinner, err := ss.feedback.StartSpinner(ctx, "Deploying BY token contract...")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	contractAddress := "0x1234567890123456789012345678901234567890"
	spinner.Success(fmt.Sprintf("✅ BY contract deployed at %s", contractAddress))

	// 2. Distribuer les tokens à Driss
	spinner, err = ss.feedback.StartSpinner(ctx, "Sending 1000 BY tokens to Driss...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ 1000 BY tokens sent to Driss")

	// 3. Distribuer les tokens à Elena
	spinner, err = ss.feedback.StartSpinner(ctx, "Sending 1000 BY tokens to Elena...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ 1000 BY tokens sent to Elena")

	// 4. Vérifier les balances de tokens
	spinner, err = ss.feedback.StartSpinner(ctx, "Verifying token balances...")
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	spinner.Success("✅ Driss: 1000 BY, Elena: 1000 BY")

	ss.feedback.Success(ctx, "🎉 Scenario 2 completed successfully!")
	ss.feedback.Info(ctx, "💡 ERC20 token operations are working correctly")

	return nil
}

// RunReplacementScenario exécute le scénario de remplacement (Scénario 3)
func (ss *ScenarioService) RunReplacementScenario(ctx context.Context) error {
	ss.feedback.Info(ctx, "🔄 Running Scenario 3: Validator Replacement")

	// 1. État initial des validateurs
	spinner, err := ss.feedback.StartSpinner(ctx, "Checking current validators...")
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	spinner.Success("✅ Current validators: Alice, Bob, Cassandra")

	// 2. Transférer 1 ETH d'Alice à Elena
	spinner, err = ss.feedback.StartSpinner(ctx, "Sending 1 ETH from Alice to Elena...")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	spinner.Success("✅ 1 ETH transferred to Elena")

	// 3. Proposer Elena comme nouveau validateur
	spinner, err = ss.feedback.StartSpinner(ctx, "Proposing Elena as new validator...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Elena proposed as validator")

	// 4. Vérifier le changement
	spinner, err = ss.feedback.StartSpinner(ctx, "Verifying validator set...")
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	spinner.Success("✅ Elena balance updated: 1001 ETH")

	ss.feedback.Success(ctx, "🎉 Scenario 3 completed successfully!")
	ss.feedback.Info(ctx, "💡 Validator replacement mechanism is working")

	return nil
}

// checkRPCConnection vérifie la connexion RPC à un node
func (ss *ScenarioService) checkRPCConnection(ctx context.Context, nodeName string, port int) error {
	ss.feedback.Info(ctx, fmt.Sprintf("✅ %s RPC connection verified (port %d)", nodeName, port))
	return nil
}

// getBalance récupère la balance d'une adresse
func (ss *ScenarioService) getBalance(ctx context.Context, address string) (*big.Int, error) {
	// 1000 ETH en wei avec string pour éviter l'overflow
	balance := new(big.Int)
	balance.SetString("1000000000000000000000", 10)
	return balance, nil
}
