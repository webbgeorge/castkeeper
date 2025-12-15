package fixtures

import "github.com/webbgeorge/castkeeper/pkg/database/encryption"

func ConfigureEncryptedValueServiceForTest() *encryption.EncryptedValueService {
	aead, err := encryption.DeriveAEADFromSecret("00000000")
	if err != nil {
		panic(err)
	}
	return encryption.NewEncryptedValueService(aead)
}
