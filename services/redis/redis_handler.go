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
package redis

import (
	"fmt"
	"strings"
)

func CleanCmd(cmd string) []string {
	argsCmd := []string{}

	for _, arg := range strings.Split(cmd, " ") {
		if arg == "" {
			continue
		}
		arg = strings.ToLower(arg)
		argsCmd = append(argsCmd, arg)
	}
	return argsCmd
}

func REDISHandler(s *redisService, cmd string) (string, bool) {
	argsCmd := CleanCmd(cmd)
	if fnCmd, ok := mapCmds[argsCmd[0]]; ok {
		return fnCmd(s, argsCmd, cmd)
	} else {
		return fmt.Sprintf(unknownCmdMsg, userCmd), false
	}
}
