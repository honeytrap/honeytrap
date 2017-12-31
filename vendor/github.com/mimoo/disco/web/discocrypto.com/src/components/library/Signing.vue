<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-wrench"></i> Signing</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p>Asymmetric cryptographic operations in libdisco are built on top of the Curve25519 primitive. Since Golang has standard libraries implementing both ed25519 (for signing) and X25519 (for key exchanges) this is what we use.</p>

		<p>Please refer to the documentation for <a href="https://godoc.org/golang.org/x/crypto/ed25519">x/crypto/ed25519</a> in order to use it.</p>

		<p>Note that since libdisco already uses these libraries, importing them in your project will not increase the size of the final binaries.</p>

	</section>

</template>
