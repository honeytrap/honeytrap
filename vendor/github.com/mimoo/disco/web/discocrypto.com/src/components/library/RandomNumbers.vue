<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-wrench"></i> Random Numbers</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p>libdisco uses golang's standard library <a href="https://golang.org/pkg/crypto/rand/">crypto/rand</a> to generate random numbers.</p>

		<p>It is possible to use Strobe to generate random numbers in userland, but it is not advised at forking or cloning a VM snapshot will end up making different programs produce the same random numbers.</p>

	</section>
</template>
