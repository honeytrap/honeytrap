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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"

	"github.com/honeytrap/honeytrap/storage"
)

func Storage() (*smtpStorage, error) {
	if s, err := storage.Namespace("smtp"); err == nil {
		return &smtpStorage{
			s,
		}, nil
	} else {
		return nil, err
	}
}

type smtpStorage struct {
	storage.Storage
}

//Returns a TLS Certificate or nil on error
func (s *smtpStorage) Certificate() *tls.Certificate {

	priv_b, err := s.Get("private-key")
	if err != nil {
		priv_b, err = generateKey()
		if err != nil {
			return nil
		}
		if err := s.Set("private-key", priv_b); err != nil {
			log.Errorf("Could not persist private-key: %s", err.Error())
		}
	}

	cert, err := generateCert(priv_b)
	if err != nil {
		log.Errorf("Could not generate a Certificate. %s", err)
		return nil
	}

	return cert
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
