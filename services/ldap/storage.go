/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ldap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/honeytrap/honeytrap/storage"
)

func getStorage() (*ldapStorage, error) {

	s, err := storage.Namespace("ldap")
	if err != nil {
		return nil, err
	}

	return &ldapStorage{s}, nil
}

type ldapStorage struct {
	storage.Storage
}

//Certificate Returns a TLS Certificate
func (s *ldapStorage) Certificate() (*tls.Certificate, error) {

	keyname := "pemkey"
	certname := "pemcert"

	pemkey, err := s.Get(keyname)
	if err != nil {
		pemkey, err = generateKey()
		if err != nil {
			return nil, err
		}
		if err = s.Set(keyname, pemkey); err != nil {
			log.Errorf("Could not persist %s: %s", keyname, err.Error())
		}
	}

	pemcert, err := s.Get(certname)
	if err != nil {
		pemcert, err = generateCert(pemkey)
		if err != nil {
			return nil, err
		}

		log.Debug("TLS new certificate generated")

		if err = s.Set(certname, pemcert); err != nil {
			log.Errorf("Could not persist %s: %s", certname, err.Error())
		}
	}

	log.Debug("TLS Certificate loaded from store")

	tlscert, err := tls.X509KeyPair(pemcert, pemkey)
	if err != nil {
		return nil, err
	}

	return &tlscert, nil
}

//Returns a PEM encoded RSA private key
func generateKey() ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		return nil, cerr
	}

	pemdata := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return pemdata, nil
}

func generateCert(pempriv []byte) ([]byte, error) {

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
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	block, _ := pem.Decode(pempriv)

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Could not parse private key: %s", err.Error())
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, priv.Public(), priv)
	if err != nil {
		return nil, err
	}

	certpem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return certpem, nil
}
