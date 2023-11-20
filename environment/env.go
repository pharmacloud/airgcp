package environment

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/naoina/toml"
)

type Config struct {
	ProjectId string            `toml:"project_id" json:"project_id,omitempty"`
	Env       map[string]string `toml:"env" json:"env,omitempty"`
	Secret    map[string]string `toml:"secret" json:"secret,omitempty"`
}

var defaultEnvfile = ".air-env.toml"

func InitEnv(ctx context.Context, path string) error {
	if path == "" {
		path = defaultEnvfile
	}
	cfg, err := readConfig(path)
	if err != nil {
		return nil
	}
	log.Printf("???: %+v", cfg)
	sec, err := readSecret(ctx, cfg)
	if err != nil {
		return err
	}

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", cfg.ProjectId); err != nil {
		return err
	}

	if err := setEnv(cfg.Env); err != nil {
		return err
	}
	if err := setEnv(sec); err != nil {
		return err
	}
	return nil
}

func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err = toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func readSecret(ctx context.Context, cfg *Config) (map[string]string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	result := map[string]string{}
	for k, v := range cfg.Secret {
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", cfg.ProjectId, v),
		}
		resp, err := client.AccessSecretVersion(ctx, req)
		if err != nil {
			return nil, err
		}
		result[k] = string(resp.Payload.Data)
	}
	return result, nil
}

func setEnv(env map[string]string) error {
	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			return nil
		}
	}
	return nil
}
