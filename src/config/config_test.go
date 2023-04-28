package config

import (
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("should load config from file and env variables", func(t *testing.T) {
		t.Setenv("GITLAB_API_TOKEN", "def456")

		tempDir, err := os.MkdirTemp("", "config-test-*")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		tempFile, err := os.CreateTemp(tempDir, "config.yaml")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		configData := []byte(`gitlabBaseApiUrl: "https://gitlab.com/api/v4"`)
		_, err = tempFile.Write(configData)
		if err != nil {
			t.Fatalf("Failed to write config to file: %v", err)
		}

		fn := tempFile.Name()
		t.Logf("Temp file: %s", fn)

		log := logrus.New()
		config := LoadConfig(tempFile.Name(), log)

		expected := &AutoCloserConfig{
			GitlabApiToken:   "def456",
			GitlabBaseApiUrl: "https://gitlab.com/api/v4",
		}

		if !reflect.DeepEqual(config, expected) {
			t.Errorf("Expected config to be %v, but got %v", expected, config)
		}
	})
}
