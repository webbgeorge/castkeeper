package encryption_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
)

func TestEncryptDecrypt_Success(t *testing.T) {
	evs := setupEVSFromSecret("testEncSecret")

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evs.Encrypt(testData, testAD)

	assert.Nil(t, err)
	assert.NotEqual(t, testData, ev.EncryptedData)

	decryptedPT, err := evs.Decrypt(ev, testAD)

	assert.Nil(t, err)
	assert.Equal(t, string(testData), string(decryptedPT))
}

func TestEncryptDecrypt_SuccessWithNewInstance(t *testing.T) {
	evs := setupEVSFromSecret("testEncSecret")

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evs.Encrypt(testData, testAD)

	assert.Nil(t, err)
	assert.NotEqual(t, testData, ev.EncryptedData)

	evs2 := setupEVSFromSecret("testEncSecret")

	decryptedPT, err := evs2.Decrypt(ev, testAD)

	assert.Nil(t, err)
	assert.Equal(t, string(testData), string(decryptedPT))
}

func TestEncryptDecrypt_WrongKey(t *testing.T) {
	evsKey1 := setupEVSFromSecret("testEncSecret")

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evsKey1.Encrypt(testData, testAD)
	assert.Nil(t, err)

	evsKey2 := setupEVSFromSecret("otherTestEncSecret")

	decryptedPT, err := evsKey2.Decrypt(ev, testAD)

	assert.Equal(t, "aes_gcm_siv: message authentication failure", err.Error())
	assert.Nil(t, decryptedPT)
}

func TestEncryptDecrypt_ModifiedAdditionalData(t *testing.T) {
	evs := setupEVSFromSecret("testEncSecret")

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evs.Encrypt(
		testData, testAD,
	)
	assert.Nil(t, err)

	decryptedPT, err := evs.Decrypt(ev, []byte("not_original_ad"))

	assert.Equal(t, "aes_gcm_siv: message authentication failure", err.Error())
	assert.Nil(t, decryptedPT)
}

func setupEVSFromSecret(secret string) *encryption.EncryptedValueService {
	aead, err := encryption.DeriveAEADFromSecret(secret)
	if err != nil {
		panic(err)
	}
	return encryption.NewEncryptedValueService(aead)
}
