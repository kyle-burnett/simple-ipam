package cmd

import (
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/cmd/add"
	"github.com/kyle-burnett/simple-ipam/internal/cmd/delete"
	"github.com/kyle-burnett/simple-ipam/internal/cmd/initialize"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var rootCmd = &cobra.Command{
	Use:   "simple-ipam",
	Short: "Simple CLI IPAM Tool",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var genDocsCmd = &cobra.Command{
	Use:    "gendocs",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateMarkdownDocs(rootCmd)
	},
}

func generateMarkdownDocs(cmd *cobra.Command) error {
	return doc.GenMarkdownTree(cmd, "./docs")
}

func Execute() {
	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
	rootCmd.AddCommand(initialize.InitCmd)
	rootCmd.AddCommand(genDocsCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
