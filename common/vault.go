package common

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"go-api/config"
	"time"
)

func read(path string) (map[string]interface{}, error) {
	ctx := context.Background()
	//get vault client
	client, err := getVaultClient()
	if err != nil {
		return nil, err
	}

	// read a secret
	resp, err := client.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	return resp.Data["data"].(map[string]interface{}), nil
}

func getVaultClient() (*vault.Client, error) {
	ctx := context.Background()
	// prepare a client with the given base address
	client, err := vault.New(
		vault.WithAddress(config.Env.VaultUrl),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// authenticate using approle
	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   config.Env.VaultRoleId,
			SecretId: config.Env.VaultSecretId,
		},
		vault.WithMountPath(""),
	)
	if err != nil {
		return nil, err
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		return nil, err
	}

	return client, nil
}
