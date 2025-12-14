package encryption

import (
	"github.com/tink-crypto/tink-go/v2/aead/subtle"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/prf"
	"github.com/tink-crypto/tink-go/v2/tink"
)

func DeriveAEADFromSecret(secret string) (tink.AEAD, error) {
	prfHandle, err := keyset.NewHandle(
		prf.HKDFSHA256PRFKeyTemplate(),
	)
	if err != nil {
		return nil, err
	}

	prfSet, err := prf.NewPRFSet(prfHandle)
	if err != nil {
		return nil, err
	}

	keyMaterial, err := prfSet.ComputePrimaryPRF(
		[]byte(secret),
		32,
	)
	if err != nil {
		return nil, err
	}

	return subtle.NewAESGCMSIV(keyMaterial)
}
