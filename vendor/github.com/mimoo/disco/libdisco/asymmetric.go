package libdisco

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/curve25519"
)

//
// The following code defines the X25519, chacha20poly1305, SHA-256 suite.
//

const (
	dhLen = 32 // A constant specifying the size in bytes of public keys and DH outputs. For security reasons, dhLen must be 32 or greater.
)

// 4.1. DH functions

// TODO: store the KeyPair's parts in *[32]byte or []byte ?

// KeyPair contains a private and a public part, both of 32-byte.
// It can be generated via the GenerateKeyPair() function.
// The public part can also be extracted via the ExportPublicKey() function.
type KeyPair struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
}

// GenerateKeypair creates a X25519 static keyPair out of a private key. If privateKey is nil the function generates a random key pair.
func GenerateKeypair(privateKey *[32]byte) *KeyPair {

	var keyPair KeyPair
	if privateKey != nil {
		copy(keyPair.PrivateKey[:], privateKey[:])
	} else {
		if _, err := rand.Read(keyPair.PrivateKey[:]); err != nil {
			panic(err)
		}
	}

	curve25519.ScalarBaseMult(&keyPair.PublicKey, &keyPair.PrivateKey)

	return &keyPair
}

// ExportPublicKey returns the public part in hex format of a static key pair.
func (kp KeyPair) ExportPublicKey() string {
	return hex.EncodeToString(kp.PublicKey[:])
}

func dh(keyPair KeyPair, publicKey [32]byte) (shared [32]byte) {

	curve25519.ScalarMult(&shared, &keyPair.PrivateKey, &publicKey)

	return
}
