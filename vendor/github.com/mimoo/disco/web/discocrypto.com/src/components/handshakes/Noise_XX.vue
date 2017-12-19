<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_XX.png" alt="Noise_XX handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2>

		<p>Any <code>X</code> pattern where a peer authenticate itself via the signature of an authoritative key (like the <code>Noise_XX</code> pattern) is useful when the other peer doesn't know in advance what peer it will communicate to.</p>

		<p>This means that <code>Noise_XX</code> is a good candidate for setups where many clients try to connect to many servers, and none of the clients or servers share the same static key.</p>

		<p>Like any <code>X</code> pattern where a static key is sent, the peer needs to also send a proof which is typically a signature over its static public key from an authoritative key (a root key). With <code>Noise_XX</code>, both peers need to provide a proof, and they both need to verify each other's proof. libdisco supplies helpers to achieve both functionalities, the following examples demonstrate how to use them.</p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>In the following example of configuration, <strong>libdisco's helper functions</strong> are used to create proofs and verify them, as well as to <a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateAndSaveDiscoRootKeyPair">generate the root key</a> which can create these proofs. Notice that the configuration is the same for both peers as we're using a single root key, but two different root keys (one for the servers and one for the clients) could be used if needed.</p>

		<p>Noise_XX requires both side to authenticate themselves as part of the handshake. For this to work:</p>
		<ul>
			<li>both peers need to have their respective public static key signed by an authoritative keypair</li>
			<li>both peers need to be aware of the authoritative public key</li>
		</ul>

		<p>Both of these requirements can be achieved using libdisco's <router-link to="/protocol/Keys">key helper functions</router-link>.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_XX">here</a>. The root signing key process is illustrated <a href="https://github.com/mimoo/disco/blob/master/libdisco/examples/RootSigningKeys/root.go">here</a>.</p>

		<h3>root key:</h3>

		<p>The authoritative root signing key can be generated using libdisco's <code>GenerateAndSaveDiscoRootKeyPair</code> helper function</p>

		<pre><code>// generating the root signing key
if err := libdisco.GenerateAndSaveDiscoRootKeyPair("./privateRoot", "./publicRoot"); err != nil {
	panic("cannot generate and save a root key")
}</code></pre>

		<p>This function (<a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateAndSaveDiscoRootKeyPair">documented here</a>) will create two files, a "privateRoot" (resp. "publicRoot") file containing the private (resp. public) part of the root signing keypair.</p>

		<p>The public part can then be retrieved via the <code>LoadDiscoRootPublicKey</code> function.</p>

		<pre><code>// loading the public part
pubkey, err := libdisco.LoadDiscoRootPublicKey("./publicRoot")
if err != nil {
	// cannot load the disco root pubkey
}

// displaying the public part
fmt.Println(hex.EncodeToString(pubkey))</code></pre>

		<p>To sign a peer's static public key, the <code>CreateStaticPublicKeyProof</code> function can be used.</p>

		<pre><code>// load the private root key
privkey, err := libdisco.LoadDiscoRootPrivateKey("./privateRoot")
if err != nil {
	// couldn't load the private root key
}

// create proof where toSign is a peer's static public key
proof := libdisco.CreateStaticPublicKeyProof(privkey, toSign)

// display the proof
fmt.Println(hex.EncodeToString(proof))</code></pre>


		<h3>server:</h3>

		<pre><code>// load the server's keypair
serverKeyPair, err := libdisco.LoadDiscoKeyPair("./serverkeyPair")
if err != nil {
	fmt.Println("couldn't load the server's key pair")
	return
}

// create a verifier for when we will receive the server's public key
verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

// configure the Disco connection with Noise_XX
serverConfig := libdisco.Config{
	HandshakePattern:     libdisco.Noise_XX,
	KeyPair:              serverKeyPair,
	PublicKeyVerifier:    verifier,
	StaticPublicKeyProof: proof,
}

// listen on port 6666
listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
if err != nil {
	fmt.Println("cannot setup a listener on localhost:", err)
	return
}
addr := listener.Addr().String()
fmt.Println("listening on:", addr)</code></pre>

<h3>client:</h3>

<pre><code>// load the client's keypair
clientKeyPair, err := libdisco.LoadDiscoKeyPair("./clientkeyPair")
if err != nil {
	fmt.Println("couldn't load the client's key pair")
	return
}

// create a verifier for when we will receive the server's public key
verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

// configure the Disco connection with Noise_XX
clientConfig := libdisco.Config{
	KeyPair:              clientKeyPair,
	HandshakePattern:     libdisco.Noise_XX,
	PublicKeyVerifier:    verifier,
	StaticPublicKeyProof: proof,
}

// Dial the port 6666 of localhost
client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
if err != nil {
	fmt.Println("client can't connect to server:", err)
	return
}
defer client.Close()
fmt.Println("connected to", client.RemoteAddr())</code></pre>

	<h3>Security Considerations</h3>

	<p>The same security discussed in the <a href="https://noiseprotocol.org/noise.html#payload-security-properties">Noise specification</a> for the relevant handshake pattern apply.</p>

	<p>This handshake pattern is tricky (like any <code>X</code>-type handshakes) as it requires a Public Key Infrastructure (PKI) where:</p>

	<ul>
		<li>the root signing key is securely generated and kept in a secure location (this is often done via a <a href="https://en.wikipedia.org/wiki/Key_ceremony">key ceremony</a>)</li>
		<li>the "proofs" (a signature from the root key on a peer's static public key) are generated and passed to the peer in a secure manner</li>
		<li>keys might need to be revoked. This mean that an additional system needs to detect revocations.</li>
	</ul>

	</section>

</template>

<script>
	import patterns from '@/assets/patterns.json';
export default {
    name: 'Noise_XX',
    data () {
    	return {
    		pattern: {}
    	}
    },
    beforeMount () {
    	patterns.forEach( (pattern) => {
    		if(pattern.name == "Noise_XX") {
    			this.pattern = pattern
    		}
    	})
    }
  }
</script>
