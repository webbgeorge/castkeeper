package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/webbgeorge/castkeeper/pkg/config"
	"golang.org/x/crypto/hkdf"
)

type EncryptedValue struct {
	EncryptedData []byte
	KeyVersion    uint
	Salt          []byte
}

type EncryptedValueService struct {
	masterKey  []byte
	keyVersion uint
}

func NewEncryptedValueService(
	cfg config.EncryptionConfig,
) *EncryptedValueService {
	if cfg.Driver != config.EncryptionDriverLocal {
		return nil
	}
	return &EncryptedValueService{
		masterKey:  []byte(cfg.LocalKeyEncryptionKey),
		keyVersion: 1, // TODO replace when key rotation is implemented
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

	salt, err := randBytes(32)
	if err != nil {
		return EncryptedValue{}, err
	}

	aead, err := s.getAEAD(salt, additionalData)
	if err != nil {
		return EncryptedValue{}, err
	}

	ciphertext := aead.Seal(nil, nil, plaintext, additionalData)

	return EncryptedValue{
		KeyVersion:    s.keyVersion,
		Salt:          salt,
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

	aead, err := s.getAEAD(ev.Salt, additionalData)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, nil, ev.EncryptedData, additionalData)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (s *EncryptedValueService) getAEAD(
	salt, additionalData []byte,
) (cipher.AEAD, error) {
	rowKey, err := s.deriveRowKey(salt, additionalData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(rowKey)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCMWithRandomNonce(block)
}

func (s *EncryptedValueService) deriveRowKey(
	salt, additionalData []byte,
) ([]byte, error) {
	h := hkdf.New(
		sha256.New,
		s.masterKey,
		salt,
		additionalData,
	)
	key := make([]byte, 32)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, err
	}
	return key, nil
}

func randBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
