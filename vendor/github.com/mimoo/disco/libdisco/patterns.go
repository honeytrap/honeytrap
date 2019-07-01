package libdisco

//
// Handshake Patterns
//

type noiseHandshakeType int8

const (
	// Noise_N is a one-way pattern where a client can send
	// data to a server with a known static key. The server
	// can only receive data and cannot reply back.
	Noise_N noiseHandshakeType = iota

	// Noise_K is a one-way pattern where a client can send
	// data to a server with a known static key. The server
	// can only receive data and cannot reply back. The server
	// authenticates the client via a known key.
	Noise_K

	// Noise_X is a one-way pattern where a client can send
	// data to a server with a known static key. The server
	// can only receive data and cannot reply back. The server
	// authenticates the client via a key transmitted as part
	// of the handshake.
	Noise_X

	// Noise_KK is a pattern where both the client static key and the
	// server static key are known.
	Noise_KK

	// Noise_NX is a "HTTPS"-like pattern where the client is
	// not authenticated, and the static public key of the server
	// is transmitted during the handshake. It is the responsability of the client to validate the received key properly.
	Noise_NX

	// Noise_NK is a "Public Key Pinning"-like pattern where the client
	// is not authenticated, and the static public key of the server
	// is already known.
	Noise_NK

	// Noise_XX is a pattern where both static keys are transmitted.
	// It is the responsability of the server and of the client to
	// validate the received keys properly.
	Noise_XX

	// Not documented
	Noise_KX
	Noise_XK
	Noise_IK
	Noise_IX
	Noise_NNpsk2

	// Not implemented
	Noise_NN
	Noise_KN
	Noise_XN
	Noise_IN
)

type token uint8

const (
	token_e token = iota
	token_s
	token_es
	token_se
	token_ss
	token_ee
	token_psk
)

type messagePattern []token

type handshakePattern struct {
	name               string
	preMessagePatterns []messagePattern
	messagePatterns    []messagePattern
}

// TODO: add more patterns
var patterns = map[noiseHandshakeType]handshakePattern{

	// 7.2. One-way patterns

	Noise_N: handshakePattern{
		name: "N",
		preMessagePatterns: []messagePattern{
			messagePattern{},        // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es}, // →
		},
	},

	/*
		K(s, rs):
		  -> s
		  <- s
		  ...
		  -> e, es, ss
	*/
	Noise_K: handshakePattern{
		name: "K",
		preMessagePatterns: []messagePattern{
			messagePattern{token_s}, // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es, token_ss}, // →
		},
	},
	/*
		X(s, rs):
		 <- s
		 ...
		 -> e, es, s, ss
	*/
	Noise_X: handshakePattern{
		name: "X",
		preMessagePatterns: []messagePattern{
			messagePattern{},        // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es, token_s, token_ss}, // →
		},
	},
	//
	// 7.3. Interactive patterns
	//
	Noise_KK: handshakePattern{
		name: "KK",
		preMessagePatterns: []messagePattern{
			messagePattern{token_s}, // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es, token_ss}, // →
			messagePattern{token_e, token_ee, token_se}, // ←
		},
	},

	Noise_NX: handshakePattern{
		name: "NX",
		preMessagePatterns: []messagePattern{
			messagePattern{}, // →
			messagePattern{}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e},                              // →
			messagePattern{token_e, token_ee, token_s, token_es}, // ←
		},
	},

	Noise_NK: handshakePattern{
		name: "NK",
		preMessagePatterns: []messagePattern{
			messagePattern{},        // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es}, // →
			messagePattern{token_e, token_ee}, // ←
		},
	},

	Noise_XX: handshakePattern{
		name: "XX",
		preMessagePatterns: []messagePattern{
			messagePattern{}, // →
			messagePattern{}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e},                              // →
			messagePattern{token_e, token_ee, token_s, token_es}, // ←
			messagePattern{token_s, token_se},                    // →
		},
	},

	/*
			KX(s, rs):
		      -> s
		      ...
		      -> e
		      <- e, ee, se, s, es
	*/
	Noise_KX: handshakePattern{
		name: "KX",
		preMessagePatterns: []messagePattern{
			messagePattern{token_s}, // →
			messagePattern{},        // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e},                                        // →
			messagePattern{token_e, token_ee, token_se, token_s, token_es}, // ←
		},
	},
	/*
			XK(s, rs):
		  <- s
		  ...
		  -> e, es
		  <- e, ee
		  -> s, se
	*/
	Noise_XK: handshakePattern{
		name: "XK",
		preMessagePatterns: []messagePattern{
			messagePattern{},        // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es}, // →
			messagePattern{token_e, token_ee}, // ←
			messagePattern{token_s, token_se}, // →
		},
	},
	/*
		IK(s, rs):
		<- s
		...
		-> e, es, s, ss
		<- e, ee, se
	*/
	Noise_IK: handshakePattern{
		name: "IK",
		preMessagePatterns: []messagePattern{
			messagePattern{},        // →
			messagePattern{token_s}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_es, token_s, token_ss}, // →
			messagePattern{token_e, token_ee, token_se},          // ←
		},
	},
	/*
		IX(s, rs):
		 -> e, s
		 <- e, ee, se, s, es
	*/
	Noise_IX: handshakePattern{
		name: "IX",
		preMessagePatterns: []messagePattern{
			messagePattern{}, // →
			messagePattern{}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e, token_s},                               // →
			messagePattern{token_e, token_ee, token_se, token_s, token_es}, // ←
		},
	},

	/*
		NNpsk2():
		  -> e
		  <- e, ee, psk
	*/
	Noise_NNpsk2: handshakePattern{
		name: "NNpsk2",
		preMessagePatterns: []messagePattern{
			messagePattern{}, // →
			messagePattern{}, // ←
		},
		messagePatterns: []messagePattern{
			messagePattern{token_e},                      // →
			messagePattern{token_e, token_ee, token_psk}, // ←
		},
	},
}
