// These Utility functions implement the net.Conn interface. Most of this code
// was either taken directly or inspired from Go's crypto/tls package.
package libdisco

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ed25519"
)

// Server returns a new Disco server side connection
// using net.Conn as the underlying transport.
// The configuration config must be non-nil and must include
// at least one certificate or else set GetCertificate.
func Server(conn net.Conn, config *Config) *Conn {
	return &Conn{conn: conn, config: config}
}

// Client returns a new Disco client side connection
// using conn as the underlying transport.
// The config cannot be nil: users must set either ServerName or
// InsecureSkipVerify in the config.
func Client(conn net.Conn, config *Config) *Conn {
	return &Conn{conn: conn, config: config, isClient: true}
}

// A listener implements a network listener (net.Listener) for Disco connections.
type listener struct {
	net.Listener
	config *Config
}

// Accept waits for and returns the next incoming Disco connection.
// The returned connection is of type *Conn.
func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return Server(c, l.config), nil
}

// Listen creates a Disco listener accepting connections on the
// given network address using net.Listen.
// The configuration config must be non-nil.
func Listen(network, laddr string, config *Config) (net.Listener, error) {
	// check Config
	if config == nil {
		return nil, errors.New("Disco: no Config set")
	}
	if err := checkRequirements(false, config); err != nil {
		panic(err)
	}

	// make net.Conn listen
	l, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}

	// create new libdisco.listener
	discoListener := new(listener)
	discoListener.Listener = l
	discoListener.config = config
	return discoListener, nil
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "Disco: DialWithDialer timed out" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// this functions checks if at some point in the protocol
// the peer needs to verify the other peer static public key
// and if the peer needs to provide a proof for its static public key
var errNoPubkeyVerifier = errors.New("Disco: no public key verifier set in Config")
var errNoProof = errors.New("Disco: no public key proof set in Config")

func checkRequirements(isClient bool, config *Config) (err error) {
	ht := config.HandshakePattern
	if ht == Noise_NX || ht == Noise_KX || ht == Noise_XX || ht == Noise_IX {
		if isClient && config.PublicKeyVerifier == nil {
			return errNoPubkeyVerifier
		} else if !isClient && config.StaticPublicKeyProof == nil {
			return errNoProof
		}
	}
	if ht == Noise_XN || ht == Noise_XK || ht == Noise_XX || ht == Noise_X || ht == Noise_IN || ht == Noise_IK || ht == Noise_IX {
		if isClient && config.StaticPublicKeyProof == nil {
			return errNoProof
		} else if !isClient && config.PublicKeyVerifier == nil {
			return errNoPubkeyVerifier
		}
	}
	if ht == Noise_NNpsk2 && len(config.PreSharedKey) != 32 {
		return errors.New("noise: a 32-byte pre-shared key needs to be passed as noise.Config")
	}
	return nil
}

// DialWithDialer connects to the given network address using dialer.Dial and
// then initiates a Disco handshake, returning the resulting Disco connection. Any
// timeout or deadline given in the dialer apply to connection and Disco
// handshake as a whole.
//
// DialWithDialer interprets a nil configuration as equivalent to the zero
// configuration; see the documentation of Config for the defaults.
// TODO: make sure sane defaults for time outs are set!!!
func DialWithDialer(dialer *net.Dialer, network, addr string, config *Config) (*Conn, error) {
	// We want the Timeout and Deadline values from dialer to cover the
	// whole process: TCP connection and Disco handshake. This means that we
	// also need to start our own timers now.
	timeout := dialer.Timeout

	if !dialer.Deadline.IsZero() {
		deadlineTimeout := time.Until(dialer.Deadline)
		if timeout == 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}

	// check Config
	if config == nil {
		panic("Disco: no Config set")
	}

	if err := checkRequirements(true, config); err != nil {
		panic(err)
	}

	// Dial the net.Conn first
	var errChannel chan error

	if timeout != 0 {
		errChannel = make(chan error, 2)
		time.AfterFunc(timeout, func() {
			errChannel <- timeoutError{}
		})
	}

	rawConn, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	// TODO: use the following code to implement some sort of SNI extension?
	/*
		colonPos := strings.LastIndex(addr, ":")
		if colonPos == -1 {
			colonPos = len(addr)
		}
		hostname := addr[:colonPos]
	*/

	// Create the libdisco.Conn
	conn := Client(rawConn, config)

	// Do the handshake
	if timeout == 0 {
		err = conn.Handshake()
	} else {
		go func() {
			errChannel <- conn.Handshake()
		}()

		err = <-errChannel
	}

	if err != nil {
		rawConn.Close()
		return nil, err
	}

	return conn, nil
}

// Dial connects to the given network address using net.Dial
// and then initiates a Disco handshake, returning the resulting
// Disco connection.
// Dial interprets a nil configuration as equivalent to
// the zero configuration; see the documentation of Config
// for the defaults.
func Dial(network, addr string, config *Config) (*Conn, error) {
	return DialWithDialer(new(net.Dialer), network, addr, config)
}

//
// Authentication helpers
//

// CreatePublicKeyVerifier can be used to create the callback
// function PublicKeyVerifier sometimes required in a libdisco.Config
// for peers that are receiving a static public key at some
// point during the handshake
func CreatePublicKeyVerifier(rootPublicKey ed25519.PublicKey) func([]byte, []byte) bool {
	return func(publicKey, proof []byte) bool {
		// ed25519.Verify panics if len(publicKey) is not PublicKeySize. We need to avoid that
		if len(publicKey) != 32 {
			return false
		}
		return ed25519.Verify(rootPublicKey, publicKey, proof)
	}
}

// CreateStaticPublicKeyProof can be used to create the proof
// StaticPublicKeyProof sometimes required in a libdisco.Config
// for peers that are sending their static public key at some
// point during the handshake
func CreateStaticPublicKeyProof(rootPrivateKey ed25519.PrivateKey, publicKey []byte) []byte {

	if len(publicKey) != 32 {
		panic("disco: length of public key passed is incorrect (should be 32)")
	}

	signature, err := rootPrivateKey.Sign(rand.Reader, publicKey, crypto.Hash(0))
	if err != nil {
		panic("disco: can't create static public key proof")
	}
	return signature
}

//
// Storage of Disco Signing Root Keys
//

// GenerateAndSaveDiscoRootKeyPair generates an ed25519 root key pair and save the private and public parts in different files.
func GenerateAndSaveDiscoRootKeyPair(discoRootPrivateKeyFile string, discoRootPublicKeyFile string) (err error) {
	// TODO: should I require a passphrase and encrypt it with it?
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	var publicKeyHex [32 * 2]byte
	var privateKeyHex [64 * 2]byte
	hex.Encode(publicKeyHex[:], publicKey)
	hex.Encode(privateKeyHex[:], privateKey)

	err = ioutil.WriteFile(discoRootPrivateKeyFile, privateKeyHex[:], 0400)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(discoRootPublicKeyFile, publicKeyHex[:], 0644)
	return err
}

// LoadDiscoRootPublicKey reads and parses a public Root key from a
// file. The file contains an 32-byte ed25519 public key in hexadecimal
func LoadDiscoRootPublicKey(discoRootPublicKey string) (rootPublicKey ed25519.PublicKey, err error) {
	publicKeyHex, err := ioutil.ReadFile(discoRootPublicKey)
	if err != nil {
		return nil, err
	}
	if len(publicKeyHex) != 32*2 {
		return nil, errors.New("Disco: Disco root public key file is not correctly formated")
	}
	publicKey := make([]byte, 32)
	_, err = hex.Decode(publicKey[:], publicKeyHex)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

// LoadDiscoRootPrivateKey reads and parses a private Root key from a
// file. The file contains an 32-byte ed25519 private key in hexadecimal
func LoadDiscoRootPrivateKey(discoRootPrivateKey string) (rootPrivateKey ed25519.PrivateKey, err error) {
	// TODO: should I require a passphrase to decrypt it?
	privateKeyHex, err := ioutil.ReadFile(discoRootPrivateKey)
	if err != nil {
		return nil, err
	}
	if len(privateKeyHex) != 64*2 {
		return nil, errors.New("Disco: Disco root private key file is not correctly formated")
	}
	privateKey := make([]byte, 64)
	_, err = hex.Decode(privateKey[:], privateKeyHex)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

//
// Storage of Disco Static Keys
//

// GenerateAndSaveDiscoKeyPair generates a disco key pair (X25519 key pair)
// and saves it to a file in hexadecimal form. You can use ExportPublicKey() to
// export the public key part.
func GenerateAndSaveDiscoKeyPair(DiscoKeyPairFile string) (keyPair *KeyPair, err error) {

	// TODO: should I require a passphrase and encrypt it with it?
	keyPair = GenerateKeypair(nil)
	var dataToWrite [128]byte
	hex.Encode(dataToWrite[:64], keyPair.PrivateKey[:])
	hex.Encode(dataToWrite[64:], keyPair.PublicKey[:])
	err = ioutil.WriteFile(DiscoKeyPairFile, dataToWrite[:], 0400)
	if err != nil {
		return nil, errors.New("Disco: could not write on file at path")
	}
	return keyPair, nil
}

// LoadDiscoKeyPair reads and parses a public/private key pair from a pair
// of files.
func LoadDiscoKeyPair(discoKeyPairFile string) (keypair *KeyPair, err error) {
	// TODO: should I require a passphrase to decrypt it?
	keyPairString, err := ioutil.ReadFile(discoKeyPairFile)
	if err != nil {
		return nil, err
	}
	if len(keyPairString) != 64*2 {
		return nil, errors.New("Disco: Disco key pair file is not correctly formated")
	}
	var keyPair KeyPair
	_, err = hex.Decode(keyPair.PrivateKey[:], keyPairString[:64])
	if err != nil {
		return nil, err
	}
	_, err = hex.Decode(keyPair.PublicKey[:], keyPairString[64:])
	if err != nil {
		return nil, err
	}

	return &keyPair, nil
}
