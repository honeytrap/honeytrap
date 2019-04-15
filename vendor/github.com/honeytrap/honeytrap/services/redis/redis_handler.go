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
package redis

import (
	"fmt"
	"strings"
)

/* args is an array of interface{} because it could be either a redisDatum or
 * an elementary datatype (int, string, etc.)
 */
func (s *redisService) REDISHandler(command string, args []interface{}) (string, bool) {
	// Convert the command to lowercase
	command = strings.ToLower(command)
	fn, ok := mapCmds[command]
	if !ok {
		return fmt.Sprintf(errorMsg("unknown"), command), false
	}
	return fn(s, args)
}
