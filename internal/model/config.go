package model

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

const GoTomcatPrefix = "go-tomcat-"

type TomcatProps struct {
	RunningTomcats []Tomcat `yaml:"running_tomcats"`
	CurrentTomcat  Tomcat   `yaml:"-"`
}
type Tomcat struct {
	AppTomcatName string `yaml:"app_tomcat_name"`
	MainPort      int    `yaml:"main_port"`
	ServerPort    int    `yaml:"server_port"`
	DebugPort     int    `yaml:"debug_port"`
	ConnectorPort int    `yaml:"-"`
	RedirectPort  int    `yaml:"-"`
}

type Acquirers struct {
	Acquirers map[string]Acquirer `yaml:"acquirers"`
}
type Acquirer struct {
	Dev string `yaml:"dev"`
	Sit string `yaml:"sit"`
	Uat string `yaml:"uat"`
}

type TomcatGlobalConfig struct {
	Env        EnvConfig `mapstructure:"env"`
	AppConfig  AppConfig
	EnvToStart string
}

func (c TomcatGlobalConfig) WithAppConfig(appConfig AppConfig) *TomcatGlobalConfig {
	c.AppConfig = appConfig
	return &c
}

type EnvConfig struct {
	MvnSettings string `mapstructure:"mvn_settings"`
	JavaHome    string `mapstructure:"java_home"`
	JreHome     string `mapstructure:"jre_home"`
	JavaOpts    string `mapstructure:"java_opts"`
}
type AppConfig struct {
	ContextFileName string `mapstructure:"context_file_name"`
	WarName         string `mapstructure:"war_name"`
	ProjectPath     string `mapstructure:"project_path"`
	TargetSuffix    string `mapstructure:"target_suffix"`
	JavaOpts        string `mapstructure:"java_opts"`
	WithAppsConfig  bool   `mapstructure:"with_apps_config"`
	WithAcquirer    bool   `mapstructure:"with_acquirer"`
	IndexFile       string `mapstructure:"index_file"`
}

func GetAppConfig(appName string) (AppConfig, error) {
	var cfg AppConfig
	key := fmt.Sprintf("app.%s", appName)
	err := viper.UnmarshalKey(key, &cfg)
	if err != nil {
		return AppConfig{}, fmt.Errorf("unable to decode into struct, %v", err)
	}
	return cfg, nil
}

type TomcatPaths struct {
	CliBasePath       string
	BaseTomcat        string
	AppTomcatName     string
	HomeAppTomcat     string
	ServerXml         string
	ContextXml        string
	AppsConfigProps   string
	CatalinaLocalhost string
	Deploy            string
	CatalinaBat       string
}

func GetTomcatPaths(basePath, appTomcatName string) *TomcatPaths {
	p := TomcatPaths{}
	p.CliBasePath = basePath
	p.AppTomcatName = appTomcatName
	p.HomeAppTomcat = filepath.Join(basePath, GoTomcatPrefix+appTomcatName)
	p.ServerXml = filepath.Join(p.HomeAppTomcat, "conf", "server.xml")
	p.ContextXml = filepath.Join(p.HomeAppTomcat, "conf", "context.xml")
	p.AppsConfigProps = filepath.Join(p.HomeAppTomcat, "apps-config", "backend.properties")
	p.CatalinaLocalhost = filepath.Join(p.HomeAppTomcat, "conf", "Catalina", "localhost")
	p.Deploy = filepath.Join(p.HomeAppTomcat, "deploy")
	p.CatalinaBat = filepath.Join(p.HomeAppTomcat, "bin", "catalina.bat")
	return &p
}
