package encryption_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
)

func TestEncryptDecrypt_Success(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	evs := encryption.NewEncryptedValueService(
		db,
		[]byte("test_key_1"),
		1,
	)

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evs.EncryptAndSave(
		"test_table", "test_id", "test_col_1", testData, testAD,
	)

	assert.Nil(t, err)
	assert.Equal(t, "test_table", ev.ParentTable)
	assert.Equal(t, "test_id", ev.ParentID)
	assert.Equal(t, "test_col_1", ev.ParentColumn)
	assert.Len(t, ev.Salt, 32)
	assert.NotEqual(t, testData, ev.EncryptedData)

	decryptedPT, err := evs.Decrypt(ev, testAD)

	assert.Nil(t, err)
	assert.Equal(t, string(testData), string(decryptedPT))
}

func TestEncryptDecrypt_WrongMasterKey(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	evsKey1 := encryption.NewEncryptedValueService(
		db,
		[]byte("test_key_1"),
		1,
	)

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evsKey1.EncryptAndSave(
		"test_table", "test_id", "test_col_1", testData, testAD,
	)
	assert.Nil(t, err)

	evsKey2 := encryption.NewEncryptedValueService(
		db,
		[]byte("test_key_2"),
		1,
	)

	decryptedPT, err := evsKey2.Decrypt(ev, testAD)

	assert.Equal(t, "cipher: message authentication failed", err.Error())
	assert.Nil(t, decryptedPT)
}

func TestEncryptDecrypt_ModifiedAdditionalData(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	evs := encryption.NewEncryptedValueService(
		db,
		[]byte("test_key_1"),
		1,
	)

	testData := []byte("test_data")
	testAD := []byte("test_ad")

	ev, err := evs.EncryptAndSave(
		"test_table", "test_id", "test_col_1", testData, testAD,
	)
	assert.Nil(t, err)

	decryptedPT, err := evs.Decrypt(ev, []byte("not_original_ad"))

	assert.Equal(t, "cipher: message authentication failed", err.Error())
	assert.Nil(t, decryptedPT)
}
