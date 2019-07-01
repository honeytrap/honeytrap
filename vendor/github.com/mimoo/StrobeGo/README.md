# StrobeGo

[![GoDoc](https://godoc.org/github.com/mimoo/StrobeGo/strobe?status.svg)](https://godoc.org/github.com/mimoo/StrobeGo/strobe)

This repository contains an implementation of the [Strobe protocol framework](https://strobe.sourceforge.io/). See [this blogpost](https://www.cryptologie.net/article/416/the-strobe-protocol-framework/) for an explanation of what is the framework.

**The implementation of Strobe has not been thoroughly tested. Do not use this in production**.

The **Strobe** implementation is heavily based on [golang.org/x/crypto/sha3](https://godoc.org/golang.org/x/crypto/sha3), which is why some of the files have been copied in the [/strobe](/strobe) directory. You do not need to have Go's SHA-3 package to make it work.

## Install

To use it, first get Go's experimental sha3's implementation:

```
go get github.com/mimoo/StrobeGo/strobe
```

## Usage

See [godoc](https://godoc.org/github.com/mimoo/StrobeGo/strobe) for thorough documentation. Here is an example usage:

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/mimoo/StrobeGo/strobe"
)

func main() {
	s := strobe.InitStrobe("myHash", 128) // 128-bit security
	message := []byte("hello, how are you good sir?")
	s.AD(false, message) // meta=false
	fmt.Println(hex.EncodeToString(s.PRF(16))) // output length = 16
}
```

## Roadmap

* Implement test vectors of SHAKE
* Generate proper test vectors and test them with the reference implementation in python of Strobe
