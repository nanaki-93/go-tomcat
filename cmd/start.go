// Package cmd /*
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/nanaki-93/go-tomcat/internal/model"
	"github.com/nanaki-93/go-tomcat/internal/operation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	startCmd.Flags().BoolP(skipMavenFlag, "s", false, "if skipMaven is true, maven task is skipped")
	startCmd.Flags().BoolP(offlineFlag, "o", false, "if offline is true, maven will run in offline mode")
	startCmd.Flags().StringP(envFlag, "e", "", "env to start")
	startCmd.Flags().StringP(acquirerFlag, "a", "", "acquirer to start")
}

func validateArgs() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("just one arg is allowed, select one from the following list: %v", validAppList)
		}

		if !slices.Contains(validAppList, args[0]) {
			return fmt.Errorf("arg not valid: %s. select one from the following list: %v", args[0], validAppList)
		}
		return nil
	}
}

func execStartCmd(cmd *cobra.Command, args []string) {
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

	dbResources, err := tm.GetDbResources()
	operation.CheckErr(err)

	dbContext, err := tm.GetDbContext()
	operation.CheckErr(err)

	err = tm.CopyAppContext()
	operation.CheckErr(err)

	fileListToAdd := []string{
		tm.TomcatPaths.ServerXml,
		tm.TomcatPaths.ContextXml,
		filepath.Join(tm.TomcatPaths.CatalinaLocalhost, tm.TomcatConfig.AppConfig.ContextFileName+".xml"),
	}

	acquirer, _ := cmd.Flags().GetString(acquirerFlag)
	acquirerToSet, err := tm.SetAcquirer(acquirer)
	operation.CheckErr(err)

	keysToReplace := map[string]string{
		"{{catalina_home}}":      tm.TomcatPaths.HomeAppTomcat,
		"{{context_file_name}}":  tm.TomcatConfig.AppConfig.ContextFileName,
		"{{debug_port}}":         fmt.Sprint(tm.TomcatProps.CurrentTomcat.DebugPort),
		"{{project_path}}":       tm.TomcatConfig.AppConfig.ProjectPath,
		"{{tomcat_deploy_path}}": tm.TomcatPaths.Deploy,
		"{{war_name}}":           tm.TomcatConfig.AppConfig.WarName,
		"{{main_port}}":          fmt.Sprint(tm.TomcatProps.CurrentTomcat.MainPort),
		"{{server_port}}":        fmt.Sprint(tm.TomcatProps.CurrentTomcat.ServerPort),
		"{{connector_port}}":     fmt.Sprint(tm.TomcatProps.CurrentTomcat.ConnectorPort),
		"{{redirect_port}}":      fmt.Sprint(tm.TomcatProps.CurrentTomcat.RedirectPort),
		"{{db_resources}}":       dbResources,
		"{{db_context}}":         dbContext,
	}

	if acquirerToSet != "" {
		keysToReplace["{{acquirer}}"] = acquirerToSet
	}

	appsConfigFile, err := tm.AddAppsConfigProps()
	operation.CheckErr(err)
	if len(appsConfigFile) > 0 {
		fileListToAdd = append(fileListToAdd, appsConfigFile)
	}

	indexPageFile, err := tm.CopyIndexPage()
	operation.CheckErr(err)
	if len(indexPageFile) > 0 {
		fileListToAdd = append(fileListToAdd, indexPageFile)
	}

	slog.Info("all the resources are added")

	err = operation.UpdatePropsInFiles(fileListToAdd, keysToReplace)
	if err != nil {
		slog.Error("error replacing in file", "error", err)
		return
	}
	slog.Info("all the resources are replaced")

	checkedKeys := make([]string, 0)
	for k := range keysToReplace {
		checkedKeys = append(checkedKeys, k)
	}
	err = operation.CheckInFile(fileListToAdd, checkedKeys)
	if err != nil {
		slog.Error("something went wrong in the replacement process", "error", err)
		return
	}
	err = buildWithMaven(cmd, tm)
	operation.CheckErr(err)

	err = tm.CopyAppToTomcat()
	operation.CheckErr(err)

	err = tm.SetSystemEnv()
	operation.CheckErr(err)

	err = tm.SetJavaOpts(keysToReplace)
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

	tomcatProps, err := operation.LoadTomcatProps(basePath)
	if err != nil {
		return &operation.TomcatManager{}, fmt.Errorf("createTomcatManager : %w", err)
	}

	return operation.NewTomcatManager(generalConfig.WithAppConfig(configSelectedApp),
		&tomcatProps, basePath, appName), nil

}

func buildWithMaven(cmd *cobra.Command, ts *operation.TomcatManager) error {

	offline, _ := cmd.Flags().GetBool(offlineFlag)

	skipMaven, _ := cmd.Flags().GetBool(skipMavenFlag)
	if skipMaven {
		slog.Info("Skipping Maven build")
		return nil
	}
	stCmd := ts.GetMvnCommand(offline)

	err := ts.SetSystemEnv()
	operation.CheckErr(err)

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
