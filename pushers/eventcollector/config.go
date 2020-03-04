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
package eventcollector

// Config defines a struct which holds configuration values for a SearchBackend.
type Config struct {
	Brokers []string `toml:"brokers"`
	Topic   string   `toml:"topic"`
	AgentName string `toml:"agent"`
	Mode string `toml:"mode"` 				// sync or async
	SecurityProtocol string `toml:"security_protocol"`
	SSLCAFile string `toml:"ssl_cafile"`
	SSLCertFile string `toml:"ssl_certfile"`
	SSLKeyFile string `toml:"ssl_keyfile"`
	SSLPassword string `toml:"ssl_password"`
}
