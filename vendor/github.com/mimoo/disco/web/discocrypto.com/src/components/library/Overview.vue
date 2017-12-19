<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-wrench"></i> Library Overview</h1>

		<p>
			<strong>libdisco</strong> is based on the <a href="https://strobe.sourceforge.io/" target="_blank">Strobe protocol framework</a> and on <a href="https://cr.yp.to/ecdh/curve25519-20060209.pdf" target="_blank">Curve25519</a>. In other words, libdisco supports a diverse range of symmetric and asymmetric cryptographic.
		</p>
		<p>
			The different cryptographic operations available are documented in the <i class="fa fa-bars" aria-hidden="true"></i> menu on the left.
		</p>
		<p>
			More complex symmetric cryptographic primitives can be built on top of Strobe. For this, use the <a href="https://github.com/mimoo/StrobeGo" target="_blank">strobe</a> package. Note that it is already imported as part of libdisco, so re-importing it and using it won't increase the size of your applications' binaries.
		</p>
	</section>
</template>

<script>
  export default {
    name: 'libraryOverview'
  }
</script>