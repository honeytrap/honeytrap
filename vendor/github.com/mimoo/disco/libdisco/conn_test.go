package libdisco

import (
	"bytes"
	"testing"
)

// TODO: add more tests from tls/conn_test.go

func verifier([]byte, []byte) bool { return true }

func TestSeveralWriteRoutines(t *testing.T) {
	// init
	clientConfig := Config{
		KeyPair:              GenerateKeypair(nil),
		HandshakePattern:     Noise_XX,
		StaticPublicKeyProof: []byte{},
		PublicKeyVerifier:    verifier,
	}
	serverConfig := Config{
		KeyPair:              GenerateKeypair(nil),
		HandshakePattern:     Noise_XX,
		StaticPublicKeyProof: []byte{},
		PublicKeyVerifier:    verifier,
	}

	// get a libdisco.listener
	listener, err := Listen("tcp", "127.0.0.1:0", &serverConfig) // port 0 will find out a free port
	if err != nil {
		t.Fatal("cannot setup a listener on localhost:", err)
	}
	addr := listener.Addr().String()

	// run the server and Accept one connection
	go func(t *testing.T) {
		serverSocket, err2 := listener.Accept()
		if err2 != nil {
			t.Fatal("a server cannot accept()")
		}

		var buf [100]byte

		for {
			n, err2 := serverSocket.Read(buf[:])
			if err2 != nil {
				t.Fatal("server can't read on socket")
			}
			if !bytes.Equal(buf[:n-1], []byte("hello ")) {
				t.Fatal("received message not as expected")
			}

			//fmt.Println("server received:", string(buf[:n]))
		}

	}(t)

	// Run the client
	clientSocket, err := Dial("tcp", addr, &clientConfig)
	if err != nil {
		t.Fatal("client can't connect to server")
	}

	for i := 0; i < 100; i++ {
		go func(i int) {
			message := "hello " + string(i)
			_, err = clientSocket.Write([]byte(message))
			if err != nil {
				t.Fatal("client can't write on socket")
			}
		}(i)
	}
}

func TestHalfDuplex(t *testing.T) {
	// init
	clientConfig := Config{
		KeyPair:              GenerateKeypair(nil),
		HandshakePattern:     Noise_XX,
		StaticPublicKeyProof: []byte{},
		PublicKeyVerifier:    verifier,
		HalfDuplex:           true,
	}
	serverConfig := Config{
		KeyPair:              GenerateKeypair(nil),
		HandshakePattern:     Noise_XX,
		StaticPublicKeyProof: []byte{},
		PublicKeyVerifier:    verifier,
		HalfDuplex:           true,
	}

	// get a libdisco.listener
	listener, err := Listen("tcp", "127.0.0.1:0", &serverConfig)
	if err != nil {
		t.Fatal("cannot setup a listener on localhost:", err)
	}
	addr := listener.Addr().String()

	// run the server and Accept one connection
	go func() {
		serverSocket, err2 := listener.Accept()
		if err2 != nil {
			t.Fatal("a server cannot accept()")
		}

		var buf [100]byte

		for {
			n, err2 := serverSocket.Read(buf[:])
			if err2 != nil {
				t.Fatal("server can't read on socket")
			}
			if !bytes.Equal(buf[:n-1], []byte("hello ")) {
				t.Fatal("received message not as expected")
			}

			//fmt.Println("server received:", string(buf[:n]))
		}

	}()

	// Run the client
	clientSocket, err := Dial("tcp", addr, &clientConfig)
	if err != nil {
		t.Fatal("client can't connect to server")
	}

	for i := 0; i < 100; i++ {
		go func(i int) {
			message := "hello " + string(i)
			_, err = clientSocket.Write([]byte(message))
			if err != nil {
				t.Fatal("client can't write on socket")
			}
		}(i)
	}
}
