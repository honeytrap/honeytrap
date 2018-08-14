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
package mongodb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

func fromBase64(src []byte) []byte {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	l, _ := base64.StdEncoding.Decode(dst, src)
	return dst[:l]
}
func toBase64(src []byte) []byte {
	out := base64.StdEncoding.EncodeToString(src)
	return []byte(out)
}

func createServerSignature(serverKey, authMsg []byte) []byte {
	mac := hmac.New(sha1.New, serverKey)
	mac.Write([]byte(authMsg))
	serverSign := mac.Sum(nil)

	serverSignature := make([]byte, base64.StdEncoding.EncodedLen(len(serverSign)))
	base64.StdEncoding.Encode(serverSignature, serverSign)
	return serverSignature
}

func createSaltPassword(salt []byte, iterCount int, pass string) []byte {
	mac := hmac.New(sha1.New, []byte(pass))
	mac.Write(salt)
	mac.Write([]byte{0, 0, 0, 1})
	ui := mac.Sum(nil)
	hi := make([]byte, len(ui))
	copy(hi, ui)
	for i := 1; i < iterCount; i++ {
		mac.Reset()
		mac.Write(ui)
		mac.Sum(ui[:0])
		for j, b := range ui {
			hi[j] ^= b
		}
	}
	return hi
}

func createClientSignature(storedKey, authMsg []byte) []byte {
	mac := hmac.New(sha1.New, storedKey)
	mac.Write(authMsg)
	return mac.Sum(nil)
}

func createAuthMsg(username, clnonce, combinedNonce, salt string, itercount int) []byte {
	return []byte("n=" + username + ",r=" + clnonce + ",r=" + combinedNonce + ",s=" + salt + ",i=" + string(itercount) + ",c=biws,r=" + combinedNonce)
}

func createHashKey(clientKey []byte) []byte {
	hash := sha1.New()
	hash.Write(clientKey)
	return hash.Sum(nil)
}

func createHmacKey(key, message []byte) []byte {
	hmac := hmac.New(sha1.New, key)
	hmac.Write(message)
	return hmac.Sum(nil)
}

func xor(clientSignature, clientProof []byte) []byte {
	if len(clientSignature) != len(clientProof) {
		fmt.Println("Warning: xor lengths are differing...", clientSignature, clientProof)
	}
	n := len(clientSignature)
	if len(clientProof) < n {
		n = len(clientProof)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = clientSignature[i] ^ clientProof[i]
	}

	return createHashKey(out)
}

func checkClientProof(clientSignature, clientProof, storedKey []byte) bool {
	temp := xor(clientSignature, clientProof)
	return bytes.Equal(temp, storedKey)
}

func (s *mongodbService) scram() ([]byte, bool) {
	itercount := s.itercounts
	saltedPass := createSaltPassword([]byte(s.Client.salt), itercount, s.Client.password)
	authMsg := createAuthMsg(s.Client.username, s.Client.clNonce, s.Client.cbNonce, s.Client.salt, itercount)
	clientKey := createHmacKey(saltedPass, []byte("Client Key"))
	serverKey := createHmacKey(saltedPass, []byte("Server Key"))
	storedKey := createHashKey(clientKey)
	clientProof := fromBase64([]byte(s.Client.clProof))
	clientSignature := createClientSignature(storedKey, []byte(authMsg))

	if !checkClientProof(clientSignature, clientProof, storedKey) {
		return []byte(""), false
	}

	serverSignature := createServerSignature(serverKey, authMsg)
	return serverSignature, true

}
