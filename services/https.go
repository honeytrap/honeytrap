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
package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/event"
	tls "github.com/honeytrap/honeytrap/services/ja3/crypto/tls"

	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("https", HTTPS)
)

func HTTPS(options ...ServicerFunc) Servicer {
	s := &httpsService{
		httpService: httpService{
			httpServiceConfig: httpServiceConfig{
				Server: "Apache",
			},
		},
		tlsConfig: &tls.Config{},
		n:         0,
		m:         sync.Mutex{},
		cache:     map[string]*tls.Certificate{},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type httpsService struct {
	httpService

	tlsConfig *tls.Config

	c pushers.Channel

	n int64

	m sync.Mutex

	cache map[string]*tls.Certificate
}

func (s *httpsService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *httpsService) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if cert, ok := s.cache[hello.ServerName]; ok {
		return cert, nil
	}

	s.n++

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(s.n),
		Subject: pkix.Name{
			Country:            []string{""},
			Organization:       []string{""},
			OrganizationalUnit: []string{""},
		},
		Issuer: pkix.Name{
			Country:            []string{""},
			Organization:       []string{""},
			OrganizationalUnit: []string{""},
			Locality:           []string{""},
			Province:           []string{""},
			StreetAddress:      []string{""},
			PostalCode:         []string{""},
			SerialNumber:       fmt.Sprintf("%d", 0),
			CommonName:         hello.ServerName,
		},
		SignatureAlgorithm:    x509.SHA512WithRSA,
		PublicKeyAlgorithm:    x509.ECDSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		IsCA:        false,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 4096)

	pub := &priv.PublicKey

	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  priv,
	}

	s.cache[hello.ServerName] = cert

	return cert, nil
}

func (s *httpsService) Handle(ctx context.Context, conn net.Conn) error {
	ja3Digest := ""
	serverName := ""

	tlsConn := tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{},
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			ja3Digest = hello.JA3Digest()
			serverName = hello.ServerName
			return s.getCertificate(hello)
		},
	})

	if err := tlsConn.Handshake(); err != nil {
		s.c.Send(event.New(
			EventOptions,
			event.Category("https"),
			event.Type("handshake-failed"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("https.ja3-digest", ja3Digest),
			event.Custom("https.server-name", serverName),
		))

		return err
	}

	return s.httpService.Handle(ctx, event.WithConn(
		tlsConn,
		event.Custom("https.ja3-digest", ja3Digest),
		event.Custom("https.server-name", serverName),
	))
}
