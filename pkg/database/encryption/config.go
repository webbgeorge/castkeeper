package encryption

import (
	"github.com/webbgeorge/castkeeper/pkg/config"
)

func ConfigureEncryptedValueService(
	cfg config.Config,
) (*EncryptedValueService, error) {
	if cfg.Encryption.Driver != config.EncryptionDriverSecretKey {
		return nil, nil
	}

	dekService, err := NewDEKService(cfg)
	if err != nil {
		return nil, err
	}

	dekAEAD, err := dekService.aead()
	if err != nil {
		return nil, err
	}

	return &EncryptedValueService{
		dekAEAD: dekAEAD,
	}, nil
}
