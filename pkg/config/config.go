package config

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/webbgeorge/castkeeper"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

const (
	applicationName          = "castkeeper"
	DatabaseDriverPostgres   = "postgres"
	DatabaseDriverSqlite     = "sqlite"
	ObjectStorageDriverLocal = "local"
	ObjectStorageDriverS3    = "awss3"
	LogLevelDebug            = "debug"
	LogLevelInfo             = "info"
	LogLevelWarn             = "warn"
	LogLevelError            = "error"
)

type Config struct {
	LogLevel      string              `validate:"required,oneof=debug info warn error"`
	EnvName       string              `validate:"required"`
	WebServer     WebServerConfig     `validate:"required"`
	Database      DatabaseConfig      `validate:"required"`
	ObjectStorage ObjectStorageConfig `validate:"required"`
}

type WebServerConfig struct {
	Port             int    `validate:"required,gt=0,lte=65535"`
	CSRFSecretKey    string `validate:"required,lte=250" secret:"true"`
	CSRFSecureCookie bool
}

type DatabaseConfig struct {
	Driver string `validate:"required,oneof=postgres sqlite"`
	DSN    string `validate:"required" secret:"true"`
}

type ObjectStorageConfig struct {
	Driver           string `validate:"required,oneof=local awss3"`
	LocalBasePath    string `validate:"required_if=Driver local"`
	S3Bucket         string `validate:"required_if=Driver awss3"`
	S3Prefix         string `validate:"lte=250"`
	S3ForcePathStyle bool   // used for localstack testing, unlikely to be ever used in prod
}

func LoadConfig(configFilePath string) (Config, *slog.Logger, error) {
	return loadConfig(viper.GetViper(), configFilePath)
}

func loadConfig(v *viper.Viper, configFilePath string) (Config, *slog.Logger, error) {
	if configFilePath != "" {
		v.SetConfigFile(configFilePath) // if provided and not empty, uses specific file instead of from paths below
	} else {
		v.SetConfigName("castkeeper")       // file called castkeeper.yml|yaml|json
		v.AddConfigPath("/etc/castkeeper/") // in this dir, or...
		v.AddConfigPath(".")                // in current working directory
	}

	v.SetDefault("LogLevel", LogLevelInfo)
	v.SetDefault("EnvName", "unknown")
	v.SetDefault("WebServer.Port", 8080)
	v.SetDefault("WebServer.CSRFSecureCookie", true)

	// allow config to optionally be set using environment variables
	// e.g. CASTKEEPER_WEBSERVER_PORT
	v.SetEnvPrefix("castkeeper")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		return Config{}, nil, err
	}

	config := Config{}
	err = v.Unmarshal(&config)
	if err != nil {
		return Config{}, nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		return Config{}, nil, err
	}

	logger := framework.NewLogger(
		applicationName,
		config.EnvName,
		castkeeper.Version,
		slogLogLevel(config.LogLevel),
	)

	logger.Info(fmt.Sprintf("successfully loaded config from: %s", v.ConfigFileUsed()))
	logger.Debug(fmt.Sprintf("loaded config: %s", debugConfig(config)))

	return config, logger, nil
}

func slogLogLevel(ll string) slog.Level {
	switch ll {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func debugConfig(cfg Config) string {
	debugVals := make([]string, 0)
	debugStruct(cfg, "", &debugVals)
	debugStruct(cfg.WebServer, "WebServer.", &debugVals)
	debugStruct(cfg.Database, "Database.", &debugVals)
	debugStruct(cfg.ObjectStorage, "ObjectStorage.", &debugVals)
	return strings.Join(debugVals, ", ")
}

// grabs config key/value pairs from a struct for debugging
// omits secret config values which have the `secret` struct tag
func debugStruct(s any, prefix string, debugVals *[]string) {
	ignoreKinds := []reflect.Kind{reflect.Slice, reflect.Struct}
	sv := reflect.ValueOf(s)
	tv := reflect.TypeOf(s)
	for i := 0; i < sv.NumField(); i++ {
		if slices.Contains(ignoreKinds, sv.Field(i).Kind()) {
			continue
		}
		val := sv.Field(i).Interface()
		if tv.Field(i).Tag.Get("secret") != "" {
			val = "(secret omitted)"
		}
		*debugVals = append(*debugVals, fmt.Sprintf("%s%s: %v", prefix, tv.Field(i).Name, val))
	}
}
