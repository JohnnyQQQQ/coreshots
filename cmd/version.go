package cmd

import (
	"coreshots/pkg/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints the current version of the cli",
	Run: func(cmd *cobra.Command, args []string) {
		println(version.Current)
	},
}
