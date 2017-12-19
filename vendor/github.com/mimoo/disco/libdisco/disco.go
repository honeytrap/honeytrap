// Package libdisco partially implements the Disco extension of the Noise protocol framework
// as specified in www.discocrypto.com/disco.html
//
// More usage helpers are available on www.discocrypto.com
//
// Author: David Wong
//
package libdisco

import (
	"errors"

	"github.com/mimoo/StrobeGo/strobe"
)

//
// SymmetricState object
//

type symmetricState struct {
	strobeState strobe.Strobe
	isKeyed     bool
}

func (s *symmetricState) initializeSymmetric(protocolName string) {
	// initializing the Strobe state
	s.strobeState = strobe.InitStrobe(protocolName, 128)
}

func (s *symmetricState) mixKey(inputKeyMaterial [32]byte) {
	s.strobeState.AD(false, inputKeyMaterial[:])
	s.isKeyed = true
}

func (s *symmetricState) mixHash(data []byte) {
	s.strobeState.AD(false, data)
}

func (s *symmetricState) mixKeyAndHash(inputKeyMaterial []byte) {
	s.strobeState.AD(false, inputKeyMaterial)
}

// TODO: documentation
// GetHandshakeHash
func (s *symmetricState) GetHandshakeHash() []byte {
	return s.strobeState.PRF(32)
}

// encrypts the plaintext and authenticates the hash
// then insert the ciphertext in the running hash
func (s *symmetricState) encryptAndHash(plaintext []byte) (ciphertext []byte, err error) {

	if s.isKeyed {
		ciphertext := s.strobeState.Send_ENC_unauthenticated(false, plaintext)
		ciphertext = append(ciphertext, s.strobeState.Send_MAC(false, 16)...)
		return ciphertext, nil
	}
	// no keys, so we don't encrypt
	return plaintext, nil
}

// decrypts the ciphertext and authenticates the hash
func (s *symmetricState) decryptAndHash(ciphertext []byte) (plaintext []byte, err error) {

	if s.isKeyed {
		if len(ciphertext) < 16 {
			return []byte{}, errors.New("disco: the received payload is shorter 16 bytes")
		}

		plaintext := s.strobeState.Recv_ENC_unauthenticated(false, ciphertext[:len(ciphertext)-16])
		ok := s.strobeState.Recv_MAC(false, ciphertext[len(ciphertext)-16:])
		if !ok {
			return []byte{}, errors.New("disco: cannot decrypt the payload")
		}
		return plaintext, nil
	}
	// no keys, so nothing to decrypt
	return ciphertext, nil
}

func (s symmetricState) Split() (s1, s2 *strobe.Strobe) {

	s1 = s.strobeState.Clone()
	s1.AD(true, []byte("initiator"))
	s1.RATCHET(32)

	s2 = &s.strobeState
	s2.AD(true, []byte("responder"))
	s2.RATCHET(32)
	return
}

//
// HandshakeState object
//

type handshakeState struct {
	// the symmetricState object
	symmetricState symmetricState
	/* Empty is a special value which indicates the variable has not yet been initialized.
	we'll use KeyPair.privateKey = 0 as Empty
	*/
	s  KeyPair // The local static key pair
	e  KeyPair // The local ephemeral key pair
	rs KeyPair // The remote party's static public key
	re KeyPair // The remote party's ephemeral public key

	// A boolean indicating the initiator or responder role.
	initiator bool
	// A sequence of message pattern. Each message pattern is a sequence
	// of tokens from the set ("e", "s", "ee", "es", "se", "ss")
	messagePatterns []messagePattern

	// A boolean indicating if the role of the peer is to WriteMessage
	// or ReadMessage
	shouldWrite bool

	// pre-shared key
	psk []byte

	// for test vectors
	debugEphemeral *KeyPair
}

// This allows you to initialize a peer.
// * see `patterns` for a list of available handshakePatterns
// * initiator = false means the instance is for a responder
// * prologue is a byte string record of anything that happened prior the Noise handshakeState
// * s, e, rs, re are the local and remote static/ephemeral key pairs to be set (if they exist)
// the function returns a handshakeState object.
func initialize(handshakeType noiseHandshakeType, initiator bool, prologue []byte, s, e, rs, re *KeyPair) (h handshakeState) {

	handshakePattern, ok := patterns[handshakeType]
	if !ok {
		panic("disco: the supplied handshakePattern does not exist")
	}

	h.symmetricState.initializeSymmetric("Noise_" + handshakePattern.name + "_25519_STROBEv1.0.2")

	h.symmetricState.mixHash(prologue)

	if s != nil {
		h.s = *s
	}
	if e != nil {
		panic("disco: fallback patterns are not implemented")
	}
	if rs != nil {
		h.rs = *rs
	}
	if re != nil {
		panic("disco: fallback patterns are not implemented")
	}

	h.initiator = initiator
	h.shouldWrite = initiator

	//Calls MixHash() once for each public key listed in the pre-messages from handshake_pattern, with the specified public key as input (see Section 7 for an explanation of pre-messages). If both initiator and responder have pre-messages, the initiator's public keys are hashed first.

	// initiator pre-message pattern
	for _, token := range handshakePattern.preMessagePatterns[0] {
		if token == token_s {
			if initiator {
				if s == nil {
					panic("disco: the static key of the client should be set")
				}
				h.symmetricState.mixHash(s.PublicKey[:])
			} else {
				if rs == nil {
					panic("disco: the remote static key of the server should be set")
				}
				h.symmetricState.mixHash(rs.PublicKey[:])
			}
		} else {
			panic("disco: token of pre-message not supported")
		}
	}

	// responder pre-message pattern
	for _, token := range handshakePattern.preMessagePatterns[1] {
		if token == token_s {
			if initiator {
				if rs == nil {
					panic("disco: the remote static key of the client should be set")
				}
				h.symmetricState.mixHash(rs.PublicKey[:])
			} else {
				if s == nil {
					panic("disco: the static key of the server should be set")
				}
				h.symmetricState.mixHash(s.PublicKey[:])
			}
		} else {
			panic("disco: token of pre-message not supported")
		}
	}

	h.messagePatterns = handshakePattern.messagePatterns

	return
}

func (h *handshakeState) writeMessage(payload []byte, messageBuffer *[]byte) (c1, c2 *strobe.Strobe, err error) {
	// is it our turn to write?
	if !h.shouldWrite {
		panic("disco: unexpected call to WriteMessage should be ReadMessage")
	}
	// do we have a token to process?
	if len(h.messagePatterns) == 0 || len(h.messagePatterns[0]) == 0 {
		panic("disco: no more tokens or message patterns to write")
	}

	// process the patterns
	for _, pattern := range h.messagePatterns[0] {

		switch pattern {

		default:
			panic("Disco: token not recognized")

		case token_e:
			// debug
			if h.debugEphemeral != nil {
				h.e = *h.debugEphemeral
			} else {
				h.e = *GenerateKeypair(nil)
			}
			*messageBuffer = append(*messageBuffer, h.e.PublicKey[:]...)
			h.symmetricState.mixHash(h.e.PublicKey[:])
			if len(h.psk) > 0 {
				h.symmetricState.mixKey(h.e.PublicKey)
			}

		case token_s:
			var ciphertext []byte
			ciphertext, err = h.symmetricState.encryptAndHash(h.s.PublicKey[:])
			if err != nil {
				return
			}
			*messageBuffer = append(*messageBuffer, ciphertext...)

		case token_ee:
			h.symmetricState.mixKey(dh(h.e, h.re.PublicKey))

		case token_es:
			if h.initiator {
				h.symmetricState.mixKey(dh(h.e, h.rs.PublicKey))
			} else {
				h.symmetricState.mixKey(dh(h.s, h.re.PublicKey))
			}

		case token_se:
			if h.initiator {
				h.symmetricState.mixKey(dh(h.s, h.re.PublicKey))
			} else {
				h.symmetricState.mixKey(dh(h.e, h.rs.PublicKey))
			}

		case token_ss:
			h.symmetricState.mixKey(dh(h.s, h.rs.PublicKey))

		case token_psk:
			h.symmetricState.mixKeyAndHash(h.psk)
		}
	}

	// Appends EncryptAndHash(payload) to the buffer
	var ciphertext []byte
	ciphertext, err = h.symmetricState.encryptAndHash(payload)
	if err != nil {
		return
	}
	*messageBuffer = append(*messageBuffer, ciphertext...)

	// are there more message patterns to process?
	if len(h.messagePatterns) == 1 {
		// If there are no more message patterns returns two new CipherState objects
		h.messagePatterns = nil
		c1, c2 = h.symmetricState.Split()
	} else {
		// remove the pattern from the messagePattern
		h.messagePatterns = h.messagePatterns[1:]
	}

	// change the direction
	h.shouldWrite = false

	return
}

// ReadMessage takes a byte sequence containing a Noise handshake message,
// and a payload_buffer to write the message's plaintext payload into.
func (h *handshakeState) readMessage(message []byte, payloadBuffer *[]byte) (c1, c2 *strobe.Strobe, err error) {
	// is it our turn to read?
	if h.shouldWrite {
		panic("disco: unexpected call to ReadMessage should be WriteMessage")
	}
	// do we have a token to process?
	if len(h.messagePatterns) == 0 || len(h.messagePatterns[0]) == 0 {
		panic("disco: no more message pattern to read")
	}

	// process the patterns
	offset := 0

	for _, pattern := range h.messagePatterns[0] {

		switch pattern {

		default:
			panic("disco: token not recognized")

		case token_e:
			if len(message[offset:]) < dhLen {
				return nil, nil, errors.New("disco: the received ephemeral key is to short")
			}
			copy(h.re.PublicKey[:], message[offset:offset+dhLen])
			offset += dhLen
			h.symmetricState.mixHash(h.re.PublicKey[:])
			if len(h.psk) > 0 {
				h.symmetricState.mixKey(h.re.PublicKey)
			}

		case token_s:
			tagLen := 0
			if h.symmetricState.isKeyed {
				tagLen = 16
			}
			if len(message[offset:]) < dhLen+tagLen {
				return nil, nil, errors.New("disco: the received static key is to short")
			}
			var plaintext []byte
			plaintext, err = h.symmetricState.decryptAndHash(message[offset : offset+dhLen+tagLen])
			if err != nil {
				return
			}
			// if we already know the remote static, compare
			copy(h.rs.PublicKey[:], plaintext)
			offset += dhLen + tagLen

		case token_ee:
			h.symmetricState.mixKey(dh(h.e, h.re.PublicKey))

		case token_es:
			if h.initiator {
				h.symmetricState.mixKey(dh(h.e, h.rs.PublicKey))
			} else {
				h.symmetricState.mixKey(dh(h.s, h.re.PublicKey))
			}

		case token_se:
			if h.initiator {
				h.symmetricState.mixKey(dh(h.s, h.re.PublicKey))
			} else {
				h.symmetricState.mixKey(dh(h.e, h.rs.PublicKey))
			}

		case token_ss:
			h.symmetricState.mixKey(dh(h.s, h.rs.PublicKey))

		case token_psk:
			h.symmetricState.mixKeyAndHash(h.psk)
		}
	}

	// Appends decrpyAndHash(payload) to the buffer
	var plaintext []byte
	plaintext, err = h.symmetricState.decryptAndHash(message[offset:])
	if err != nil {
		return
	}
	*payloadBuffer = append(*payloadBuffer, plaintext...)

	// remove the pattern from the messagePattern
	if len(h.messagePatterns) == 1 {
		// If there are no more message patterns returns two new CipherState objects
		h.messagePatterns = nil
		c1, c2 = h.symmetricState.Split()
	} else {
		h.messagePatterns = h.messagePatterns[1:]
	}

	// change the direction
	h.shouldWrite = true

	return
}

//
// Clearing stuff
//

// TODO: is there a better way to get rid of secrets in Go?
func (h *handshakeState) clear() {
	h.s.clear()
	h.e.clear()
	h.rs.clear()
	h.re.clear()
}

// TODO: is there a better way to get rid of secrets in Go?
func (kp *KeyPair) clear() {
	for i := 0; i < len(kp.PrivateKey); i++ {
		kp.PrivateKey[i] = 0
	}
	for i := 0; i < len(kp.PublicKey); i++ {
		kp.PublicKey[i] = 0
	}
}
