package encryption

import (
	"errors"
	"os"
	"path"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
	"github.com/webbgeorge/castkeeper/pkg/config"
)

const dekFileName = "dek.json"

type dekService struct {
	kekAEAD tink.AEAD
	dataDir *os.Root
}

func NewDEKService(cfg config.Config) (*dekService, error) {
	kekAEAD, err := DeriveAEADFromSecret(
		cfg.Encryption.SecretKey,
	)
	if err != nil {
		return nil, err
	}

	return &dekService{
		kekAEAD: kekAEAD,
		dataDir: config.MustOpenLocalFSRoot(
			path.Join(cfg.DataPath),
		),
	}, nil
}

func (s *dekService) loadOrCreate() (*keyset.Handle, error) {
	_, err := s.dataDir.Stat(dekFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.create()
		}
		return nil, err
	}
	return s.load()
}

func (s *dekService) load() (*keyset.Handle, error) {
	f, err := s.dataDir.Open(dekFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := keyset.NewJSONReader(f)
	return keyset.Read(reader, s.kekAEAD)
}

func (s *dekService) create() (*keyset.Handle, error) {
	handle, err := keyset.NewHandle(aead.AES256GCMSIVKeyTemplate())
	if err != nil {
		return nil, err
	}

	err = s.write(handle)
	if err != nil {
		return nil, err
	}

	return handle, nil
}

func (s *dekService) write(handle *keyset.Handle) error {
	f, err := s.dataDir.Create(dekFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := keyset.NewJSONWriter(f)
	err = handle.Write(writer, s.kekAEAD)
	if err != nil {
		return err
	}

	// we want to be sure the key is written to disk
	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

const (
	KeyStatusDisabled = "Disabled"
	KeyStatusEnabled  = "Enabled"
)

type KeyInfo struct {
	ID        uint32
	Status    string
	IsPrimary bool
}

func (s *dekService) GetKey(keyID uint32) (KeyInfo, error) {
	keys, err := s.ListKeys()
	if err != nil {
		return KeyInfo{}, err
	}

	for _, key := range keys {
		if key.ID == keyID {
			return key, nil
		}
	}

	return KeyInfo{}, errors.New("key not found")
}

func (s *dekService) ListKeys() ([]KeyInfo, error) {
	handle, err := s.loadOrCreate()
	if err != nil {
		return nil, err
	}

	keys := make([]KeyInfo, 0)
	for i := range handle.Len() {
		entry, err := handle.Entry(i)
		if err != nil {
			return nil, err
		}

		keys = append(keys, KeyInfo{
			ID:        entry.KeyID(),
			Status:    entry.KeyStatus().String(),
			IsPrimary: entry.IsPrimary(),
		})
	}

	return keys, nil
}

func (s *dekService) AddKey() (uint32, error) {
	handle, err := s.loadOrCreate()
	if err != nil {
		return 0, err
	}

	mgr := keyset.NewManagerFromHandle(handle)

	// add new key
	newKeyID, err := mgr.Add(aead.AES256GCMSIVKeyTemplate())
	if err != nil {
		return 0, err
	}

	// make it primary
	err = mgr.SetPrimary(newKeyID)
	if err != nil {
		return 0, err
	}

	newHandle, err := mgr.Handle()
	if err != nil {
		return 0, err
	}

	err = s.write(newHandle)
	if err != nil {
		return 0, err
	}

	return newKeyID, nil
}

func (s *dekService) DisableKey(keyID uint32) error {
	handle, err := s.loadOrCreate()
	if err != nil {
		return err
	}

	mgr := keyset.NewManagerFromHandle(handle)

	err = mgr.Disable(keyID)
	if err != nil {
		return err
	}

	newHandle, err := mgr.Handle()
	if err != nil {
		return err
	}

	err = s.write(newHandle)
	if err != nil {
		return err
	}

	return nil
}

func (s *dekService) EnableKey(keyID uint32) error {
	handle, err := s.loadOrCreate()
	if err != nil {
		return err
	}

	mgr := keyset.NewManagerFromHandle(handle)

	err = mgr.Enable(keyID)
	if err != nil {
		return err
	}

	newHandle, err := mgr.Handle()
	if err != nil {
		return err
	}

	err = s.write(newHandle)
	if err != nil {
		return err
	}

	return nil
}

func (s *dekService) DeleteKey(keyID uint32) error {
	handle, err := s.loadOrCreate()
	if err != nil {
		return err
	}

	mgr := keyset.NewManagerFromHandle(handle)

	err = mgr.Delete(keyID)
	if err != nil {
		return err
	}

	newHandle, err := mgr.Handle()
	if err != nil {
		return err
	}

	err = s.write(newHandle)
	if err != nil {
		return err
	}

	return nil
}

func ConfigureEncryptedValueService(
	cfg config.Config,
) (*EncryptedValueService, error) {
	if cfg.Encryption.Driver != config.EncryptionDriverSecretKey {
		return nil, nil
	}

	// TODO KEK dependent on driver
	kekAEAD, err := DeriveAEADFromSecret(
		cfg.Encryption.SecretKey,
	)
	if err != nil {
		return nil, err
	}

	dekAEAD, err := loadOrCreateDEK(
		kekAEAD,
		config.MustOpenLocalFSRoot(
			path.Join(cfg.DataPath),
		),
	)
	if err != nil {
		return nil, err
	}

	return &EncryptedValueService{
		dekAEAD: dekAEAD,
	}, nil
}

func loadOrCreateDEK(kekAEAD tink.AEAD, dataDir *os.Root) (tink.AEAD, error) {
	_, err := dataDir.Stat(dekFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return createDEK(kekAEAD, dataDir)
		}
		return nil, err
	}
	return loadDEK(kekAEAD, dataDir)
}

func loadDEK(kekAEAD tink.AEAD, dataDir *os.Root) (tink.AEAD, error) {
	f, err := dataDir.Open(dekFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := keyset.NewJSONReader(f)
	handle, err := keyset.Read(reader, kekAEAD)
	if err != nil {
		return nil, err
	}

	return aead.New(handle)
}

func createDEK(kekAEAD tink.AEAD, dataDir *os.Root) (tink.AEAD, error) {
	handle, err := keyset.NewHandle(aead.AES256GCMSIVKeyTemplate())
	if err != nil {
		return nil, err
	}

	f, err := dataDir.Create(dekFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	writer := keyset.NewJSONWriter(f)
	err = handle.Write(writer, kekAEAD)
	if err != nil {
		return nil, err
	}

	// we want to be sure the key is written to disk before we start using it
	err = f.Sync()
	if err != nil {
		return nil, err
	}

	return aead.New(handle)
}
