package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// generating the server key pair
	serverKeyPair := libdisco.GenerateKeypair(nil)
	fmt.Println("server's public key:", serverKeyPair.ExportPublicKey())

	// configuring the Disco connection
	serverConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_K,
		KeyPair:          serverKeyPair,
	}

	// retrieve the client's public key from an argument
	fmt.Println("please enter the client's public key in hexadecimal")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	clientKey, _ := hex.DecodeString(scanner.Text())
	serverConfig.RemoteKey = clientKey

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
