<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_X.png" alt="Noise_X handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2>

		<p>This handshake pattern is useful when many different clients want to talk to one server while authenticating both side of the connection. In addition, since it is a one-way pattern, the server never talks back to them.</p>

		<p>If client authentication is not needed, refer to <router-link to="/protocol/Noise_N">Noise_N</router-link>. If a single client exist, refer to <router-link to="/protocol/Noise_K">Noise_K</router-link>.</p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>The client needs to have prior knowledge to the server's public static key in order to authenticate the server during the handshake phase.</p>

		<p>As for the client's authentication:</p>
		<ul>
			<li>the client needs to have its public static key signed by an authoritative key pair</li>
			<li>the server needs to be aware of the authoritative public key</li>
		</ul>

		<p>Both can be done using libdisco's <router-link to="/protocol/Keys">key helper functions</router-link>.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_N">here</a>. The root signing key process is illustrated <a href="https://github.com/mimoo/disco/blob/master/libdisco/examples/RootSigningKeys/root.go">here</a>.</p>

		<h3>root key:</h3>

		<p>The authoritative root signing key can be generated using libdisco's <code>GenerateAndSaveDiscoRootKeyPair</code> helper function</p>

		<pre><code>// generating the root signing key
if err := libdisco.GenerateAndSaveDiscoRootKeyPair("./privateRoot", "./publicRoot"); err != nil {
	panic("cannot generate and save a root key")
}</code></pre>

		<p>This function (<a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateAndSaveDiscoRootKeyPair">documented here</a>) will create two files, a "privateRoot" (resp. "publicRoot") file containing the private (resp. public) part of the root signing key pair.</p>

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
serverKeyPair, err := libdisco.LoadDiscoKeyPair("./serverKeyPair")
if err != nil {
	fmt.Println("couldn't load the server's key pair")
	return
}

// retrieve the root signing public key
rootPublicKey, err := hex.DecodeString(...)
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
fmt.Println("listening on:", addr)</code></pre>

		<h3>client:</h3>

		<pre><code>// load the client's keypair
clientKeyPair, err := libdisco.LoadDiscoKeyPair("./clientkeyPair")
if err != nil {
	fmt.Println("couldn't load the client's key pair")
	return
}

// retrieve server's static public key
serverPublicKey, err := hex.DecodeString(...)
if err != nil || len(serverPublicKey) != 32 {
	fmt.Println("server's static public key passed is not a 32-byte value in hexadecimal (", len(serverPublicKey), ")")
	return
}

// retrieve signature/proof over the client's static public key
proof, err := hex.DecodeString(...)
if err != nil || len(proof) != 64 {
	fmt.Println("proof passed is not a 64-byte value in hexadecimal (", len(proof), ")")
	return
}

// configure the Disco connection with Noise_XX
clientConfig := libdisco.Config{
	KeyPair:              clientKeyPair,
	RemoteKey:            serverPublicKey,
	HandshakePattern:     libdisco.Noise_X,
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

	</section>

</template>

<script>
import patterns from '@/assets/patterns.json';

export default {
	name: 'Noise_X',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_X") {
				this.pattern = pattern
			}
		})
	}
}
</script>
