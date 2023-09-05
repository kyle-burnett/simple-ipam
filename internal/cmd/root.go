package cmd

import (
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/cmd/add"
	"github.com/kyle-burnett/simple-ipam/internal/cmd/delete"
	"github.com/kyle-burnett/simple-ipam/internal/cmd/initialize"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "simple-ipam",
	Short: "A brief description of your application",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
	rootCmd.AddCommand(initialize.InitCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
