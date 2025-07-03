package cli

import (
	"context"
	"fmt"

	"benchy/internal/application/handlers"
	"github.com/spf13/cobra"
)

// launchCmd représente la commande launch-network
var launchCmd = &cobra.Command{
	Use:   "launch-network",
	Short: "Launch a private Ethereum network",
	Long: `Launch a private Ethereum network with 5 nodes:
- Alice, Bob, Cassandra (validators)
- Driss, Elena (normal nodes)
- Mix of Geth and Nethermind clients
- Clique consensus algorithm`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Créer le handler
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		// Créer le contexte
		ctx := context.Background()

		// Exécuter le lancement du réseau
		return handler.HandleLaunchNetwork(ctx)
	},
}
