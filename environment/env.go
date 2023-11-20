package environment

import (
	"context"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/naoina/toml"
)

type Config struct {
	Param     map[string]string `toml:"param" json:"param,omitempty"`
	Env       map[string]string `toml:"env" json:"env,omitempty"`
	GcpSecret map[string]string `toml:"gcp_secret" json:"gcp_secret,omitempty"`
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

	// 環境変数の読み込み
	env := readEnv(cfg)
	// シークレットの読み込み
	sec, err := readSecret(ctx, cfg)
	if err != nil {
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
		result[k] = replaceParam(v, cfg.Param)
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
	for k, v := range cfg.GcpSecret {
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: replaceParam(v, cfg.Param),
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

func replaceParam(s string, param map[string]string) string {
	for k, v := range param {
		s = strings.Replace(s, fmt.Sprintf("{%s}", k), v, -1)
	}
	return s
}
