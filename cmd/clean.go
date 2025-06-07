/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

// cleanCmd represents the command to initialize the config and folders
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean all the config and directories",
	Long:  `clean all the config and directories. It will remove all the configuration and directories.`,
	Run:   execCleanCmd,
}

func execCleanCmd(cmd *cobra.Command, args []string) {

	_, err := os.Stat(CliBasePath)
	operation.CheckErr(err, "Directory does not exist.")

	res := operation.YesNoPrompt("Do you wanna delete "+CliBasePath+"? ", false)
	if res {
		err = os.RemoveAll(CliBasePath)
		operation.CheckErr(err, "Error copying resources folder")
		slog.Info("all the config and directories removed")
	}
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
