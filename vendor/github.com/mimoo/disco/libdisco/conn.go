package libdisco

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/mimoo/StrobeGo/strobe"
)

// A Conn represents a secured connection.
// It implements the net.Conn interface.
type Conn struct {
	conn     net.Conn
	isClient bool

	// handshake
	config            *Config // configuration passed to constructor
	hs                handshakeState
	handshakeComplete bool
	handshakeMutex    sync.Mutex

	// Authentication thingies
	isRemoteAuthenticated bool

	// input/output
	in, out         *strobe.Strobe
	inLock, outLock sync.Mutex
	inputBuffer     []byte

	// half duplex
	isHalfDuplex   bool
	halfDuplexLock sync.Mutex
}

// Access to net.Conn methods.
// Cannot just embed net.Conn because that would
// export the struct field too.

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated with the connection.
// A zero value for t means Read and Write will not time out.
// After a Write has timed out, the Disco state is corrupt and all future writes will return the same error.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline on the underlying connection.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline on the underlying connection.
// A zero value for t means Write will not time out.
// After a Write has timed out, the Disco state is corrupt and all future writes will return the same error.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// Write writes data to the connection.
func (c *Conn) Write(b []byte) (int, error) {

	//
	if hp := c.config.HandshakePattern; !c.isClient && (hp == Noise_N || hp == Noise_K || hp == Noise_X) {
		panic("disco: a server should not write on one-way patterns")
	}

	// Make sure to go through the handshake first
	if err := c.Handshake(); err != nil {
		return 0, err
	}

	// Lock the write socket
	if c.isHalfDuplex {
		c.halfDuplexLock.Lock()
		defer c.halfDuplexLock.Unlock()
	} else {
		c.outLock.Lock()
		defer c.outLock.Unlock()
	}

	// process the data in a loop
	var n int
	data := b
	for len(data) > 0 {

		// fragment the data
		m := len(data)
		if m > NoiseMaxPlaintextSize {
			m = NoiseMaxPlaintextSize
		}

		// Encrypt
		ciphertext := c.out.Send_ENC_unauthenticated(false, data[:m])
		ciphertext = append(ciphertext, c.out.Send_MAC(false, 16)...)

		// header (length)
		length := []byte{byte(len(ciphertext) >> 8), byte(len(ciphertext) % 256)}

		// Send data
		_, err := c.conn.Write(append(length, ciphertext...))
		if err != nil {
			return n, err
		}
		/*
			// TODO: should we test if we sent the correct number of bytes?
			if _ != len(ciphertext) {
				return errors.New("disco: cannot send the whole data")
			}
		*/

		// prepare next loop iteration
		n += m
		data = data[m:]
	}

	return n, nil
}

// Read can be made to time out and return a net.Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	// Make sure to go through the handshake first
	if err = c.Handshake(); err != nil {
		return
	}

	// Put this after Handshake, in case people were calling
	// Read(nil) for the side effect of the Handshake.
	if len(b) == 0 {
		return
	}

	// If this is a one-way pattern, do some checks
	if hp := c.config.HandshakePattern; c.isClient && (hp == Noise_N || hp == Noise_K || hp == Noise_X) {
		panic("disco: a client should not read on one-way patterns")
	}

	// Lock the read socket
	if c.isHalfDuplex {
		c.halfDuplexLock.Lock()
		defer c.halfDuplexLock.Unlock()
	} else {
		c.inLock.Lock()
		defer c.inLock.Unlock()
	}

	// read whatever there is to read in the buffer
	readSoFar := 0
	if len(c.inputBuffer) > 0 {
		copy(b, c.inputBuffer)
		if len(c.inputBuffer) >= len(b) {
			c.inputBuffer = c.inputBuffer[len(b):]
			return len(b), nil
		}
		readSoFar += len(c.inputBuffer)
		c.inputBuffer = c.inputBuffer[:0]
	}

	// read header from socket
	bufHeader, err := readFromUntil(c.conn, 2)
	if err != nil {
		return 0, err
	}
	length := (int(bufHeader[0]) << 8) | int(bufHeader[1])
	if length > NoiseMessageLength {
		return 2, errors.New("disco: Disco message received exceeds DiscoMessageLength")
	}

	// read noise message from socket
	noiseMessage, err := readFromUntil(c.conn, length)
	if err != nil {
		return 2, err
	}

	// decrypt
	if length < 16 {
		return 2 + length, errors.New("disco: the received payload is shorter 16 bytes")
	}

	plaintext := c.in.Recv_ENC_unauthenticated(false, noiseMessage[:len(noiseMessage)-16])
	ok := c.in.Recv_MAC(false, noiseMessage[len(noiseMessage)-16:])
	if !ok {
		return 2 + length, errors.New("disco: cannot decrypt the payload")
	}

	// append to the input buffer
	c.inputBuffer = append(c.inputBuffer, plaintext...)

	// read whatever we can read
	rest := len(b) - readSoFar
	copy(b[readSoFar:], c.inputBuffer)
	if len(c.inputBuffer) >= rest {
		c.inputBuffer = c.inputBuffer[rest:]
		return len(b), nil
	}

	// we haven't filled the buffer
	readSoFar += len(c.inputBuffer)
	c.inputBuffer = c.inputBuffer[:0]
	return readSoFar, nil

	// TODO: should we continue to try and read other messages?

}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

//
// Disco-related functions
//

// Handshake runs the client or server handshake protocol if
// it has not yet been run.
// Most uses of this package need not call Handshake explicitly:
// the first Read or Write will call it automatically.
func (c *Conn) Handshake() error {

	// Locking the handshakeMutex
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()

	// did we already go through the handshake?
	if c.handshakeComplete {
		return nil
	}

	// Disco.initialize(handshakePattern string, initiator bool, prologue []byte, s, e, rs, re *KeyPair) (h handshakeState)
	var remoteKeyPair *KeyPair
	if c.config.RemoteKey != nil {
		if len(c.config.RemoteKey) != 32 {
			return errors.New("disco: the provided remote key is not 32-byte")
		}
		remoteKeyPair = &KeyPair{}
		copy(remoteKeyPair.PublicKey[:], c.config.RemoteKey)
	}
	hs := initialize(c.config.HandshakePattern, c.isClient, c.config.Prologue, c.config.KeyPair, nil, remoteKeyPair, nil)

	// pre-shared key
	hs.psk = c.config.PreSharedKey

	// start handshake
	var c1, c2 *strobe.Strobe
	var err error
	var receivedPayload []byte
ContinueHandshake:
	if hs.shouldWrite {
		// we're writing the next message pattern
		// if it's the message pattern and we're sending a static key, we also send a proof
		// TODO: is this the best way of sending a proof :/ ?
		var bufToWrite []byte
		var proof []byte
		if len(hs.messagePatterns) <= 2 {
			proof = c.config.StaticPublicKeyProof
		}
		c1, c2, err = hs.writeMessage(proof, &bufToWrite)
		if err != nil {
			return err
		}
		// header (length)
		length := []byte{byte(len(bufToWrite) >> 8), byte(len(bufToWrite) % 256)}
		// write
		_, err = c.conn.Write(append(length, bufToWrite...))
		if err != nil {
			return err
		}

	} else {
		// we're reading the next message pattern, as well as reacting to any received data
		bufHeader, err := readFromUntil(c.conn, 2) // length header
		if err != nil {
			return err
		}
		length := (int(bufHeader[0]) << 8) | int(bufHeader[1])
		if length > NoiseMessageLength {
			return errors.New("disco: Disco message received exceeds DiscoMessageLength")
		}
		noiseMessage, err := readFromUntil(c.conn, length) // noise message
		if err != nil {
			return err
		}

		c1, c2, err = hs.readMessage(noiseMessage, &receivedPayload)
		if err != nil {
			return err
		}
	}

	// handshake not finished
	if c1 == nil {
		goto ContinueHandshake
	}

	// setup the Write and Read secure channels
	if c1 == nil {
		return errors.New("noise: the handshake did not return a secure channel to Write and Read from")
	}

	// Has the other peer been authenticated so far?
	if !c.isRemoteAuthenticated && c.config.PublicKeyVerifier != nil {
		// test if remote static key is empty
		isRemoteStaticKeySet := byte(0)
		for _, val := range hs.rs.PublicKey {
			isRemoteStaticKeySet |= val
		}
		if isRemoteStaticKeySet != 0 {
			// a remote static key has been received. Verify it
			if !c.config.PublicKeyVerifier(hs.rs.PublicKey[:], receivedPayload) {
				return errors.New("disco: the received public key could not be authenticated")
			}
		}
	}

	// Processing the final handshake message returns two CipherState objects
	// the first for encrypting transport messages from initiator to responder
	// and the second for messages in the other direction.
	if c2 != nil {
		if c.isClient {
			c.out, c.in = c1, c2
		} else {
			c.out, c.in = c2, c1
		}
	} else {
		c.isHalfDuplex = true
		c.in = c1
		c.out = c1
	}

	// TODO: preserve c.hs.symmetricState.h
	// At that point the HandshakeState should be deleted except for the hash value h, which may be used for post-handshake channel binding (see Section 11.2).
	c.hs.clear()

	// no errors :)
	c.handshakeComplete = true
	return nil
}

// IsRemoteAuthenticated can be used to check if the remote peer has been properly authenticated. It serves no real purpose for the moment as the handshake will not go through if a peer is not properly authenticated in patterns where the peer needs to be authenticated.
func (c *Conn) IsRemoteAuthenticated() bool {
	return c.isRemoteAuthenticated
}

//
// input/output functions
//

func readFromUntil(r io.Reader, n int) ([]byte, error) {
	result := make([]byte, n)
	offset := 0
	for {
		m, err := r.Read(result[offset:])
		if err != nil {
			return result, err
		}
		offset += m
		if offset == n {
			break
		}
	}
	return result, nil
}

/*
TODO: Do we need such a function? (this comes from go.TLS)
// ConnectionState returns basic Disco details about the connection.
func (c *Conn) ConnectionState() ConnectionState {
}
*/
