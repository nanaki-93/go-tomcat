package operation

import (
	"fmt"
	"github.com/nanaki-93/go-tomcat/internal/model"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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
	err = os.WriteFile(ts.joinBasePath(runningAppsYamlName), data, os.ModePerm)
	CheckErr(err)
}

func (ts *TomcatManager) CreateTomcat() error {
	if err := os.CopyFS(ts.TomcatPaths.HomeAppTomcat, os.DirFS(ts.joinBasePath("tomcat"))); err != nil {
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

	if err = os.WriteFile(ts.joinBasePath(runningAppsYamlName), runningAppsYaml, os.ModePerm); err != nil {
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
				if err = os.RemoveAll(ts.joinBasePath(folder.Name())); err != nil {
					return fmt.Errorf("RemoveUnusedTomcatFolder : %w", err)
				}
				slog.Info("Tomcat folder removed successfully", "folder", folder.Name())
			}
		}
	}
	return nil
}
func WithRoutine(wg *sync.WaitGroup, fn func(bool) error) {
	defer wg.Done()
	if err := fn(false); err != nil {
		slog.Error("WithRoutine error", "err", err)
	}
}
func (ts *TomcatManager) AddAppsConfigProps(isRetry bool) error {

	if !ts.TomcatConfig.AppConfig.WithAppsConfig {
		if err := os.Remove(ts.TomcatPaths.AppsConfigProps); err != nil {
			return fmt.Errorf("AddAppsConfigProps : %w", err)
		}
		return nil
	}

	if err := ReplaceInFile(ts.TomcatPaths.AppsConfigProps, ts.TomcatPaths.AppsConfigProps, map[string]string{"{{project_path}}": ts.TomcatConfig.AppConfig.ProjectPath}); err != nil {
		return fmt.Errorf("AddAppsConfigProps : %w", err)
	}

	if err := CheckInFile(ts.TomcatPaths.ServerXml, []string{"{{project_path}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("AddAppsConfigProps : %w", err)
		}
		_ = ts.AddAppsConfigProps(true)
	}

	return nil
}

func (ts *TomcatManager) AddDbResources(isRetry bool) error {

	data, err := os.ReadFile(filepath.Join(ts.TomcatPaths.CliBasePath, dbResourcesYamlName))
	if err != nil {
		return fmt.Errorf("addDbResources : %w", err)
	}
	var dbConfig model.DbConfig
	if err = yaml.Unmarshal(data, &dbConfig); err != nil {
		return fmt.Errorf("addDbResources : %w", err)
	}

	dbResourceEnvMap := map[string]string{
		"dev": dbConfig.DbResource.Dev,
		"sit": dbConfig.DbResource.Sit,
	}
	dbResourceToAdd, ok := dbResourceEnvMap[ts.TomcatConfig.EnvToStart]
	if !ok {
		slog.Warn("env unknown, using dev")
		dbResourceToAdd = dbConfig.DbResource.Dev
	}

	if err = ReplaceInFile(ts.TomcatPaths.ServerXml, ts.TomcatPaths.ServerXml, map[string]string{"{{db_resources}}": dbResourceToAdd}); err != nil {
		return fmt.Errorf("addDbResources : %w", err)
	}

	if err = CheckInFile(ts.TomcatPaths.ServerXml, []string{"{{db_resources}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("addDbResources : %w", err)
		}
		_ = ts.AddDbResources(true)
	}

	return nil
}
func (ts *TomcatManager) AddDbContext(isRetry bool) error {

	data, err := os.ReadFile(filepath.Join(ts.TomcatPaths.CliBasePath, dbResourcesYamlName))
	if err != nil {
		return fmt.Errorf("AddDbContext : %w", err)
	}
	var dbConfig model.DbConfig
	if err = yaml.Unmarshal(data, &dbConfig); err != nil {
		return fmt.Errorf("AddDbContext : %w", err)
	}

	dbContextEnvMap := map[string]string{
		"dev": dbConfig.DbContext.Dev,
		"sit": dbConfig.DbContext.Sit,
	}
	dbContextToAdd, ok := dbContextEnvMap[ts.TomcatConfig.EnvToStart]
	if !ok {
		slog.Warn("env unknown, using dev")
		dbContextToAdd = dbConfig.DbContext.Dev
	}

	if err = ReplaceInFile(ts.TomcatPaths.ContextXml, ts.TomcatPaths.ContextXml, map[string]string{"{{db_resources}}": dbContextToAdd}); err != nil {
		return fmt.Errorf("AddDbContext : %w", err)
	}
	if err = CheckInFile(ts.TomcatPaths.ContextXml, []string{"{{db_resources}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("AddDbContext : %w", err)
		}
		_ = ts.AddDbContext(true)
	}
	return nil
}

func (ts *TomcatManager) AddAppContext(isRetry bool) error {

	inputContextPath := ts.joinBasePath("contexts", ts.TomcatConfig.AppConfig.ContextFileName+".xml")
	outputContextPath := filepath.Join(ts.TomcatPaths.CatalinaLocalhost, ts.TomcatConfig.AppConfig.ContextFileName+".xml")

	if err := ReplaceInFile(inputContextPath, outputContextPath, map[string]string{"{{tomcat_deploy_path}}": ts.TomcatPaths.Deploy, "{{war_name}}": ts.TomcatConfig.AppConfig.WarName}); err != nil {
		return fmt.Errorf("addContext : %w", err)
	}

	if err := CheckInFile(outputContextPath, []string{"{{tomcat_deploy_path}}", "{{war_name}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("addContext : %w", err)
		}
		_ = ts.AddAppContext(true)
	}

	return nil
}

func (ts *TomcatManager) SetPortsToServer(isRetry bool) error {

	portsToAdd := map[string]string{
		"{{main_port}}":      fmt.Sprint(ts.TomcatProps.CurrentTomcat.MainPort),
		"{{server_port}}":    fmt.Sprint(ts.TomcatProps.CurrentTomcat.ServerPort),
		"{{connector_port}}": fmt.Sprint(ts.TomcatProps.CurrentTomcat.ConnectorPort),
		"{{redirect_port}}":  fmt.Sprint(ts.TomcatProps.CurrentTomcat.RedirectPort),
	}

	if err := ReplaceInFile(ts.TomcatPaths.ServerXml, ts.TomcatPaths.ServerXml, portsToAdd); err != nil {
		return fmt.Errorf("setPortsToServer : %w", err)
	}

	if err := CheckInFile(ts.TomcatPaths.ServerXml, []string{"{{main_port}}", "{{server_port}}", "{{connector_port}}", "{{redirect_port}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("setPortsToServer : %w", err)
		}
		_ = ts.SetPortsToServer(true)
	}
	return nil
}

func (ts *TomcatManager) AddIndexPage(isRetry bool) error {

	if len(ts.TomcatConfig.AppConfig.IndexFile) == 0 {
		slog.Info("No index Props for the app:", "app", ts.TomcatProps.CurrentTomcat.AppTomcatName)
		return nil
	}
	portsToAdd := map[string]string{
		"{{server_port}}": fmt.Sprint(ts.TomcatProps.CurrentTomcat.ServerPort),
	}

	//replace port in IndexFile
	if err := ReplaceInFile(ts.joinBasePath(ts.TomcatConfig.AppConfig.IndexFile), filepath.Join(ts.TomcatPaths.Deploy, ts.TomcatConfig.AppConfig.IndexFile), portsToAdd); err != nil {
		return fmt.Errorf("AddIndexPage : %w", err)
	}

	if err := CheckInFile(filepath.Join(ts.TomcatPaths.Deploy, ts.TomcatConfig.AppConfig.IndexFile), []string{"{{server_port}}"}); err != nil {
		if isRetry {
			return fmt.Errorf("AddIndexPage : %w", err)
		}
		_ = ts.AddIndexPage(true)
	}
	return nil
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

	if err := os.Setenv("JAVA_HOME", ts.joinBasePath(envConfig.JavaHome)); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	if err := os.Setenv("JRE_HOME", ts.joinBasePath(envConfig.JreHome)); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	if err := os.Setenv("CATALINA_HOME", ts.TomcatPaths.HomeAppTomcat); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}

	javaOpts := envConfig.JavaOpts + " " + ts.TomcatConfig.AppConfig.JavaOpts
	outJavaOpts := strings.ReplaceAll(javaOpts, "{{catalina_home}}", ts.TomcatPaths.HomeAppTomcat)
	outJavaOpts = strings.ReplaceAll(outJavaOpts, "{{context_file_name}}", ts.TomcatConfig.AppConfig.ContextFileName)
	outJavaOpts = strings.ReplaceAll(outJavaOpts, "{{debug_port}}", fmt.Sprint(ts.TomcatProps.CurrentTomcat.DebugPort))
	if err := os.Setenv("JAVA_OPTS", outJavaOpts); err != nil {
		return fmt.Errorf("setSystemEnv : %w", err)
	}
	return nil
}

func (ts *TomcatManager) joinBasePath(suffix ...string) string {
	joinSuffix := filepath.Join(suffix...)
	return filepath.Join(ts.TomcatPaths.CliBasePath, joinSuffix)
}
