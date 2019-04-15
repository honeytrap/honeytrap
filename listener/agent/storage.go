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
package agent

import (
	"encoding/hex"

	"github.com/mimoo/disco/libdisco"

	"github.com/honeytrap/honeytrap/storage"
)

func Storage() (*agentListenerStorage, error) {
	s, err := storage.Namespace("agent")
	if err == nil {
		return &agentListenerStorage{
			s,
		}, nil
	}
	return nil, err
}

type agentListenerStorage struct {
	storage.Storage
}

func (s *agentListenerStorage) KeyPair() (*libdisco.KeyPair, error) {
	keyPair := &libdisco.KeyPair{}

	if key, err := s.Get("key"); err == nil {
		if _, err = hex.Decode(keyPair.PublicKey[:], key[64:]); err != nil {
			return nil, err
		} else if _, err = hex.Decode(keyPair.PrivateKey[:], key[:64]); err != nil {
			return nil, err
		}

		return keyPair, nil
	}

	key := make([]byte, 128)

	keyPair = libdisco.GenerateKeypair(nil)
	hex.Encode(key[:64], keyPair.PrivateKey[:])
	hex.Encode(key[64:], keyPair.PublicKey[:])

	if err := s.Set("key", key); err != nil {
		log.Errorf("Could not persist key: %s", err.Error())
		return nil, err
	}

	return keyPair, nil
}
