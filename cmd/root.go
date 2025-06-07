/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"

	"path/filepath"
)

var CliBasePath string
var validAppList []string

var validEnvList = []string{DevEnv, SitEnv}

const (
	skipMavenFlag = "skipMaven"
	envFlag       = "env"
	DevEnv        = "dev"
	SitEnv        = "sit"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gtom",
	Short: "cli to start some applications with tomcat",
	Long: `go-tomcat is a cli to start some applications with tomcat.
It allows you to start, stop, update and init the tomcat server with the specified app and env.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	setCliBasePath()
	rootCmd.AddCommand(completionCmd)
}

// redefine the completion command to make it hidden
var completionCmd = &cobra.Command{
	Use:    "completion",
	Short:  "completion command",
	Long:   `redifine the completion command to generate shell completion scripts`,
	Hidden: true,
}

func setCliBasePath() {
	userHome, err := os.UserHomeDir()
	operation.CheckErr(err, "Error getting user home dir")
	CliBasePath = filepath.Join(userHome, ".go-tomcat")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Search config in home directory with name ".go-tomcat" (without extension).
	viper.AddConfigPath(CliBasePath)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".go-tomcat")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Error reading config file: %v\n", err)
	}

	validAppList = viper.GetStringSlice("apps")

}
