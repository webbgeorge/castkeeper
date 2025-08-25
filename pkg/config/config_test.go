package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/config"
)

func TestLoadConfig_ValidLocal(t *testing.T) {
	cfg, _, err := config.LoadConfig("testdata/valid-local.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "testdata",
		LogLevel: "debug",
		BaseURL:  "http://www.example.com",
		DataPath: "./data",
		WebServer: config.WebServerConfig{
			Port: 80,
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver: "local",
		},
	}, cfg)
}

func TestLoadConfig_ValidS3(t *testing.T) {
	cfg, _, err := config.LoadConfig("testdata/valid-s3.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "testdata",
		LogLevel: "error",
		BaseURL:  "http://www.example.com",
		DataPath: "./data",
		WebServer: config.WebServerConfig{
			Port: 80,
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
	os.Setenv("CASTKEEPER_DATAPATH", "./data")
	os.Setenv("CASTKEEPER_WEBSERVER_PORT", "80")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_DRIVER", "awss3")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_S3BUCKET", "my-bucket")
	os.Setenv("CASTKEEPER_OBJECTSTORAGE_S3PREFIX", "some-prefix")
	defer func() {
		os.Unsetenv("CASTKEEPER_ENVNAME")
		os.Unsetenv("CASTKEEPER_LOGLEVEL")
		os.Unsetenv("CASTKEEPER_BASEURL")
		os.Unsetenv("CASTKEEPER_DATAPATH")
		os.Unsetenv("CASTKEEPER_WEBSERVER_PORT")
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
		DataPath: "./data",
		WebServer: config.WebServerConfig{
			Port: 80,
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

	cfg, _, err := config.LoadConfig("testdata/valid-local.yml")
	assert.Nil(t, err)
	assert.Equal(t, config.Config{
		EnvName:  "OverridenByEnv",
		LogLevel: "debug",
		BaseURL:  "http://www.example.com",
		DataPath: "./data",
		WebServer: config.WebServerConfig{
			Port: 80,
		},
		ObjectStorage: config.ObjectStorageConfig{
			Driver: "local",
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
		"missingDataPath": {
			configFile:  "testdata/invalid-missing-data-path.yml",
			expectedErr: "Key: 'Config.DataPath' Error:Field validation for 'DataPath' failed on the 'required' tag",
		},
		"invalidStorageDriver": {
			configFile:  "testdata/invalid-storage-driver.yml",
			expectedErr: "Key: 'Config.ObjectStorage.Driver' Error:Field validation for 'Driver' failed on the 'oneof' tag",
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
