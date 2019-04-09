// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

func makePrivateKey(data []byte) *privateKey {
	privblk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   data,
	}

	privateBytes := pem.EncodeToMemory(&privblk)

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil
	}

	return &privateKey{private}
}

// privateKey holds the ssh.Signer instance to unsign received data.
type privateKey struct {
	ssh.Signer
}

// UnmarshalText unmarshalls the giving text as the Signers data.
func (t *privateKey) UnmarshalText(data []byte) (err error) {
	private, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return err
	}

	*t = privateKey{private}
	return err
}

func generateKey() ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		return nil, cerr
	}

	data := x509.MarshalPKCS1PrivateKey(priv)
	return data, nil
}
