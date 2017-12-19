<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-exchange"></i> {{pattern.name}} <span class="tag" v-for="tag in pattern.tags">{{tag}}</span></h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<img src="./assets/Noise_NNpsk2.png" alt="Noise_NNpsk2 handshake">

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 

		<p>If you are only dealing with one client and one server, and can manually hardcode a common 32-byte random value on each peer, you probably do not need to deal with public keys and should use this handshake pattern.</p>

		<article class="message is-danger">
		  <div class="message-header">
		    <p>Security Consideration</p>
		  </div>
		  <div class="message-body">
		    The same amount of care as with private keys should be taken to store and protect shared secrets. 
		  </div>
		</article>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>

		<p>The shared secret can be generated using a cryptographically random number generator (see <router-link to="/library/RandomNumers">Generating Random Numbers</router-link>). It then needs to be manually hardcoded on each peer's.</p>

		<p>You can play with the full example <a href="https://github.com/mimoo/disco/tree/master/libdisco/examples/Noise_NNpsk2">here</a>.</p>

		<h3>server:</h3>

		<pre><code>// configuring the Disco connection
serverConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_NNpsk2,
	// your 32-byte shared secret
	PreSharedKey: sharedSecret, 
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

		<pre><code>// configure the Disco connection
clientConfig := libdisco.Config{
	HandshakePattern: libdisco.Noise_NNpsk2,
	// your 32-byte shared secret
	PreSharedKey: sharedSecret, 
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
	name: 'Noise_NNpsk2',
	data () {
		return {
			pattern: {}
		}
	},
	beforeMount () {
		patterns.forEach( (pattern) => {
			if(pattern.name == "Noise_NNpsk2") {
				this.pattern = pattern
			}
		})
	}
}
</script>