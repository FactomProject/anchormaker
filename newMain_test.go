package main

import (
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMain(t *testing.T) {
	priv, pub, add, err := factoid.HumanReadiblePrivateKeyStringToEverythingString("Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH")
	t.Errorf("%v, %v, %v, %v", priv, pub, add, err)

	//Es38CboJYYSovciHHtigHiv9kkKf5uzAAQjBst5QYcgt9a4m6ywv

	ecPriv, err := primitives.HumanReadableECPrivateKeyToPrivateKeyString("Es38CboJYYSovciHHtigHiv9kkKf5uzAAQjBst5QYcgt9a4m6ywv")
	t.Errorf("%v, %v", ecPriv, err)

	ecpub, err := primitives.PrivateKeyStringToPublicKeyString(ecPriv)
	t.Errorf("%v, %v", ecpub, err)

	ecpub, err = primitives.PrivateKeyStringToPublicKeyString("397c49e182caa97737c6b394591c614156fbe7998d7bf5d76273961e9fa1edd406ed9e69bfdf85db8aa69820f348d096985bc0b11cc9fc9dcee3b8c68b41dfd5")
	t.Errorf("%v, %v", ecpub, err)

	main()
}
