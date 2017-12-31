<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_NX.png" alt="Noise_NX handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2>

		<p>Noise_NX is a fitting handshake pattern if:</p>
		<ul>
			<li>clients talk to several servers with no prior knowledge of their public keys</li>
			<li>clients do not authenticate themselves</li>
		</ul>
		<p>if a server's public key is known prior to starting the handshake, refer to <router-link to="/protocol/Noise_NK">Noise_NK</router-link>. If clients need to authenticate themselves, refer to <router-link to="/protocol/Noise_XX">Noise_XX</router-link>.</p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

				<p>Noise_NX requires the server to authenticate itself as part of the handshake. For this to work:</p>
				<ul>
					<li>the server needs to have its public static key signed by an authoritative key pair</li>
					<li>the client needs to be aware of the authoritative public key</li>
				</ul>

				<p>Both of these requirements can be achieved using libdisco's <router-link to="/protocol/Keys">key helper functions</router-link>.</p>

				<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_NX">here</a>. The root signing key process is illustrated <a href="https://github.com/mimoo/disco/blob/master/libdisco/examples/RootSigningKeys/root.go">here</a>.</p>

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

		<p>As part of Noise_NX, the server needs to be configured with a static public key, as well as a signature over that key.</p>

		<p>To keep things simple, libdisco provides <router-link to="/protocol/Keys">utility functions</router-link> to avoid X.509 certificates. These utility functions simply allow you to construct a signature over a public static key and to verify such signatures:</p>

		<pre><code>// CreateStaticPublicKeyProof helps in creating a signature over the peer's static public key
// for that, it needs the private part of a signing root key pair that is trusted by the client.
proof := CreateStaticPublicKeyProof(rootPrivateKey, peerPublicKey)
</code></pre>

		<article class="message is-info">
		  <div class="message-header">
		    <p>Public Key Infrastructure</p>
		  </div>
		  <div class="message-body">
		    Similarly to our typical browser ↔ HTTPS webserver scenario, a proof could also be an X.509 certificate containing the <code>serverKeyPair</code> as well as a signature of the certificate from a certificate authority's public key.  If such a complex public key infrastructure is required, you can construct the <code>PublicKeyVerifier</code> and <code>StaticPublicKeyProof</code> yourself to verify a certificate's signature and accept certificates as proofs. See the <router-link to="/protocol/Overview">Configuration section of the protocol Overview</router-link>.
		  </div>
		</article>

		<p>Once the proof has been computed, it can be passed to the server which will be able to configure itself for a Noise_NX setup:</p>

		<pre><code>serverConfig := libdisco.Config{
  HandshakePattern:     libdisco.Noise_NX,
  KeyPair:              serverKeyPair,
  StaticPublicKeyProof: proof,
}

// listen on port 6666
listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
if err != nil {
	// cannot setup a listener on localhost
}
addr := listener.Addr().String()
fmt.Println("listening on:", addr)</code></pre>

		<h3>client:</h3>

		<p>the client needs to be configured with a function capable of acting on the static public key the server will send (as part of the handshake). Without this, there are no guarantees that the static public key the server sends is "legit".</p>

		<pre><code>clientConfig := libdisco.Config{
  HandshakePattern:  libdisco.Noise_NX,
  PublicKeyVerifier: verifier,
}</code></pre>

		<p>Again, libdisco provides utility functions to create a useful <code>PublicKeyVerifier</code> callback:</p>

		<pre><code>// CreatePublicKeyVerifier helps in creating a callback function that will verify a signature
// for this it needs the public part of the signing root public key that we trust.
verifier := CreatePublicKeyVerifier(rootPublicKey)</code></pre>

		<p>Finally the full example for a client:</p>

		<pre><code>// create a verifier for when we will receive the server's public key
verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

// configure the Disco connection
clientConfig := libdisco.Config{
	HandshakePattern:  libdisco.Noise_NX,
	PublicKeyVerifier: verifier,
}

// Dial the port 6666 of localhost
client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
if err != nil {
	// client can't connect to server
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
	name: 'Noise_NX',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_NX") {
				this.pattern = pattern
			}
		})
	}
}
</script>
