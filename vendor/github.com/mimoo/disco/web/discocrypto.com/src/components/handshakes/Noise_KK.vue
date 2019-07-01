<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_KK.png" alt="Noise_KK handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 

		<p>If your protocol involves several peers and both sides of a connection have a way to know the other side's public static key prior to starting the handshake, Noise_KK is a good fit.</p>

		<p>A simple illustration would be a secure messaging application where the peer's public key can be retrieved in advance via a trusted third party (key server). If the third party cannot be trusted, public keys can be further verified out-of-band (usually via a fingerprint mixing both public keys, see <a href="https://signal.org/blog/safety-number-updates/" target="_blank">Signal's safety numbers</a>).</p>

		<p>If your protocol involves a single client and a single server, refer to <router-link to="/protocol/Noise_NNpsk2">Noise_NNpsk2</router-link>.</p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>To understand how to generate both peer's key pair, refer to <router-link to="/protocol/Keys">libdisco's documentation on keys</router-link>.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_KK">here</a>.</p>

		<h3>server:</h3>

		<pre><code>serverConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_KK,
	KeyPair:          serverKeyPair,
	// the public static key of the client
	RemoteKey: clientPublicKey
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

		<pre><code>clientConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_KK,
	KeyPair:          clientKeyPair,
	// the public static key of the server
	RemoteKey: serverPublicKey
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