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
package ldap

const (
	// LDAP App Codes
	AppBindRequest           = 0
	AppBindResponse          = 1
	AppUnbindRequest         = 2
	AppSearchRequest         = 3
	AppSearchResultEntry     = 4
	AppSearchResultDone      = 5
	AppModifyRequest         = 6
	AppModifyResponse        = 7
	AppAddRequest            = 8
	AppAddResponse           = 9
	AppDelRequest            = 10
	AppDelResponse           = 11
	AppModifyDNRequest       = 12
	AppModifyDNResponse      = 13
	AppCompareRequest        = 14
	AppCompareResponse       = 15
	AppAbandonRequest        = 16
	AppSearchResultReference = 19
	AppExtendedRequest       = 23
	AppExtendedResponse      = 24

	// LDAP result codes
	ResSuccess                  = 0
	ResOperationsError          = 1
	ResProtocolError            = 2
	ResNoSuchObject             = 32
	ResInvalidCred              = 49
	ResInsufficientAccessRights = 50
	ResUnwillingToPerform       = 53
	ResOther                    = 80
)
