package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Flag global pour l'option -u (update interval)
	updateInterval int
)

// rootCmd représente la commande de base quand appelée sans sous-commandes
var rootCmd = &cobra.Command{
	Use:   "benchy",
	Short: "Ethereum network benchmarking tool",
	Long: `Benchy is a CLI tool for launching, monitoring and benchmarking Ethereum networks.

It allows you to:
- Launch private Ethereum networks with multiple clients
- Monitor nodes performance in real-time  
- Run scenarios to test network behavior
- Simulate failures to test resilience`,
	Version: "1.0.0",
}

// Execute ajoute toutes les commandes enfants à la commande root et configure les flags appropriés.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag global pour l'option -u
	rootCmd.PersistentFlags().IntVarP(&updateInterval, "update", "u", 0, 
		"Update interval in seconds for continuous monitoring (0 = no update)")

	// Ajouter toutes les sous-commandes
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(infosCmd)
	rootCmd.AddCommand(scenarioCmd)
	rootCmd.AddCommand(failureCmd)
}

// initConfig lit la configuration depuis un fichier config et les variables d'environnement
func initConfig() {
	// Chercher la config dans le répertoire home
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Chercher dans le répertoire de travail actuel également
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.SetConfigName(".benchy")
	viper.SetConfigType("yaml")

	// Lire les variables d'environnement
	viper.AutomaticEnv()

	// Si un fichier config est trouvé, le lire
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
