package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "kronos",
	Short: "Kronos - Trading infrastructure platform",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
