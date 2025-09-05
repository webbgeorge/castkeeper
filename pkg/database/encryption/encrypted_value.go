package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
	"gorm.io/gorm"
)

type EncryptedValue struct {
	gorm.Model
	ParentTable   string
	ParentID      string
	ParentColumn  string
	EncryptedData []byte
	KeyVersion    uint
	Salt          []byte
}

type EncryptedValueService struct {
	db         *gorm.DB
	masterKey  []byte
	keyVersion uint
}

func NewEncryptedValueService(
	db *gorm.DB,
	masterKey []byte,
	keyVersion uint,
) *EncryptedValueService {
	return &EncryptedValueService{
		db:         db,
		masterKey:  masterKey,
		keyVersion: keyVersion,
	}
}

func (s *EncryptedValueService) EncryptAndSave(
	parentTable, parentID, parentColumn string,
	plaintext []byte,
	additionalData []byte,
) (EncryptedValue, error) {
	salt, err := randBytes(32)
	if err != nil {
		return EncryptedValue{}, err
	}

	aead, err := s.getAEAD(
		salt, additionalData, parentTable, parentID, parentColumn,
	)
	if err != nil {
		return EncryptedValue{}, err
	}

	ciphertext := aead.Seal(nil, nil, plaintext, additionalData)

	ev := EncryptedValue{
		ParentTable:   parentTable,
		ParentID:      parentID,
		ParentColumn:  parentColumn,
		KeyVersion:    s.keyVersion,
		Salt:          salt,
		EncryptedData: ciphertext,
	}
	if err := s.db.Create(&ev).Error; err != nil {
		return EncryptedValue{}, err
	}

	return ev, nil
}

func (s *EncryptedValueService) Decrypt(
	ev EncryptedValue,
	additionalData []byte,
) ([]byte, error) {
	aead, err := s.getAEAD(
		ev.Salt, additionalData, ev.ParentTable, ev.ParentID, ev.ParentColumn,
	)
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
	parentTable, parentID, parentColumn string,
) (cipher.AEAD, error) {
	rowKey, err := s.deriveRowKey(
		salt, additionalData, parentTable, parentID, parentColumn)
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
	parentTable, parentID, parentColumn string,
) ([]byte, error) {
	h := hkdf.New(
		sha256.New,
		s.masterKey,
		salt,
		append([]byte(parentTable+parentID+parentColumn), additionalData...),
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
