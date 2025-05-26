/*
Copyright © 2025 Marco Andreose <andreose.marco93@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/nanaki-93/go-tomcat/internal/model"
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
)

// startCmd represents the master command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the tomcat server",
	Long:  `start the tomcat server. It will create a new tomcat instance with the specified app and env.`,
	Run:   execStartCmd,
	Args:  validateArgs(),
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP(skipMavenFlag, "s", false, "if skipMaven is true, maven task is skipped ")
	startCmd.Flags().StringP(envFlag, "e", "", "env to start")
}

func validateArgs() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("just one arg is allowed")
		}

		if !slices.Contains(validAppList, args[0]) {
			return fmt.Errorf("arg not noy valid: %s. select of from the following list: %v", args[0], validAppList)
		}
		return nil
	}
}

func execStartCmd(cmd *cobra.Command, args []string) {
	wg := new(sync.WaitGroup)

	var err error

	tm, err := createTomcatManager(CliBasePath, args[0])
	operation.CheckErr(err)

	tm.TomcatConfig.EnvToStart = setEnvToStart(cmd)

	checkInterrupt(tm)

	err = tm.RemoveFromRunningAppsConfig()
	operation.CheckErr(err)
	//Clean Tomcat folders if exists and they are not running
	err = tm.RemoveUnusedTomcatFolder()
	operation.CheckErr(err)

	tm.SetTomcatPorts()

	err = tm.CreateTomcat()
	operation.CheckErr(err)

	wg.Add(6)
	go operation.WithRoutine(wg, tm.AddAppsConfigProps)
	go operation.WithRoutine(wg, tm.AddDbResources)
	go operation.WithRoutine(wg, tm.AddDbContext)
	go operation.WithRoutine(wg, tm.AddAppContext)
	go operation.WithRoutine(wg, tm.SetPortsToServer)
	go operation.WithRoutine(wg, tm.AddIndexPage)
	wg.Wait()

	slog.Info("all the resources are added")

	err = buildWithMaven(cmd, tm)
	operation.CheckErr(err)

	err = tm.CopyAppToTomcat()
	operation.CheckErr(err)

	err = tm.SetSystemEnv()
	operation.CheckErr(err)

	err = tm.RunTomcat()
	operation.CheckErr(err)

}

func checkInterrupt(tomcatService *operation.TomcatManager) {
	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		cleanupFunction(tomcatService)
		os.Exit(0)
	}()
}
func cleanupFunction(ts *operation.TomcatManager) {
	slog.Info("CLEANING UP RESOURCES...")

	err := ts.RemoveCurrentFromRunningAppsConfig()
	operation.CheckErr(err)
	ts.UpdateAppRunningYaml()
}

func createTomcatManager(basePath, appName string) (*operation.TomcatManager, error) {

	configSelectedApp, err := model.GetAppConfig(appName)
	if err != nil {
		return &operation.TomcatManager{}, fmt.Errorf("createTomcatManager : %w", err)
	}

	generalConfig := model.TomcatGlobalConfig{}
	if err := viper.Unmarshal(&generalConfig); err != nil {
		return &operation.TomcatManager{}, fmt.Errorf("createTomcatManager : %w", err)
	}

	tomcatProps, err := operation.CreateTomcatProps(basePath)
	if err != nil {
		return &operation.TomcatManager{}, fmt.Errorf("createTomcatManager : %w", err)
	}

	return operation.NewTomcatManager(generalConfig.WithAppConfig(configSelectedApp),
		&tomcatProps, basePath, appName), nil

}

func buildWithMaven(cmd *cobra.Command, ts *operation.TomcatManager) error {

	skipMaven, _ := cmd.Flags().GetBool(skipMavenFlag)
	if skipMaven {
		slog.Info("Skipping Maven build")
		return nil
	}

	stCmd := exec.Command("mvn.cmd",
		"clean",
		"install",
		"-f",
		ts.TomcatConfig.AppConfig.ProjectPath,
		"-s",
		filepath.Join(ts.TomcatPaths.CliBasePath, ts.TomcatConfig.Env.MvnSettings),
		"-Denv=tom",
		"-DskipTests")

	operation.PrintCmd(stCmd)
	if err := stCmd.Run(); err != nil {
		return fmt.Errorf("buildWithMaven : %w", err)
	}
	return nil
}

func setEnvToStart(cmd *cobra.Command) string {
	envToStart, _ := cmd.Flags().GetString(envFlag)

	if slices.Contains(validEnvList, envToStart) {
		return envToStart
	}
	slog.Warn("no env flag, using Dev env")
	return DevEnv
}
