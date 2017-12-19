<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_N.png" alt="Noise_N handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 

		<p>This handshake pattern is useful for clients that always talk to a single server. In addition, since it is a one-way pattern, the server never talks back to them. The server also doesn't require the client to authenticate itself.</p>

		<p>If client authentication is needed, refer to <router-link to="/protocol/Noise_K">Noise_K</router-link> or <router-link to="/protocol/Noise_X">Noise_X</router-link>.</p>
	
		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>The client needs to have prior knowledge to the server's public static key. In this example we just pass it as an <code>stdin</code> argument to the client's CLI, but in practice it should be hardcoded.</p>

		<p>In addition, every time the server is ran it is generating a new static key pair. In practice this should only be done once, possibly using the <router-link to="protocol/Keys">key helper functions</router-link> that libdisco provides.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_N">here</a>.</p>

		<h3>server:</h3>

		<pre><code>// generating the server key pair
serverKeyPair := libdisco.GenerateKeypair(nil)

// configuring the Disco connection with a Noise_N handshake
// in which the client already knows the server's static public key
serverConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_N,
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
fmt.Println("server's public key:", serverKeyPair.ExportPublicKey())</code></pre>

		<h3>client:</h3>

		<pre><code>// retrieve the server's public key from an argument
serverPubKey := os.Args[1]
serverKey, _ := hex.DecodeString(serverPubKey)

// configure the Disco connection with Noise_N
// meaning the client knows the server's static public key (retrieved from the CLI)
clientConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_N,
	RemoteKey:        serverKey,
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
	name: 'Noise_N',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_N") {
				this.pattern = pattern
			}
		})
	}
}
</script>