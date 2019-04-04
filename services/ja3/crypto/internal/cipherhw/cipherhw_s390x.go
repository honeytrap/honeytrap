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
package cipherhw

// hasHWSupport reports whether the AES-128, AES-192 and AES-256 cipher message
// (KM) function codes are supported. Note that this function is expensive.
// defined in asm_s390x.s
func hasHWSupport() bool

var hwSupport = hasHWSupport()

func AESGCMSupport() bool {
	return hwSupport
}
