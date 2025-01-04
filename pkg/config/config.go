package config

import (
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	DatabaseDriverPostgres   = "postgres"
	DatabaseDriverSqlite     = "sqlite"
	ObjectStorageDriverLocal = "local"
	ObjectStorageDriverS3    = "awss3"
)

type Config struct {
	EnvName       string              `validate:"required"`
	WebServer     WebServerConfig     `validate:"required"`
	Database      DatabaseConfig      `validate:"required"`
	ObjectStorage ObjectStorageConfig `validate:"required"`
	OpenTelemetry OpenTelemetryConfig `validate:""`
}

type WebServerConfig struct {
	Port int `validate:"required,gt=0,lte=65535"`
}

type DatabaseConfig struct {
	Driver string `validate:"required,oneof=postgres sqlite"`
}

type ObjectStorageConfig struct {
	Driver string `validate:"required,oneof=local awss3"`
}

type OpenTelemetryConfig struct {
	Enabled bool
}

func LoadConfig() (Config, error) {
	return loadConfig(viper.GetViper())
}

func loadConfig(v *viper.Viper) (Config, error) {
	v.SetConfigName("castkeeper")       // file called castkeeper.yml|yaml|json
	v.AddConfigPath("/etc/castkeeper/") // in this dir, or...
	v.AddConfigPath(".")                // in current working directory

	v.SetDefault("EnvName", "unknown")
	v.SetDefault("WebServer.Port", 8080)

	// allow config to optionally be set using environment variables
	// e.g. CASTKEEPER_WEBSERVER_PORT
	v.SetEnvPrefix("castkeeper")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = v.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		return Config{}, err
	}

	log.Printf("successfully loaded config from %s", v.ConfigFileUsed())

	return config, nil
}
