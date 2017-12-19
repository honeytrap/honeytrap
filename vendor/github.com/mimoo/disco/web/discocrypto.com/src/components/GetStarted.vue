<template>
	<section class="content">

	    <h1 class="title"><i class="fa fa-book"></i> Get Started</h1>

		<p>General documentation for the protocol parts of libdisco can be found on <a href="https://godoc.org/github.com/mimoo/disco/libdisco"><img src="https://godoc.org/github.com/mimoo/disco/libdisco?status.svg" alt="GoDoc" style="vertical-align: text-top;"></a></p>

		<p>Usages help can be found through-out this website. There are two important parts to libdisco:</p>

		<ul>
			<li><strong>the protocol</strong>. libdisco allows you to setup a secure channel between two endpoints to encrypt their communication. More information is available in the <router-link to="/protocol/Overview">protocol</router-link> part of this documentation.</li>
			<li><strong>the library</strong>. libdisco exports several cryptographic functions to hash, derive keys, encrypt, authenticate, ... usage example are available in the <router-link to="/library/Overview">cryptographic library</router-link> part of this documentation.</li>
		</ul>

		<p>To install the package simply use <code>go get</code>:</p>

		<pre><code>
go get github.com/mimoo/disco/libdisco
		</code></pre>

		<p>
		and import it in your application:
		</p>

		<pre><code>package main

import(
  "github.com/mimoo/disco/libdisco"
)</code></pre>
	</section>
</template>