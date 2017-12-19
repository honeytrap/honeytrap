<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-wrench"></i> Deriving Keys</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p>libdisco provides functions to "spread" a single key into several keys. The input key must be random and larger or equal to 128 bits (16 bytes) for security reasons. The following snippet takes a 16-byte key and derives two 32-byte keys:</p>

		<pre><code>input, _ := hex.DecodeString("2825a87129f065748121cdf6c1dcf8b7")
keys := libdisco.DeriveKeys(input, 64))
key1 := keys[:32]
key2 := keys[32:]</code></pre>

	</section>

</template>
