<template>
	<section class="content">
		
		<h1 class="title"><i class="fa fa-wrench"></i> Encryption</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p>To encrypt, simply use the <code>Encrypt</code> function with a key of at minimum 128 bits (16 bytes)</p>

		<pre><code>key, _ := hex.DecodeString("eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75")
plaintext := []byte("hello, how are you?")

ciphertext := libdisco.Encrypt(key, plaintext)</code></pre>

		<p>To decrypt, use the corresponding <code>Decrypt</code> function with the same key.</p>

		<pre><code>key, _ := hex.DecodeString("eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75")

decrypted, err := libdisco.Decrypt(key, ciphertext)</code></pre>

	</section>

</template>
