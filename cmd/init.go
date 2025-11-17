/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
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
		return
	}

	err = operation.CheckCopiedFiles(filepath.Join(cmdBaseDir, "resources"), CliBasePath)
	operation.CheckErr(err, "Some Files didn't get copied, clean and try the init command again. "+
		"If you still have the problem, copy the resource folder manually.")

	slog.Info("all the config and directories are copied to cli folder")

	SetProjectPathAndMvnRepository(err)
}

func SetProjectPathAndMvnRepository(err error) {
	projectDirectory := operation.StringPrompt("Insert your project base directory:")
	projectDirectory = strings.ReplaceAll(projectDirectory, "\\", "/")
	_, err = os.Stat(projectDirectory)
	operation.CheckErr(err, "Project directory does not exist")

	mvnRepository := operation.StringPrompt("Insert your maven repository path:")
	mvnRepository = strings.ReplaceAll(mvnRepository, "\\", "/")
	_, err = os.Stat(mvnRepository)
	operation.CheckErr(err, "Maven repository does not exist")

	filesToReplace := []string{
		filepath.Join(CliBasePath, ".go-tomcat.yaml"),
		filepath.Join(CliBasePath, "mvn-settings.xml")}

	keysToReplace := map[string]string{
		"{{project_base_path}}":   projectDirectory,
		"{{mvn_repository_path}}": mvnRepository}

	err = operation.UpdatePropsInFiles(filesToReplace, keysToReplace)
	operation.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(initCmd)
}
