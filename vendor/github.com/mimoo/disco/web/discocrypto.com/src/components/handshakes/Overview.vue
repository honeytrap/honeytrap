<template>
	<section class="content">

	    <h1 class="title"><i class="fa fa-exchange"></i> Protocol Overview</h1>

		<p>
			<strong>libdisco</strong>'s protocol is based on the <a href="/disco.html"><i class="fa fa-file-text-o" aria-hidden="true"></i> Disco extension</a> of the <a href="https://noiseprotocol.org">Noise protocol framework</a>. What this means to you is that the library supports a subset (although potentially all) of the handshakes that are specified in the Noise protocol framework.
		</p>
		<p>
			Theses <strong>handshakes</strong> are different ways to <strong>setup a secure connection</strong> that your application can use to encrypt data between two endpoints. They have names like <code>Noise_XX</code> and <code>Noise_NK</code> and are explained (along with examples on how to use them) in this documentation. See the <i class="fa fa-bars" aria-hidden="true"></i> menu on the left. <br>
			By the way, the letters after the "Noise_" have a meaning! You do not have to learn it but it can help you in your choice. <a href="https://noiseprotocol.org/noise.html#one-way-patterns">See the Noise specification for more information</a>.
		</p>
		<p>
			In addition, we provide the following short <strong>Quizz</strong> to help you figure out what is the best way for you to connect your endpoints (or <strong>peers</strong>) securely.
		</p>


		<Quizz></Quizz>


		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Configuration</h2>

		<p>General documentation for the protocol parts of libdisco can be found on <a href="https://godoc.org/github.com/mimoo/disco/libdisco"><img src="https://godoc.org/github.com/mimoo/disco/libdisco?status.svg" alt="GoDoc" style="vertical-align: text-top;"></a></p>

		<p>A <code>libdisco.Config</code> is mandatory for setting up both clients and servers.</p>

		<pre><code>type Config struct {
  HandshakePattern     noiseHandshakeType
  KeyPair              *KeyPair
  RemoteKey            []byte
  Prologue             []byte
  StaticPublicKeyProof []byte
  PublicKeyVerifier    func(publicKey, proof []byte) bool
  PreSharedKey         []byte
  HalfDuplex           bool
}</code></pre>

		<p>
			<strong>HandshakePattern</strong>: You will have to choose a handshake pattern from the list of implemented patterns first. We've included some explanations in the documentation and on this page, but the Noise specification contains the most amount of information about these. If the client and the server do not choose the same handshake pattern, they will not succeed in creating a secure channel. (If something is not clear, or if a pattern has not been implemented, please use the issues on this repo to tell us.)
		</p>

		<p>
			<strong>KeyPair</strong>: if the handshake pattern chosen requires the peer to be initialized with a static key (because it will send its static key to the other peer during the handshake), this should be filled with a X25519 KeyPair structure. Several utility functions exist to create and load one, see <code>GenerateKeypair()</code>, <code>GenerateAndSaveDiscoKeyPair()</code> and <code>LoadDiscoKeyPair()</code> in the documentation.
		</p>

		<p>
			<strong>RemoteKey</strong>: if the handshake pattern chosen requires the peer to be initialized with the static key of the other peer (because it is supposed to know its peer's static key. Think about public-key pinning). This should be a 32-byte X25519 public key. A peer's public key can be obtained via the <code>KeyPair.ExtractPublicKey()</code> function.
		</p>

		<p>
			<strong>Prologue</strong>: any messages that have been exchanged between a client and a server, prior to the encryption of the channel via Disco, can be authenticated via the prologue. This means that if a man-in-the-middle attacker has removed, added or re-ordered messages prior to setting up a Disco channel, the client and the servers will not be able to setup a secure channel with Noise (and thus will inform both peers that the prologue information is not the same on both sides). To use this, simply concatenate all these messages (on both the client and the server) and pass them in the prologue value.
		</p>

		<p>
			<strong>StaticPublicKeyProof</strong>: if the handshake pattern chosen has the peer send its static public key at some point in the handshake, the peer might need to provide a "proof" that the public key is "legit". For example, the <code>StaticPublicKeyProof</code> can be a signature over the peer's static public key from an authoritative root key. This "proof" will be sent as part of the handshake, possibly non-encrypted and visible to passive observers. More information is available in the Disco Keys section.
		</p>

		<p>
			<strong>PublicKeyVerifier</strong>: if the handshake pattern chosen has the peer receive a static public key at some point in the handshake, then the peer needs a function to verify the validity of the received key. During the handshake a "proof" might have been sent. <code>PublicKeyVerifier</code> is a callback function that must be implemented by the application using Disco and that will be called on both the static public key that has been received and any payload that has been received so far (usually the payload sent by the previous <code>StaticPublicKeyProof</code> function). If this function returns true, the handshake will continue. Otherwise the handshake will fail. More information is available in the Disco Keys section.
		</p>

		<p>
			<strong>PreSharedKey</strong>: if the handshake pattern chosen requires both peers to be aware of a shared secret (of 32-byte), this pre-shared secret must be shared in the configuration prior to starting the handshake.
		</p>

		<p>
			<strong>HalfDuplex</strong>: In some situation, one of the peer might be constrained by the size of its memory. In such scenarios, communication over a single writing channel might be a solution. Disco provides half-duplex channels where the client and the server take turn to write or read on the secure channel. For this to work this value must be set to true on both side of the connection. The server and client MUST NOT write or read on the secure channel at the same time.
		</p>

		<article class="message is-info">
		  <div class="message-header">
		    <p>Authenticated?</p>
		  </div>
		  <div class="message-body">
		    What do we mean by <strong>authenticated</strong>? When you visit <code>https://www.google.com</code>, your browser authenticates the webserver (and thus <code>google.com</code> is authenticated) but the browser is generally not authenticated (or it is later via a simple form asking you for credentials). To authenticate <code>google.com</code>, the webserver provides a signature from an authority that the web browser trust and thus can verify.
		  </div>
		</article>

<h3><i class="fa fa-caret-right" aria-hidden="true"></i> Server</h3>

	<p>
	Simply use the <code>Listen()</code> and <code>Accept()</code> paradigm. You then get an object implementing the <code>net.Conn</code> interface. You can then <code>Write()</code> and <code>Read()</code>.</p>

	<p>
	The following example use the Noise_NK handshake where the client is not authenticated and the server's key is known to the client in advance.
	</p>

	<pre><code>package main

import (
	"fmt"

	"github.com/mimoo/disco/libdisco"
)

func main() {

	serverKeyPair := libdisco.GenerateKeypair(nil)

	serverConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_NK,
		KeyPair:          serverKeyPair,
	}

	listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
	if err != nil {
		fmt.Println("cannot setup a listener on localhost:", err)
		return
	}
	addr := listener.Addr().String()
	fmt.Println("listening on:", addr)
	fmt.Println("server's public key:", serverKeyPair.ExportPublicKey())

	server, err := listener.Accept()
	if err != nil {
		fmt.Println("server cannot accept()")
		return
	}
	defer server.Close()

	buf := make([]byte, 100)
	for {
		n, err := server.Read(buf)
		if err != nil {
			fmt.Println("server can't read on socket", err)
			return
		}
		fmt.Println("server received some data:", string(buf[:n]))
	}
}</code></pre>

	<h3><i class="fa fa-caret-right" aria-hidden="true"></i> Client</h3>

	<p>The client can simply use the <code>Dial()</code> paradigm using the public key of the server:</p>
	
	<pre><code>package main

import (
	"encoding/hex"
	"fmt"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// replace this with the server's public key!
	serverKey, _ := hex.DecodeString("e424214ab16f56def7778e9a3d36d891221c4f5b38c8b2679ccbdaed5c27e735")
	clientConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_NK,
		RemoteKey:        serverKey,
	}

	client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
	if err != nil {
		fmt.Println("client can't connect to server:", err)
		return
	}
	defer client.Close()

	for {
		var in string
		fmt.Scanf("%s", &in)
		_, err = client.Write([]byte(in))
		if err != nil {
			fmt.Println("client can't write on socket:", err)
		}
	}
}</code></pre>


		<h3><i class="fa fa-caret-right" aria-hidden="true"></i> Handshake Patterns</h3>

		<p>The following is a list of all the supported handshake patterns and a brief description for each of them.</p>



		<div class="box" v-for="pattern in patterns">
	      <div class="content">
	        <p>
	          <strong><router-link :to="'/handshakes/' + pattern.name">{{pattern.name}}</router-link></strong> <span class="tag" v-for="tag in pattern.tags">{{tag}}</span>
	          <div v-html="pattern.description"></div>
	        </p>
	      </div>
		</div>


		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Examples</h2>

		<p>Further examples for all supported handshake patterns can be found in <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples" target="_blank">the github repository</a>.</p>

		<p>If you need help, head to the <a href="https://github.com/mimoo/disco/issues"><i class="fa fa-question" aria-hidden="true"></i> issues on github</a> or the <a href="https://www.reddit.com/r/discocrypto/"><i class="fa fa-envelope-o" aria-hidden="true"></i>
     subreddit over at r/discocrypto</a>.</p>

	</section>
</template>

<script>
  import Quizz from '@/components/Quizz'
	import patterns from '@/assets/patterns.json';

  export default {
    name: 'protocolOverview',
    components: {
      Quizz
    },
    data () {
    	return {
    		patterns: []
    	}
    },
    beforeMount () {
    	this.patterns = patterns
    }
  }
</script>