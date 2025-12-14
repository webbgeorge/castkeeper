package encryption

import (
	"errors"

	"github.com/tink-crypto/tink-go/v2/tink"
)

type EncryptedValue struct {
	EncryptedData []byte
	// TODO store in new table with references, additionalData, etc
}

type EncryptedValueService struct {
	dekAEAD tink.AEAD
}

func NewEncryptedValueService(dekAEAD tink.AEAD) *EncryptedValueService {
	return &EncryptedValueService{
		dekAEAD: dekAEAD,
	}
}

var ErrEncryptionNotConfigured = errors.New("encryption is not configured")

func (s *EncryptedValueService) Encrypt(
	plaintext []byte,
	additionalData []byte,
) (EncryptedValue, error) {
	if s == nil {
		return EncryptedValue{}, ErrEncryptionNotConfigured
	}

	ciphertext, err := s.dekAEAD.Encrypt(plaintext, additionalData)
	if err != nil {
		return EncryptedValue{}, err
	}

	return EncryptedValue{
		EncryptedData: ciphertext,
	}, nil
}

func (s *EncryptedValueService) Decrypt(
	ev EncryptedValue,
	additionalData []byte,
) ([]byte, error) {
	if s == nil {
		return nil, errors.New("encryption is not configured")
	}

	return s.dekAEAD.Decrypt(ev.EncryptedData, additionalData)
}
