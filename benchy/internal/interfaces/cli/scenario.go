package cli

import (
	"context"
	"fmt"
	"strconv"

	"benchy/internal/application/handlers"
	"github.com/spf13/cobra"
)

// scenarioCmd représente la commande scenario
var scenarioCmd = &cobra.Command{
	Use:   "scenario [0|1|2|3|init|transfers|erc20|replacement]",
	Short: "Run network test scenarios",
	Long: `Run predefined scenarios to test network behavior:

Scenario 0 (init):        Initialize network with ETH for validators
Scenario 1 (transfers):   Alice sends 0.1 ETH to Bob every 10 seconds  
Scenario 2 (erc20):       Deploy ERC20 token and distribute to Driss/Elena
Scenario 3 (replacement): Test transaction replacement with higher fee`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scenario := args[0]
		
		// Convertir les alias en numéros
		switch scenario {
		case "init":
			scenario = "0"
		case "transfers":
			scenario = "1"
		case "erc20":
			scenario = "2"
		case "replacement":
			scenario = "3"
		}
		
		// Valider le numéro de scénario
		scenarioNum, err := strconv.Atoi(scenario)
		if err != nil || scenarioNum < 0 || scenarioNum > 3 {
			return fmt.Errorf("invalid scenario. Use: 0, 1, 2, 3 or init, transfers, erc20, replacement")
		}
		
		// Créer le handler
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		// Créer le contexte
		ctx := context.Background()

		// Exécuter le scénario
		return handler.HandleScenario(ctx, args[0])
	},
}
