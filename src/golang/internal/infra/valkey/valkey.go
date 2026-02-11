package valkey

import (
	"fmt"

	"github.com/valkey-io/valkey-go"
)

type ValkeyConfig struct {
	Addr     string
	Username string
	Password string
}

func NewValkeyClient(cfg ValkeyConfig) (valkey.Client, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.Addr},
		Username:    cfg.Username,
		Password:    cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init Valkey connection: %w", err)
	}
	return client, nil
}
