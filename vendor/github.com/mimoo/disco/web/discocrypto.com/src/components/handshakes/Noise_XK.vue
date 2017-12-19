<template>
	<section class="content">
		<h1 class="title is-1">Noise_NNpsk2</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p v-html="pattern.description"></p>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Use cases</h2> 



		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Example of configuration</h2>



		<article class="message is-danger">
		  <div class="message-header">
		    <p>Security Consideration</p>
		  </div>
		  <div class="message-body">
		    Note that in this example the private part of the root signing key is loaded on both peers. In practice, this and the computation of the proof should be done on a different machine.
		  </div>
		</article>

		<h3>server:</h3>

		<pre><code></code></pre>

		<h3>client:</h3>

		<pre><code></code></pre>

		<h3>Security Considerations</h3>

		<ul>
			<li>the pre-shared key must be safely stored on both endpoints.</li>
			<li>you usually do not want to share a single pre-shared key over many endpoints, better ways to do this usually involve a PKI. See <router-link to="/handshakes/Noise_XX">Noise_XX</router-link>.</li>
		</ul>

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