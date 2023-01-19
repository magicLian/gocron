package setting

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
	"gopkg.in/ini.v1"
)

type Scheme string

var (
	Env             = DEV
	ROLE            = ROLE_MASTER
	LogxLevel       = logx.InfoLevel
	HomePath        string
	configFiles     []string
	AppUrl          string
	HttpPort        string
	ApplicationName string
	APIVersion      string
)

const (
	ROLE_MASTER = "master"
	ROLE_WORKER = "worker"
	DEV         = "dev"
	PROD        = "prod"
)

type Cfg struct {
	Raw *ini.File

	log logx.Logger

	DataPath string
	LogsPath string
}

type CommandLineArgs struct {
	Args []string
}

func ProvideSettingCfg(args *CommandLineArgs) (*Cfg, error) {
	cfg := &Cfg{
		Raw: ini.Empty(),
	}
	if err := cfg.Load(args); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Cfg) Load(args *CommandLineArgs) error {
	ROLE = utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "role"), ROLE_MASTER)
	SetHomePath(args)

	_, err := cfg.LoadConfiguration(args)
	if err != nil {
		cfg.log.Error(err.Error())
		return err
	}

	logLevel := utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "log", "level"), "info")
	logxLevel, err := logx.ParseLevel(logLevel)
	if err != nil {
		cfg.log.Error(err.Error())
		return err
	}
	LogxLevel = logxLevel
	cfg.log = logx.NewLogx("cfg", logxLevel)

	Env = utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "development"), DEV)
	HttpPort = utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "http_port"), "8080")
	APIVersion = utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "apiversion"), "0.0.1")

	return nil
}

func SetHomePath(args *CommandLineArgs) error {
	HomePath, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	confFile := fmt.Sprintf("conf/%s-defaults.ini", ROLE)
	if pathExists(filepath.Join(HomePath, confFile)) {
		return fmt.Errorf("conf file[%s] not found", confFile)
	}

	return nil
}

func (cfg *Cfg) LoadConfiguration(args *CommandLineArgs) (*ini.File, error) {
	defaultConfigFile := path.Join(HomePath, fmt.Sprintf("conf/%s-defaults.ini", ROLE))
	configFiles = append(configFiles, defaultConfigFile)

	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		os.Exit(1)
	}
	parsedFile, err := ini.Load(defaultConfigFile)
	if err != nil {
		log.Fatalf("Failed to parse defaults.ini, %v\n", err)
		return nil, err
	}

	parsedFile.BlockMode = false
	evalConfigValues(parsedFile)
	cfg.Raw = parsedFile

	return parsedFile, err
}

func evalEnvVarExpression(value string) string {
	regex := regexp.MustCompile(`\${(\w+)}`)
	return regex.ReplaceAllStringFunc(value, func(envVar string) string {
		envVar = strings.TrimPrefix(envVar, "${")
		envVar = strings.TrimSuffix(envVar, "}")
		envValue := os.Getenv(envVar)

		if envVar == "HOSTNAME" && envValue == "" {
			envValue, _ = os.Hostname()
		}

		return envValue
	})
}

func evalConfigValues(file *ini.File) {
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			key.SetValue(evalEnvVarExpression(key.Value()))
		}
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
