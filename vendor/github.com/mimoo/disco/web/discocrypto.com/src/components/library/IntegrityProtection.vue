<template>
	<section class="content">
		<h1 class="title"><i class="fa fa-wrench"></i> Integrity Protection</h1>

		<h2><i class="fa fa-caret-right" aria-hidden="true"></i> Description</h2>

		<p>The <code>ProtectIntegrity</code> and <code>VerifyIntegrity</code> functions can be used to make sure that a message has not been tampered with. Note that this creates a payload containing your message in clear. This is not encryption!</p>

		<img src="./assets/integrity.png" alt="integrity protection">

		<p>to create your payload with a message:</p>

		<pre><code>key, _ := hex.DecodeString("eda8506c1fb0bbcc3f62626fef074bbf2d09a8c7c608f3fa1482c9a625d00f75")

message := []byte("hoy, how are you?")

payload := ProtectIntegrity(key, message)</code></pre>

		<p>to extract your payload use the function <code>VerifyIntegrity</code>. It will return an error if the payload has been modified since its creation, otherwise it will return the original message:</p>

		<pre><code>retrievedMessage, err := VerifyIntegrity(key, plaintextAndTag)

if err != nil {
	// the message was tampered with
}</code></pre>

	</section>

</template>
