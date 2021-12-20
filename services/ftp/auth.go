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
package ftp

// Auth interface for authentication
type Auth interface {
	CheckPasswd(string, string) (bool, error)
}

// User is a map[username]password
type User map[string]string

// CheckPasswd authenticate a user
func (u User) CheckPasswd(name, password string) (bool, error) {
	login := false

	if pw, ok := u[name]; ok && pw == password {
		login = true
	}

	return login, nil
}
