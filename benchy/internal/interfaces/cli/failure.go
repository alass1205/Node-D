package cli

import (
	"context"
	"fmt"
	"strings"

	"benchy/internal/application/handlers"
	"github.com/spf13/cobra"
)

// failureCmd représente la commande temporary-failure
var failureCmd = &cobra.Command{
	Use:   "temporary-failure [alice|bob|cassandra|driss|elena]",
	Short: "Simulate node failure",
	Long: `Simulate a temporary failure by stopping a node for 40 seconds:
- The node will be stopped immediately
- After 40 seconds, it will be restarted automatically
- Use 'benchy infos' to monitor the recovery process`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeName := strings.ToLower(args[0])
		
		// Valider le nom du node
		validNodes := []string{"alice", "bob", "cassandra", "driss", "elena"}
		isValid := false
		for _, valid := range validNodes {
			if nodeName == valid {
				isValid = true
				break
			}
		}
		
		if !isValid {
			return fmt.Errorf("invalid node name '%s'. Valid nodes: alice, bob, cassandra, driss, elena", nodeName)
		}
		
		// Créer le handler
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		// Créer le contexte
		ctx := context.Background()

		// Exécuter la simulation de panne
		return handler.HandleTemporaryFailure(ctx, nodeName)
	},
}
