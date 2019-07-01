package libdisco

import (
	"os"
	"testing"
)

func TestCreationKeys(t *testing.T) {

	// temporary files
	discoKeyPairFile := "./discoKeyPairFile"
	defer func() {
		if err := os.Remove(discoKeyPairFile); err != nil {
			panic(err)
		}
	}()
	rootPrivateKeyFile := "./rootPrivateKeyFile"
	defer os.Remove(rootPrivateKeyFile)
	rootPublicKeyFile := "./rootPublicKeyFile"
	defer os.Remove(rootPublicKeyFile)

	// Generate Disco Key pair
	keyPair, err := GenerateAndSaveDiscoKeyPair(discoKeyPairFile)
	if err != nil {
		t.Error("Disco key pair couldn't be written on disk")
		return
	}
	// Load Disco Key pair
	keyPairTemp, err := LoadDiscoKeyPair(discoKeyPairFile)
	if err != nil {
		t.Error("Disco key pair couldn't be loaded from disk")
		return
	}
	// compare
	for i := 0; i < 32; i++ {
		if keyPair.PublicKey[i] != keyPairTemp.PublicKey[i] ||
			keyPair.PrivateKey[i] != keyPairTemp.PrivateKey[i] {
			t.Error("Disco key pair generated and loaded are different")
			return
		}
	}
	// generate root key
	err = GenerateAndSaveDiscoRootKeyPair(rootPrivateKeyFile, rootPublicKeyFile)
	if err != nil {
		t.Error("Disco key pair couldn't be written on disk")
		return
	}

	// load private root key
	rootPriv, err := LoadDiscoRootPrivateKey(rootPrivateKeyFile)
	if err != nil {
		t.Error("Disco root private key couldn't be loaded from disk")
		return
	}
	// load public root key
	rootPub, err := LoadDiscoRootPublicKey(rootPublicKeyFile)
	if err != nil {
		t.Error("Disco root public key couldn't be loaded from disk")
		return
	}

	// create a proof
	proof := CreateStaticPublicKeyProof(rootPriv, keyPair.PublicKey[:])

	// verify the proof
	verifior := CreatePublicKeyVerifier(rootPub)
	if !verifior(keyPair.PublicKey[:], proof) {
		t.Error("cannot verify proof")
		return
	}

	// end
}
