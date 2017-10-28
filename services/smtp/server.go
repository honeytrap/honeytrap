/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package smtp

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

type Handler interface {
	Serve(msg Message) error
}

type HandlerFunc func(msg Message) error

type ServeMux struct {
	m  []HandlerFunc
	mu sync.RWMutex
}

func (mux *ServeMux) HandleFunc(handler func(msg Message) error) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.m = append(mux.m, handler)
}

func HandleFunc(handler func(msg Message) error) *ServeMux {
	DefaultServeMux.HandleFunc(handler)
	return DefaultServeMux
}

var DefaultServeMux = NewServeMux()

func NewServeMux() *ServeMux { return &ServeMux{m: make([]HandlerFunc, 0)} }

func (s *ServeMux) Serve(msg Message) error {
	for _, h := range s.m {
		if err := h(msg); err != nil {
			return err
		}
	}
	return nil
}

type Server struct {
	Banner string

	Handler Handler

	tlsConf *tls.Config
}

func (s *Server) NewConn(rwc net.Conn) (c *conn, err error) {
	c = &conn{
		server: s,
		rwc:    rwc,
		i:      0,
	}

	c.msg = c.newMessage()
	return c, nil
}

func (s *Server) tlsConfig(cert *tls.Certificate) {

	s.tlsConf = &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
	}
}

func generateCert(priv_b []byte) (*tls.Certificate, error) {
	log.Debug("START generateCert()")

	snLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, snLimit)

	if err != nil {
		log.Debug("Could not generate certificate serial number")
	}

	ca := &x509.Certificate{
		SerialNumber: sn,
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
			//CommonName:         s.Banner,
		},
		SignatureAlgorithm: x509.SHA512WithRSA,
		//PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := x509.ParsePKCS1PrivateKey(priv_b)
	if err != nil {
		log.Errorf("Could not parse private key: %s", err.Error())
		return nil, err
	}

	pub := priv.Public()

	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}, nil
}

type serverHandler struct {
	srv *Server
}

func (sh serverHandler) Serve(msg Message) {
	handler := sh.srv.Handler
	if handler == nil {
		handler = DefaultServeMux
	}

	handler.Serve(msg)
}
