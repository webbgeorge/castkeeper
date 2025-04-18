package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/config"
)

func TestLoadConfig_ValidSqliteLocal(t *testing.T) {
	cfg, _, err := config.LoadConfig("testdata/valid-sqlite-local.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "testdata",
		LogLevel: "debug",
		BaseURL:  "http://www.example.com",
		WebServer: config.WebServerConfig{
			Port:             80,
			CSRFSecretKey:    "testValueDoNotUseInProd",
			CSRFSecureCookie: true,
		},
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    "./data/db.sql",
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver:        "local",
			LocalBasePath: "./data/objects",
		},
	}, cfg)
}

func TestLoadConfig_ValidPostgresS3(t *testing.T) {
	cfg, _, err := config.LoadConfig("testdata/valid-postgres-s3.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "testdata",
		LogLevel: "error",
		BaseURL:  "http://www.example.com",
		WebServer: config.WebServerConfig{
			Port:             80,
			CSRFSecretKey:    "testValueDoNotUseInProd",
			CSRFSecureCookie: true,
		},
		Database: config.DatabaseConfig{
			Driver: "postgres",
			DSN:    "testValue",
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver:   "awss3",
			S3Bucket: "my-bucket",
			S3Prefix: "some-prefix",
		},
	}, cfg)
}

func TestLoadConfig_EnvVarsOnly(t *testing.T) {
	os.Setenv("CASTKEEPER_ENVNAME", "testdata")
	os.Setenv("CASTKEEPER_LOGLEVEL", "error")
	os.Setenv("CASTKEEPER_BASEURL", "http://www.example.com")
	os.Setenv("CASTKEEPER_WEBSERVER_PORT", "80")
	os.Setenv("CASTKEEPER_WEBSERVER_CSRFSECRETKEY", "testValueDoNotUseInProd")
	os.Setenv("CASTKEEPER_WEBSERVER_CSRFSECURECOOKIE", "true")
	os.Setenv("CASTKEEPER_DATABASE_DRIVER", "postgres")
	os.Setenv("CASTKEEPER_DATABASE_DSN", "testValue")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_DRIVER", "awss3")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_S3BUCKET", "my-bucket")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_S3PREFIX", "some-prefix")
	defer func() {
		os.Unsetenv("CASTKEEPER_ENVNAME")
		os.Unsetenv("CASTKEEPER_LOGLEVEL")
		os.Unsetenv("CASTKEEPER_BASEURL")
		os.Unsetenv("CASTKEEPER_WEBSERVER_PORT")
		os.Unsetenv("CASTKEEPER_WEBSERVER_CSRFSECRETKEY")
		os.Unsetenv("CASTKEEPER_WEBSERVER_CSRFSECURECOOKIE")
		os.Unsetenv("CASTKEEPER_DATABASE_DRIVER")
		os.Unsetenv("CASTKEEPER_DATABASE_DSN")
		os.Unsetenv("CASTKEEPER_OBJECTSTORAGE_DRIVER")
		os.Unsetenv("CASTKEEPER_OBJECTSTORAGE_S3BUCKET")
		os.Unsetenv("CASTKEEPER_OBJECTSTORAGE_S3PREFIX")
	}()

	cfg, _, err := config.LoadConfig("")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "testdata",
		LogLevel: "error",
		BaseURL:  "http://www.example.com",
		WebServer: config.WebServerConfig{
			Port:             80,
			CSRFSecretKey:    "testValueDoNotUseInProd",
			CSRFSecureCookie: true,
		},
		Database: config.DatabaseConfig{
			Driver: "postgres",
			DSN:    "testValue",
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver:   "awss3",
			S3Bucket: "my-bucket",
			S3Prefix: "some-prefix",
		},
	}, cfg)
}

func TestLoadConfig_EnvVarOverride(t *testing.T) {
	os.Setenv("CASTKEEPER_ENVNAME", "OverridenByEnv")
	defer func() {
		os.Unsetenv("CASTKEEPER_ENVNAME")
	}()

	cfg, _, err := config.LoadConfig("testdata/valid-sqlite-local.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "OverridenByEnv",
		LogLevel: "debug",
		BaseURL:  "http://www.example.com",
		WebServer: config.WebServerConfig{
			Port:             80,
			CSRFSecretKey:    "testValueDoNotUseInProd",
			CSRFSecureCookie: true,
		},
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    "./data/db.sql",
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver:        "local",
			LocalBasePath: "./data/objects",
		},
	}, cfg)
}

func TestLoadConfig_ValidationErr(t *testing.T) {
	testCases := map[string]struct {
		configFile  string
		expectedErr string
	}{
		"invalidLogLevel": {
			configFile:  "testdata/invalid-log-level.yml",
			expectedErr: "Key: 'Config.LogLevel' Error:Field validation for 'LogLevel' failed on the 'oneof' tag",
		},
		"invalidPort": {
			configFile:  "testdata/invalid-port.yml",
			expectedErr: "Key: 'Config.WebServer.Port' Error:Field validation for 'Port' failed on the 'lte' tag",
		},
		"invalidDatabaseDriver": {
			configFile:  "testdata/invalid-db-driver.yml",
			expectedErr: "Key: 'Config.Database.Driver' Error:Field validation for 'Driver' failed on the 'oneof' tag",
		},
		"missingDSN": {
			configFile:  "testdata/invalid-missing-dsn.yml",
			expectedErr: "Key: 'Config.Database.DSN' Error:Field validation for 'DSN' failed on the 'required' tag",
		},
		"invalidStorageDriver": {
			configFile:  "testdata/invalid-storage-driver.yml",
			expectedErr: "Key: 'Config.ObjectStorage.Driver' Error:Field validation for 'Driver' failed on the 'oneof' tag",
		},
		"missingLocalBasePath": {
			configFile:  "testdata/invalid-missing-local-base-path.yml",
			expectedErr: "Key: 'Config.ObjectStorage.LocalBasePath' Error:Field validation for 'LocalBasePath' failed on the 'required_if' tag",
		},
		"missingS3Bucket": {
			configFile:  "testdata/invalid-s3-bucket.yml",
			expectedErr: "Key: 'Config.ObjectStorage.S3Bucket' Error:Field validation for 'S3Bucket' failed on the 'required_if' tag",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, _, err := config.LoadConfig(tc.configFile)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedErr, err.Error())
		})
	}
}
