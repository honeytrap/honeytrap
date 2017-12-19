<template>
	<section class="content">
	    <h1 class="title"><i class="fa fa-exchange"></i> Disco Keys</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> The Different Keys</h2>

		<p>Disco makes use of several key pairs:</p>

		<ul>
<li><strong>Ephemeral keys</strong>. They are freshly generated, behind the curtains, for each new client ↔ server connection. For this reason you do not have to worry about these and you can just ignore the fact that they exist.</li>
<li><strong>Static keys</strong>. Clients and servers can be required to have their own long-term static key in case the handshake pattern in use requires them to authenticate themselves. If this is the case, they will need to generate a static key pair once (and only once) and use it for authenticating themselves during handshakes.</li>
<li><strong>Root signing keys</strong>. These are authoritative keys that sign the static keys in setups where static public keys are not known in advance to the other peer. These are similar to the internet's certificate authorities.</li>
		</ul>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Generation and Storage</h2>

<p>
		<strong>Static keys</strong> can be generated via a call to <a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateKeypair"><code>GenerateKeypair(nil)</code></a>. The package also provides some file utility functions:
</p>
		<ul>
			<li><a href="https://godoc.org/github.com/mimoo/disco/libdisco#KeyPair.ExportPublicKey"><code>KeyPair.ExportPublicKey()</code></a> retrieves the public part of a static key pair.</li>
			<li><a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateAndSaveDiscoKeyPair"><code>GenerateAndSaveDiscoKeyPair()</code></a> creates and saves a static key pair on disk.</li>
			<li><a href="https://godoc.org/github.com/mimoo/disco/libdisco#LoadDiscoKeyPair"><code>LoadDiscoKeyPair(discoPrivateKeyPairFile()</code></a> loads a static key pair from such a file.</li>
		</ul>
<p>
		<strong>Root signing keys</strong> can be generated and saved on disk directly via the <a href="https://godoc.org/github.com/mimoo/disco/libdisco#GenerateAndSaveDiscoRootKeyPair"><code>GenerateAndSaveDiscoRootKeyPair()</code></a> function. As different peers might need different parts, the private and public parts of the key pair will be saved in different files. To retrieve them you can use the <a href="https://godoc.org/github.com/mimoo/disco/libdisco#LoadDiscoRootPublicKey"><code>LoadDiscoRootPublicKey()</code></a> and <a href="https://godoc.org/github.com/mimoo/disco/libdisco#LoadDiscoRootPrivateKey"><code>LoadDiscoRootPrivateKey()</code></a> functions.
</p>


	<article class="message is-danger">
	  <div class="message-header">
	    <p>Storing keys</p>
	  </div>
	  <div class="message-body">
	    Private part of key pairs should be stored in secure places. Such sensitive information should not be checked in version control systems like git or svn. Instead, they can be stored outside of the program's repository, or passed as environment variables.
	  </div>
	</article>


		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Configuration of Peers</h2>

		<p>Imagine a handshake pattern like <router-link to="/protocol/Noise_NX">Noise_NX</router-link> where only the server sends its static public key along with a proof.</p>

		<p>1. create the root signing key:</p>

		<pre><code>if err := libdisco.GenerateAndSaveDiscoRootKeyPair("./discoRootPrivateKey", "./discoRootPublicKey"); err != nil {
  panic("didn't work")
}</code></pre>

	<article class="message is-danger">
	  <div class="message-header">
	    <p>Storing root keys</p>
	  </div>
	  <div class="message-body">
	    Note that in this example the private key of the root signing key pair is stored on disk next to the application. Extra care should be taken so that this private key stays accessible only to the root signing program, and inaccessible from the peers (clients and servers).
	  </div>
	</article>

		<p>2. sign the server's static public key:</p>

		<pre><code>// we load the private part of the root signing key
rootPrivateKey, err := libdisco.LoadDiscoRootPrivateKey("./discoRootPrivateKey")
if err != nil {
  panic("couldn't load the root signing private key")
}
// we compute our proof over the server's public static key
proof := libdisco.CreateStaticPublicKeyProof(rootPrivateKey, serverKeyPair.PublicKey[:])</code></pre>

		<p>3. configure the server to send its static public key along with its proof:</p>

		<pre><code>serverConfig := libdisco.Config{
  HandshakePattern:     libdisco.Noise_NX,
  KeyPair:              serverKeyPair,
  StaticPublicKeyProof: proof,
}</code></pre>

		<p>4. once the <code>discoRootPublicKey</code> file has been passed to the client, we can configure it:</p>

		<pre><code>// we load the public part of the root signing key
rootPublicKey, err := LoadDiscoRootPublicKey("./discoRootPublicKey")
if err != nil {
  panic("didn't work")
}

// we create our verifier
someCallbackFunction := CreatePublicKeyVerifier(rootPublicKey)

// we configure the client
clientConfig := libdisco.Config{
  HandshakePattern:  libdisco.Noise_NK,
  PublicKeyVerifier: someCallbackFunction,
}
</code></pre>

	<p>5. And that's it!</p>

	<p>For more example please check each handshake pattern individually on the examples on the <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples" target="_blank">source code repository</a></p>

	</section>

</template>

<script>
import patterns from '@/assets/patterns.json';

export default {
	name: 'Noise_KK',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_KK") {
				this.pattern = pattern
			}
		})
	}
}
</script>
