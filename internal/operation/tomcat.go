package operation

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nanaki-93/go-tomcat/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	goTomcatPrefix      = "go-tomcat-"
	runningAppsYamlName = ".running-tomcats.yaml"
	dbResourcesYamlName = ".db-resources.yaml"
	hostToCheck         = "0.0.0.0"
	StartMainPort       = 9000
	StartServerPort     = 8000
	StartDebugPort      = 5000
	StartConnectorPort  = 8100
	StartRedirectPort   = 8400
)

type TomcatManager struct {
	TomcatConfig *model.TomcatGlobalConfig
	TomcatProps  *model.TomcatProps
	TomcatPaths  *model.TomcatPaths
}

func NewTomcatManager(config *model.TomcatGlobalConfig, tomcatProps *model.TomcatProps, cliBasePath, appName string) *TomcatManager {
	return &TomcatManager{
		TomcatConfig: config,
		TomcatProps:  tomcatProps,
		TomcatPaths:  model.GetTomcatPaths(cliBasePath, appName),
	}
}

func CreateTomcatProps(basePath string) (model.TomcatProps, error) {

	tomcat := model.TomcatProps{}

	data, err := os.ReadFile(filepath.Join(basePath, runningAppsYamlName))
	if err != nil {
		return model.TomcatProps{}, fmt.Errorf("SetRunningAppsConfig : %w", err)
	}

	if err = yaml.Unmarshal(data, &tomcat); err != nil {
		return model.TomcatProps{}, fmt.Errorf("SetRunningAppsConfig : %w", err)
	}
	return tomcat, nil
}

func (ts *TomcatManager) RemoveFromRunningAppsConfig() error {
	changed := false
	for _, tomcat := range ts.TomcatProps.RunningTomcats {

		if isFreePort(fmt.Sprint(tomcat.ServerPort)) {
			ts.TomcatProps.RunningTomcats = RemoveTomcatFromRunning(ts.TomcatProps.RunningTomcats, tomcat.AppTomcatName)
			changed = true
		}
	}
	if changed {
		ts.UpdateAppRunningYaml()
	}
	return nil
}

func (ts *TomcatManager) RemoveCurrentFromRunningAppsConfig() error {

	ts.TomcatProps.RunningTomcats = RemoveTomcatFromRunning(ts.TomcatProps.RunningTomcats, ts.TomcatProps.CurrentTomcat.AppTomcatName)
	ts.UpdateAppRunningYaml()

	return nil
}

func (ts *TomcatManager) SetTomcatPorts() {

	nextServerPort := findNextPort(sliceToSlice(ts.TomcatProps.RunningTomcats, func(t model.Tomcat) int { return t.ServerPort }), StartServerPort)
	nextRedirectPort := findNextPort(sliceToSlice(ts.TomcatProps.RunningTomcats, func(t model.Tomcat) int { return t.RedirectPort }), StartRedirectPort)
	nextMainPort := findNextPort(sliceToSlice(ts.TomcatProps.RunningTomcats, func(t model.Tomcat) int { return t.MainPort }), StartMainPort)
	nextDebugPort := findNextPort(sliceToSlice(ts.TomcatProps.RunningTomcats, func(t model.Tomcat) int { return t.DebugPort }), StartDebugPort)
	nextConnectorPort := findNextPort(sliceToSlice(ts.TomcatProps.RunningTomcats, func(t model.Tomcat) int { return t.ConnectorPort }), StartConnectorPort)

	ts.TomcatProps.CurrentTomcat = model.Tomcat{
		AppTomcatName: ts.TomcatPaths.AppTomcatName,
		MainPort:      nextMainPort,
		ServerPort:    nextServerPort,
		DebugPort:     nextDebugPort,
		ConnectorPort: nextConnectorPort,
		RedirectPort:  nextRedirectPort,
	}
	slog.Info("Tomcat ports set", "tomcat", ts.TomcatProps.CurrentTomcat)
}

func findNextPort(usedPorts []int, startPort int) int {
	nextServerPort := startPort
	for i := 0; ; i++ {
		nextServerPort = startPort + i
		if !slices.Contains(usedPorts, nextServerPort) {
			if isFreePort(fmt.Sprint(nextServerPort)) {
				slog.Info("found free port", "nextServerPort", nextServerPort)
				break
			}
		}
	}
	return nextServerPort
}

func RemoveTomcatFromRunning(runningTomcats []model.Tomcat, appTomcatNameToRemove string) []model.Tomcat {
	return slices.DeleteFunc(
		runningTomcats,
		func(tom model.Tomcat) bool {
			return tom.AppTomcatName == appTomcatNameToRemove
		})
}

func (ts *TomcatManager) UpdateAppRunningYaml() {
	data, err := yaml.Marshal(ts.TomcatProps)
	CheckErr(err)
	err = os.WriteFile(ts.JoinBasePath(runningAppsYamlName), data, os.ModePerm)
	CheckErr(err)
}

func (ts *TomcatManager) CreateTomcat() error {
	if err := os.CopyFS(ts.TomcatPaths.HomeAppTomcat, os.DirFS(ts.JoinBasePath("tomcat"))); err != nil {
		return fmt.Errorf("CreateTomcat : %w", err)
	}
	return nil
}

func (ts *TomcatManager) RunTomcat() error {

	if err := ts.addTomcatToRunningApps(); err != nil {
		return fmt.Errorf("RunTomcat : %w", err)
	}
	stCmd := exec.Command(ts.TomcatPaths.CatalinaBat, "run")
	PrintCmd(stCmd)
	if err := stCmd.Run(); err != nil {
		return fmt.Errorf("RunTomcat : %w", err)
	}
	return nil
}

func (ts *TomcatManager) addTomcatToRunningApps() error {
	//write runningAppsConfig to .running-apps.yaml

	ts.TomcatProps.RunningTomcats = append(ts.TomcatProps.RunningTomcats, ts.TomcatProps.CurrentTomcat)
	runningAppsYaml, err := yaml.Marshal(ts.TomcatProps)
	if err != nil {
		return fmt.Errorf("addTomcatToRunningApps : %w", err)
	}

	if err = os.WriteFile(ts.JoinBasePath(runningAppsYamlName), runningAppsYaml, os.ModePerm); err != nil {
		return fmt.Errorf("addTomcatToRunningApps : %w", err)
	}
	return nil
}

func (ts *TomcatManager) RemoveUnusedTomcatFolder() error {

	baseFolderList, err := os.ReadDir(ts.TomcatPaths.CliBasePath)
	if err != nil {
		return fmt.Errorf("RemoveUnusedTomcatFolder : %w", err)
	}

	folderToAvoid := make(map[string]struct{}, len(ts.TomcatProps.RunningTomcats))
	for _, tomcat := range ts.TomcatProps.RunningTomcats {
		folderToAvoid[goTomcatPrefix+tomcat.AppTomcatName] = struct{}{}
	}

	for _, folder := range baseFolderList {
		name := folder.Name()
		if strings.Contains(name, goTomcatPrefix) {
			if _, avoid := folderToAvoid[name]; !avoid {
				if err = os.RemoveAll(ts.JoinBasePath(folder.Name())); err != nil {
					return fmt.Errorf("RemoveUnusedTomcatFolder : %w", err)
				}
				slog.Info("Tomcat folder removed successfully", "folder", folder.Name())
			}
		}
	}
	return nil
}

func (ts *TomcatManager) AddAppsConfigProps() (string, error) {

	if !ts.TomcatConfig.AppConfig.WithAppsConfig {
		if err := os.Remove(ts.TomcatPaths.AppsConfigProps); err != nil {
			return "", fmt.Errorf("AddAppsConfigProps : %w", err)
		}
		return "", nil
	}
	return ts.TomcatPaths.AppsConfigProps, nil
}

func (ts *TomcatManager) GetDbResources() (string, error) {

	data, err := os.ReadFile(filepath.Join(ts.TomcatPaths.CliBasePath, dbResourcesYamlName))
	if err != nil {
		return "", fmt.Errorf("addDbResources : %w", err)
	}
	var dbConfig model.DbConfig
	if err = yaml.Unmarshal(data, &dbConfig); err != nil {
		return "", fmt.Errorf("addDbResources : %w", err)
	}

	dbResourceEnvMap := map[string]string{
		"local": dbConfig.DbResource.Local,
		"dev":   dbConfig.DbResource.Dev,
		"sit":   dbConfig.DbResource.Sit,
	}
	dbResourceToAdd, ok := dbResourceEnvMap[ts.TomcatConfig.EnvToStart]
	if !ok {
		slog.Warn("env unknown, using dev")
		dbResourceToAdd = dbConfig.DbResource.Dev
	}

	return dbResourceToAdd, nil

}
func (ts *TomcatManager) GetDbContext() (string, error) {

	data, err := os.ReadFile(filepath.Join(ts.TomcatPaths.CliBasePath, dbResourcesYamlName))
	if err != nil {
		return "", fmt.Errorf("GetDbContext : %w", err)
	}
	var dbConfig model.DbConfig
	if err = yaml.Unmarshal(data, &dbConfig); err != nil {
		return "", fmt.Errorf("GetDbContext : %w", err)
	}

	dbContextEnvMap := map[string]string{
		"local": dbConfig.DbContext.Local,
		"dev":   dbConfig.DbContext.Dev,
		"sit":   dbConfig.DbContext.Sit,
	}
	dbContextToAdd, ok := dbContextEnvMap[ts.TomcatConfig.EnvToStart]
	if !ok {
		slog.Warn("env unknown, using dev")
		dbContextToAdd = dbConfig.DbContext.Dev
	}

	return dbContextToAdd,
		nil

}

func (ts *TomcatManager) CopyAppContext() error {

	inputContextPath := ts.JoinBasePath("contexts", ts.TomcatConfig.AppConfig.ContextFileName+".xml")
	outputContextPath := filepath.Join(ts.TomcatPaths.CatalinaLocalhost, ts.TomcatConfig.AppConfig.ContextFileName+".xml")

	data, err := os.ReadFile(inputContextPath)
	if err != nil {
		return fmt.Errorf("CopyAppContext ReadFile: %w", err)
	}
	err = os.WriteFile(outputContextPath, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("CopyAppContext WriteFile: %w", err)
	}
	return nil

}

func (ts *TomcatManager) CopyIndexPage() (string, error) {

	if len(ts.TomcatConfig.AppConfig.IndexFile) == 0 {
		slog.Info("No index Props for the app:", "app", ts.TomcatProps.CurrentTomcat.AppTomcatName)
		return "", nil
	}

	data, err := os.ReadFile(ts.JoinBasePath(ts.TomcatConfig.AppConfig.IndexFile))
	if err != nil {
		return "", fmt.Errorf("CopyIndexPage ReadFile: %w", err)
	}
	err = os.WriteFile(filepath.Join(ts.TomcatPaths.Deploy, ts.TomcatConfig.AppConfig.IndexFile), data, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("CopyIndexPage WriteFile: %w", err)
	}

	return filepath.Join(ts.TomcatPaths.Deploy, ts.TomcatConfig.AppConfig.IndexFile), nil

}

func (ts *TomcatManager) CopyAppToTomcat() error {

	appConfig := ts.TomcatConfig.AppConfig

	entries, err := os.ReadDir(filepath.Join(appConfig.ProjectPath, appConfig.TargetSuffix))
	if err != nil {
		return fmt.Errorf("copyAppToTomcat : %w", err)
	}
	var targetAppToCopy string
	for _, e := range entries {
		if (strings.Contains(e.Name(), appConfig.WarName)) && (strings.Contains(e.Name(), ".war")) {
			targetAppToCopy = filepath.Join(appConfig.ProjectPath, appConfig.TargetSuffix, e.Name())
		}
	}

	if err = os.CopyFS(filepath.Join(ts.TomcatPaths.Deploy, appConfig.WarName+".war"), os.DirFS(targetAppToCopy)); err != nil {
		return fmt.Errorf("copyAppToTomcat : %w", err)
	}
	return nil
}

func (ts *TomcatManager) SetSystemEnv() error {

	envConfig := ts.TomcatConfig.Env

	if err := os.Setenv("JAVA_HOME", ts.JoinBasePath(envConfig.JavaHome)); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	if err := os.Setenv("JRE_HOME", ts.JoinBasePath(envConfig.JreHome)); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	if err := os.Setenv("CATALINA_HOME", ts.TomcatPaths.HomeAppTomcat); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}

	return nil
}

func (ts *TomcatManager) SetJavaOpts(keyToReplace map[string]string) error {

	envConfig := ts.TomcatConfig.Env

	javaOpts := envConfig.JavaOpts + " " + ts.TomcatConfig.AppConfig.JavaOpts
	javaOpts = replaceKeysInString(javaOpts, keyToReplace)
	if err := os.Setenv("JAVA_OPTS", javaOpts); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	return nil
}
func replaceKeysInString(input string, keysToReplace map[string]string) string {
	out := input
	for oldStr, newStr := range keysToReplace {
		out = strings.ReplaceAll(out, oldStr, newStr)
	}
	return out
}

func (ts *TomcatManager) JoinBasePath(suffix ...string) string {
	joinSuffix := filepath.Join(suffix...)
	return filepath.Join(ts.TomcatPaths.CliBasePath, joinSuffix)
}
