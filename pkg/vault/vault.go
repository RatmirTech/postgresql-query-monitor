package vault

import (
	"context"
	"fmt"

	"github.com/dreadew/go-common/pkg/config/pg"
	"github.com/hashicorp/vault/api"
)

// safeString пытается достать строку из map
func safeString(data map[string]interface{}, key string) (string, error) {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("key %q found but not a string", key)
	}
	return "", fmt.Errorf("key %q not found", key)
}

// GetConnectionConfig получает конфиг к базе данных из Vault
func GetConnectionConfig(ctx context.Context, client *api.Client, path string) (*pg.DbConfig, error) {
	secret, err := client.KVv2("secret").Get(ctx, path)
	if err != nil {
		return nil, err
	}

	data := secret.Data
	host, err := safeString(data, "Host")
	if err != nil {
		return nil, err
	}
	port, err := safeString(data, "Port")
	if err != nil {
		return nil, err
	}
	user, err := safeString(data, "Username")
	if err != nil {
		return nil, err
	}
	pass, err := safeString(data, "Password")
	if err != nil {
		return nil, err
	}
	db, err := safeString(data, "Database")
	if err != nil {
		return nil, err
	}
	ssl, err := safeString(data, "SSLMode")
	if err != nil {
		return nil, err
	}

	return &pg.DbConfig{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Database: db,
		SSLMode:  ssl,
	}, nil
}
