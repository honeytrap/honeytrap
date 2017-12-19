package main

import (
	"fmt"
	"net"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// generating the server key pair
	serverKeyPair := libdisco.GenerateKeypair(nil)

	// configuring the Disco connection with a Noise_NK handshake
	// in which the client already knows the server's public key
	serverConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_NK,
		KeyPair:          serverKeyPair,
	}
	// listen on port 6666
	listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
	if err != nil {
		fmt.Println("cannot setup a listener on localhost:", err)
		return
	}
	addr := listener.Addr().String()
	fmt.Println("listening on:", addr)
	// export public key so that client can retrieve it out of band
	fmt.Println("server's public key:", serverKeyPair.ExportPublicKey())

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
