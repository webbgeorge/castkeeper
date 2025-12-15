package encryption_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
)

// TODO secret same, but file changed or deleted

func TestConfigureEncryptedValueService_SecretKeyDriver_CreatesNewDEK(t *testing.T) {
	randomHex := fixtures.RandomHex()

	// DEK doesn't exist yet
	rootPath := path.Join(os.TempDir(), "castkeepertest", randomHex)
	_, err := os.Stat(path.Join(rootPath, "dek.json"))
	assert.Contains(t, err.Error(), "no such file or directory")

	evs, _ := configureEVSForTest(randomHex, "secretKeyForTest111")

	// DEK was created
	_, err = os.Stat(path.Join(rootPath, "dek.json"))
	assert.Nil(t, err)

	// can encrypt and decrypt with new DEK
	ev, err := evs.Encrypt([]byte("test"), []byte("testad"))
	assert.Nil(t, err)
	pt, err := evs.Decrypt(ev, []byte("testad"))
	assert.Nil(t, err)
	assert.Equal(t, "test", string(pt))
}

func TestConfigureEncryptedValueService_SecretKeyDriver_LoadsDEK(t *testing.T) {
	randomHex := fixtures.RandomHex()

	// this evs creates the DEK in the test dir for the first time
	evs, _ := configureEVSForTest(randomHex, "secretKeyForTest111")

	// encrypts with new dek after it was first created
	ev, err := evs.Encrypt([]byte("test"), []byte("testad"))
	if err != nil {
		panic(err)
	}

	// this evs loads the DEK created by the previous evs
	evs2, _ := configureEVSForTest(randomHex, "secretKeyForTest111")

	// check we can decrypt with the loaded DEK
	pt, err := evs2.Decrypt(ev, []byte("testad"))
	assert.Nil(t, err)
	assert.Equal(t, "test", string(pt))
}

func TestConfigureEncryptedValueService_SecretKeyDriver_ChangingSecretPreventsLoadingDEK(t *testing.T) {
	randomHex := fixtures.RandomHex()

	// this evs creates the DEK in the test dir for the first time
	_, err := configureEVSForTest(randomHex, "secretKeyForTest111")
	assert.Nil(t, err)

	// this evs uses an incorrect secret for the existing DEK
	_, err = configureEVSForTest(randomHex, "differentSecret111")
	assert.Equal(t, "keyset.Handle: decryption failed: aes_gcm_siv: message authentication failure", err.Error())
}

func TestConfigureEncryptedValueService_NoDriver(t *testing.T) {
	rootPath := path.Join(os.TempDir(), "castkeepertest", fixtures.RandomHex())
	evs, _ := encryption.ConfigureEncryptedValueService(config.Config{
		DataPath: rootPath,
		Encryption: config.EncryptionConfig{
			Driver: "",
		},
	})

	_, err := evs.Encrypt([]byte("test"), []byte("testad"))
	assert.Equal(t, "encryption is not configured", err.Error())
}

func configureEVSForTest(randomHex, secret string) (*encryption.EncryptedValueService, error) {
	rootPath := path.Join(os.TempDir(), "castkeepertest", randomHex)
	return encryption.ConfigureEncryptedValueService(config.Config{
		DataPath: rootPath,
		Encryption: config.EncryptionConfig{
			Driver:    "secretkey",
			SecretKey: secret,
		},
	})
}
