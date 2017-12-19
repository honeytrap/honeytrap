<template>
	<section class="content">
	    <h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_K.png" alt="Noise_K handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 

		<p>Like any one-way pattern. If the server never needs to (or cannot) reply to the client, Noise_K might be a good fit. In addition, this handshake authenticates both side of the connection via <strong>public-key pinning</strong>. This means that both sides need to know in advance each other's static public key.</p>

		<p>If only one side needs to be authenticated, refer to <router-link to="/protocol/Noise_N">Noise_N</router-link>.</p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>To configure both the client and the server, they need to have each other's public static key. In this example we just pass these as <code>stdin</code> argument to each CLI, but in practice these should be hardcoded.</p>

		<p>In addition, every time the client and server are ran,  they are generating new static key pairs. In practice this should only be done once, possibly using the <router-link to="protocol/Keys">key helper functions</router-link> that libdisco provides.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_K">here</a>.</p>

		<h3>server:</h3>

		<pre><code>// generating the server key pair
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
fmt.Println("listening on:", listener.Addr().String())</code></pre>

		<h3>client:</h3>

		<pre><code>// generating the client key pair
clientKeyPair := libdisco.GenerateKeypair(nil)
fmt.Println("client's public key:", clientKeyPair.ExportPublicKey())

// configure the Disco connection
clientConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_K,
	KeyPair:          clientKeyPair,
}

// retrieve the server's public key from an argument
fmt.Println("please enter the server's public key in hexadecimal")
scanner := bufio.NewScanner(os.Stdin)
scanner.Scan()
serverKey, _ := hex.DecodeString(scanner.Text())
clientConfig.RemoteKey = serverKey

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
	name: 'Noise_K',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_K") {
				this.pattern = pattern
			}
		})
	}
}
</script>