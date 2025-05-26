/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path/filepath"
)

// initCmd represents the command to initialize the config and folders
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init all the config and directories",
	Long:  `init all the config and directories. It will create the necessary folders and files to start using the cli.`,
	Run:   execInitCmd,
}

func execInitCmd(cmd *cobra.Command, args []string) {
	cmdBaseDir := os.Getenv("GO_TOMCAT_HOME")
	if cmdBaseDir == "" {
		slog.Error("GO_TOMCAT_HOME not set")
		return
	}
	slog.Info("GO_TOMCAT_HOME:", cmdBaseDir)
	_, err := os.Stat(CliBasePath)
	if os.IsNotExist(err) {
		err = os.CopyFS(CliBasePath, os.DirFS(filepath.Join(cmdBaseDir, "resources")))
		operation.CheckErr(err, "Error copying resources folder")
	} else {
		slog.Warn("cli folder already exists")
	}
	slog.Info("all the config and directories are created")
}

func init() {
	rootCmd.AddCommand(initCmd)
}
