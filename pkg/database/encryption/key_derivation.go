package encryption

import (
	"github.com/tink-crypto/tink-go/v2/aead/subtle"
	"github.com/tink-crypto/tink-go/v2/tink"
	"golang.org/x/crypto/argon2"
)

func DeriveAEADFromSecret(secret string) (tink.AEAD, error) {
	// salt is fixed because this KDF has to be deterministic, based on
	// the user's secret
	fixedSalt := []byte("castkeeper:kek")

	key := argon2.IDKey(
		[]byte(secret),
		fixedSalt,
		1,
		64*1024,
		1,
		32,
	)

	return subtle.NewAESGCMSIV(key)
}
