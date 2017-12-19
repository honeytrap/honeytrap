package libdisco

import (
	"encoding/hex"
	"testing"
)

func TestHash(t *testing.T) {

	input := []byte("hi, how are you?")

	if hex.EncodeToString(Hash(input, 32)) != "eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75" {
		t.Fatal("Hash does not produce a correct output")
	}
}

func TestSum(t *testing.T) {
	message1 := "hello"
	message2 := "how are you good sir?"
	message3 := "sure thing"
	fullmessage := message1 + message2

	// trying with NewHash with streaming and without streaming
	h1 := NewHash(32)
	h1.Write([]byte(message1))
	h1.Write([]byte(message2))
	out1 := h1.Sum()

	h2 := NewHash(32)
	h2.Write([]byte(fullmessage))
	out2 := h2.Sum()

	for idx, _ := range out1 {
		if out1[idx] != out2[idx] {
			t.Fatal("Sum function does not work")
		}
	}

	// trying with Hash()
	out3 := Hash([]byte(fullmessage), 32)

	for idx, _ := range out1 {
		if out1[idx] != out3[idx] {
			t.Fatal("Sum function does not work")
		}
	}

	// trying the streaming even more
	h1.Write([]byte(message3))
	out1 = h1.Sum()
	h2.Write([]byte(message3))
	out2 = h2.Sum()

	for idx, _ := range out1 {
		if out1[idx] != out2[idx] {
			t.Fatal("Sum function does not work")
		}
	}

	// tring with Hash()
	out3 = Hash([]byte(fullmessage+message3), 32)

	for idx, _ := range out1 {
		if out1[idx] != out3[idx] {
			t.Fatal("Sum function does not work")
		}
	}
}

func TestDeriveKeys(t *testing.T) {

	input := []byte("hi, how are you?")

	if hex.EncodeToString(DeriveKeys(input, 64)) != "d6350bb9b83884774fb9b0881680fc656be1071fff75d3fa94519d50a10b92644e3cc1cae166a60167d7bf00137018345bb8057be4b09f937b0e12066d5dc3df" {
		t.Fatal("DeriveKeys does not produce a correct output")
	}
}

func TestProtectVerifyIntegrity(t *testing.T) {
	key, _ := hex.DecodeString("eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75")

	message := []byte("hoy, how are you?")

	plaintextAndTag := ProtectIntegrity(key, message)

	retrievedMessage, err := VerifyIntegrity(key, plaintextAndTag)

	if err != nil {
		t.Fatal("Protect/Verify did not work")
	}
	for idx, _ := range message {
		if message[idx] != retrievedMessage[idx] {
			t.Fatal("Verify did not work")
		}
	}

	// tamper
	plaintextAndTag[len(plaintextAndTag)-1] += 1

	_, err = VerifyIntegrity(key, plaintextAndTag)
	if err == nil {
		t.Fatal("Verify did not work")
	}

}

func TestEncryptDecrypt(t *testing.T) {

	key, _ := hex.DecodeString("eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75")
	plaintext := []byte("hello, how are you?")

	ciphertext := Encrypt(key, plaintext)

	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatal("Encrypt/Decrypt did not work")
	}

	for idx, _ := range plaintext {
		if plaintext[idx] != decrypted[idx] {
			t.Fatal("Decrypt did not work")
		}
	}
}
