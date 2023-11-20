package environment

import (
	"context"
	"os"
	"strings"

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
	// パスが指定していない場合にはディフォルトパスを指定
	if path == "" {
		path = defaultEnvfile
	}
	// 設定の読み込み
	cfg, err := readConfig(path)
	if err != nil {
		return nil
	}
	// ProjectIdが指定していない場合は何もしない
	if cfg.ProjectId == "" {
		return nil
	}

	// 環境変数の読み込み
	env := readEnv(cfg)
	// シークレットの読み込み
	sec, err := readSecret(ctx, cfg)
	if err != nil {
		return err
	}

	// ProjectIdを既定の環境変数として設定
	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", cfg.ProjectId); err != nil {
		return err
	}
	// 環境変数の設定
	if err := setEnv(env); err != nil {
		return err
	}
	// シークレットを環境変数として指定
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

func readEnv(cfg *Config) map[string]string {
	result := map[string]string{}
	for k, v := range cfg.Env {
		result[k] = replactProjectId(v, cfg.ProjectId)
	}
	return result
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
			Name: replactProjectId(v, cfg.ProjectId),
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

func replactProjectId(s, pid string) string {
	return strings.Replace(s, "{project_id}", pid, -1)
}
