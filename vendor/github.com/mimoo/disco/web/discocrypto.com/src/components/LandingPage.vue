<template>
  <section class="content">


    <section class="hero is-info is-bold">
      <div class="hero-body">
        <h1 class="title">
          <strong>libdisco</strong>
        </h1>
        <h2 class="subtitle">
          libdisco is a <strong>modern plug-and-play secure protocol</strong> and a <strong>cryptographic library</strong> in <strong>Golang</strong>. It offers different ways of encrypting communications, as well as different cryptographic primitives for all of an application's needs.
        </h2>
      </div>
    </section>

    <article class="message is-danger" style="margin:20px 0">
      <div class="message-header">
        <p>Warning</p>
      </div>
      <div class="message-body">
        libdisco is <strong>experimental</strong>. It has not been thoroughly reviewed and relies on an unstable specification. It should not be used in production.
      </div>
    </article>

    <p>
      <strong>libdisco</strong> is a library built by merging the <a href="https://noiseprotocol.org">Noise protocol framework</a> and the <a href="https://strobe.sourceforge.io">Strobe protocol framework</a>. This means that it supports a subset of Noise's handshakes while offering the cryptographic primitive Strobe has to offer. In other words, you can use libdisco to securely connect peers together, or to do basic cryptographic operations like hashing or encrypting.</p>

          <p>
      <ul class="main_list">
        <li>The <span>secure protocol parts</span> are based on the <a href="/disco.html"><i class="fa fa-file-text-o" aria-hidden="true"></i> Disco specification</a> which extends the <strong>Noise protocol framework</strong>.</li>
        <li>The <span>symmetric cryptographic primitives</span> are all based on the <strong>Strobe protocol framework</strong> which only relies on the <strong>SHA-3</strong> permutation (called <strong>keccak-f</strong>).</li>
        <li>The <span>asymmetric cryptographic primitives</span> (<strong>X25519</strong> and <strong>ed25519</strong>) are based on the <a href="https://golang.org/">golang</a> standard library.</li>
      </ul>
    </p>

    <hr>

    <p>To set it up, follow <a href="https://golang.org/pkg/net/" target="_blank">Golang's net/conn</a> standard way of setting up a server with a libdisco config:</p>

    <pre><code>serverConfig := libdisco.Config{
  HandshakePattern: libdisco.Noise_NK,
  KeyPair:          serverKeyPair,
}
listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
server, err := listener.Accept()</code></pre>

    <p>and the standard way of setting up a client with a libdisco config:</p>

    <pre><code>clientConfig := libdisco.Config{
  HandshakePattern: libdisco.Noise_NK,
  RemoteKey:        serverKey,
}
client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
</code></pre>

    <p>it's that simple! Check out the <router-link to="/get_started">get started</router-link> section for more information.</p>


<section class="hero is-warning" style="margin-bottom:40px">
  <div class="hero-body">
    <h1 class="title" style="margin-bottom:0">
      <strong>Why use libdisco?</strong>
    </h1>
      <ul style="font-size:18px">
        <li>libdisco's source code is around <strong>1000 lines of code</strong> without the cryptographic primitives. Around 2000 lines of code with the symmetric cryptographic primitive (Strobe), and 4000 lines of code with X25519 (for key exchanges). It makes it easy to audit and a pleasure to fit into tiny devices.</li>
        <li>The protocol relies on the solid <strong>Noise protocol framework</strong> which is used by many others including WhatsApp, Wireguard, the Bitcoin Lightning Network, ...</li>
        <li>The library relies only on <strong>two cryptographic primitives</strong> and nothing else: <strong>Curve25519</strong> and the <strong>SHA-3</strong> permutation.</li>
        <li>libdisco is <strong>flexible</strong>. libdisco supports many different ways of securely connecting two peers together while avoiding complex certificates or public key infrastructures (unless you want to use them).</li>
        <li>libdisco is <strong>versatile</strong>. While it allows you to securely link peers together, it is also an entire cryptographic library!</li>
      </ul>
  </div>
</section>



    <!--          If you want to know more you can read the whitepaper or watch this video -->


    <p>To make use of the <strong><i class="fa fa-exchange" aria-hidden="true"></i>
    protocol parts</strong>, you must first choose how you want to authenticate the connection. For that, it's easy! We have made a <strong>small quizz</strong> for you bellow, but if you already know what you want you can directly click on your favorite way of doing this in the menu under "protocol" (Noise_NK, Noise_XX, ...) and copy the usage examples.</p>

    <Quizz></Quizz>

    <p>To make use of the <strong><i class="fa fa-wrench" aria-hidden="true"></i>
    cryptographic library</strong>, check our <router-link to="/library/Overview">overview here</router-link> or directly access them through the menu on the left.</p>

    <hr>

    <p>To learn more about it, you can read <a href="https://www.cryptologie.net/article/432/disco/" target="_blank">this blog post</a>.</p>

    <p>If you want help, head to the <a href="https://github.com/mimoo/disco/issues"><i class="fa fa-question" aria-hidden="true"></i>
    issues on github</a>.</p>

    <p>If you want to stay tuned on what we're doing, we don't have a mailing list but we have better: <a href="https://www.reddit.com/r/discocrypto/"><i class="fa fa-envelope-o" aria-hidden="true"></i>
    a subreddit over at r/discocrypto</a>.</p>
  </section>
</template>

<style scoped>
  .title{
    border-bottom:0;
  }
</style>

<script>
  import Quizz from './Quizz'

  export default {
    name: 'LandingPage',
    components: {
      Quizz
    }
  }
</script>
