<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_NK.png" alt="Noise_NK handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 

		<p>Noise_NK is a relevant handshake pattern if your clients already know a server's static public key. This pattern does not authenticate the client nor does it rely on an external root signing key.</p>

		<p>If the client needs to be authenticated, refer to <router-link to="/protocol/Noise_KK">Noise_KK</router-link>. For clients who do not know the server's static public key in advance, refer to <router-link to="/protocol/Noise_NX">Noise_NX</router-link>. </p>

		<!-- if the clients need to be authenticated, there should also be <router-link to="/protocol/Noise_XK">Noise_XK</router-link> -->

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>To understand how to generate the server's static key pair and export its public key part, refer to <router-link to="/protocol/Keys">libdisco's documentation on keys</router-link>.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_NK">here</a>.</p>


		<h3>server:</h3>

		<pre><code>serverConfig := libdisco.Config{
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
fmt.Println("listening on:", addr)</code></pre>

		<h3>client:</h3>

		<pre><code>clientConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_NK,
	// the server's static public key.
	RemoteKey:        serverPublicKey,
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
	name: 'Noise_NK',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_NK") {
				this.pattern = pattern
			}
		})
	}
}
</script>