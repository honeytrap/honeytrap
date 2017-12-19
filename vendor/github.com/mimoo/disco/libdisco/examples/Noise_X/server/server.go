package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// usage
	if len(os.Args) != 2 {
		fmt.Println("usage:go run server.go hex_root_public_key")
		return
	}

	// generating the server key pair
	serverKeyPair := libdisco.GenerateKeypair(nil)
	fmt.Println("server's public key:", serverKeyPair.ExportPublicKey())

	// retrieve root key
	rootPublicKey, err := hex.DecodeString(os.Args[1])
	if err != nil || len(rootPublicKey) != 32 {
		fmt.Println("public root key passed is not a 32-byte value in hexadecimal (", len(rootPublicKey), ")")
		return
	}

	// create a verifier for when we will receive the server's public key
	verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

	// configuring the Disco connection
	// in which the client already knows the server's public key
	serverConfig := libdisco.Config{
		HandshakePattern:  libdisco.Noise_X,
		KeyPair:           serverKeyPair,
		PublicKeyVerifier: verifier,
	}

	// listen on port 6666
	listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
	if err != nil {
		fmt.Println("cannot setup a listener on localhost:", err)
		return
	}
	addr := listener.Addr().String()
	fmt.Println("listening on:", addr)

	for {
		// accept a connection
		server, err := listener.Accept()
		if err != nil {
			fmt.Println("server cannot accept()")
			server.Close()
			continue
		}
		fmt.Println("server accepted connection from", server.RemoteAddr())
		// read what the socket has to say until connection is closed
		go func(server net.Conn) {
			buf := make([]byte, 100)
			for {
				n, err := server.Read(buf)
				if err != nil {
					fmt.Println("server can't read on socket for", server.RemoteAddr(), ":", err)
					break
				}
				fmt.Println("received data from", server.RemoteAddr(), ":", string(buf[:n]))
			}
			fmt.Println("shutting down the connection with", server.RemoteAddr())
			server.Close()
		}(server)

	}

}
