package cli

import (
	"context"
	"fmt"

	"benchy/internal/application/handlers"
	"github.com/spf13/cobra"
)

// infosCmd représente la commande infos
var infosCmd = &cobra.Command{
	Use:   "infos",
	Short: "Display information about network nodes",
	Long: `Display comprehensive information about each node in the network:
- Latest block number
- Connected peers
- Mempool transactions count
- CPU and memory consumption
- Ethereum address and balance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Créer le handler
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		// Créer le contexte
		ctx := context.Background()

		// Exécuter le monitoring
		return handler.HandleInfos(ctx, updateInterval)
	},
}
