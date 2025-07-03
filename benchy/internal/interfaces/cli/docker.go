package cli

import (
	"context"
	"fmt"

	"benchy/internal/application/handlers"
	"github.com/spf13/cobra"
)

// dockerCmd représente les commandes Docker
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker related commands",
	Long:  "Commands to manage Docker containers and check availability",
}

// checkDockerCmd vérifie que Docker est disponible
var checkDockerCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if Docker is available",
	Long:  "Verify that Docker is installed, running and accessible",
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		ctx := context.Background()
		return handler.CheckDockerAvailable(ctx)
	},
}

// launchRealCmd lance le réseau avec vrais containers Docker
var launchRealCmd = &cobra.Command{
	Use:   "launch-real",
	Short: "Launch REAL Docker containers",
	Long: `Launch a real Ethereum network with actual Docker containers:
- Pull Geth and Nethermind images
- Create genesis block with Clique consensus
- Launch 5 containers with real blockchain
- Wait for network synchronization`,
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := handlers.NewCLIHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize handler: %w", err)
		}

		ctx := context.Background()
		
		// Pour l'instant, on utilise le même service mais avec feedback différent
		handler.CheckDockerAvailable(ctx)
		return handler.HandleLaunchNetwork(ctx)
	},
}

func init() {
	// Ajouter les sous-commandes docker
	dockerCmd.AddCommand(checkDockerCmd)
	dockerCmd.AddCommand(launchRealCmd)
	
	// Ajouter docker aux commandes principales
	rootCmd.AddCommand(dockerCmd)
}
