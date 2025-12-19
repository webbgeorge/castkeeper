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

//	type dekService struct {
//		kekAEAD tink.AEAD
//		dataDir *os.Root
//	}
//
//	func (s *dekService) loadOrCreate() (*keyset.Handle, error) {
//		_, err := s.dataDir.Stat(dekFileName)
//		if err != nil {
//			if errors.Is(err, os.ErrNotExist) {
//				return s.create()
//			}
//			return nil, err
//		}
//		return s.load()
//	}
//
//	func (s *dekService) load() (*keyset.Handle, error) {
//		f, err := s.dataDir.Open(dekFileName)
//		if err != nil {
//			return nil, err
//		}
//		defer f.Close()
//
//		reader := keyset.NewJSONReader(f)
//		return keyset.Read(reader, s.kekAEAD)
//	}
//
//	func (s *dekService) create() (*keyset.Handle, error) {
//		handle, err := keyset.NewHandle(aead.AES256GCMSIVKeyTemplate())
//		if err != nil {
//			return nil, err
//		}
//
//		f, err := s.dataDir.Create(dekFileName)
//		if err != nil {
//			return nil, err
//		}
//		defer f.Close()
//
//		writer := keyset.NewJSONWriter(f)
//		err = handle.Write(writer, s.kekAEAD)
//		if err != nil {
//			return nil, err
//		}
//
//		// we want to be sure the key is written to disk before we start using it
//		err = f.Sync()
//		if err != nil {
//			return nil, err
//		}
//
//		return handle, nil
//	}
//
// func (s *dekService) listKeys(handle *keyset.Handle) {} // TODO required?
//
//	func (s *dekService) addKey(handle *keyset.Handle) (uint32, error) {
//		mgr := keyset.NewManagerFromHandle(handle)
//
//		// add new key and make it primary
//		newKeyID, err := mgr.Add(aead.AES256GCMSIVKeyTemplate())
//		if err != nil {
//			return 0, err
//		}
//
//		// TODO might need to be separate
//		err = mgr.SetPrimary(newKeyID)
//		if err != nil {
//			return 0, err
//		}
//
//		return newKeyID, nil
//	}
//
//	func (s *dekService) disableKey(handle *keyset.Handle, keyID uint32) error {
//		mgr := keyset.NewManagerFromHandle(handle)
//		return mgr.Disable(keyID)
//	}
//
//	func (s *dekService) deleteKey(handle *keyset.Handle, keyID uint32) error {
//		mgr := keyset.NewManagerFromHandle(handle)
//		return mgr.Delete(keyID)
//	}
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
