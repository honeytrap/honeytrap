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
package ssh

import (
	_ "log"
	"net"
	"reflect"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func FromLUA(m lua.LValue) interface{} {
	if m == nil {
		return nil
	}

	// check for int?
	switch v := m.(type) {
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := map[string]interface{}{}

			for key := lua.LNil; ; {
				var value lua.LValue
				key, value = v.Next(key)
				if key == lua.LNil {
					break
				}

				keyVal := FromLUA(key)
				if v2, ok := keyVal.(string); ok {
					ret[v2] = FromLUA(value)
				}
			}
			return ret
		} else {
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, FromLUA(v.RawGetInt(i)))
			}
			return ret
		}
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case *lua.LBool:
		return bool(*v)
	case lua.LNumber:
		return float64(v)
	case *lua.LNumber:
		return float64(*v)
	case lua.LString:
		s := string(v)

		return string(s)
	case *lua.LString:
		return string(*v)
	default:
		log.Warningf("Type unsupported: %s %#+v", reflect.TypeOf(m), m)
		return v.String()
	}
}

func ToLUA(L *lua.LState, m interface{}) lua.LValue {
	if m == nil {
		return lua.LNil
	}

	// table.Append(lua.LString("BLA"))
	switch v := m.(type) {
	case map[string]interface{}:
		table := L.NewTable()
		for k, v2 := range v {
			table.RawSetString(k, ToLUA(L, v2))
		}
		return table
	case []interface{}:
		table := L.NewTable()
		for _, v2 := range v {
			table.Append(ToLUA(L, v2))
		}
		return table
	case map[interface{}]interface{}:
		table := L.NewTable()
		for k, v2 := range v {
			table.RawSet(ToLUA(L, k), ToLUA(L, v2))
		}
		return table
	case float64:
		return lua.LNumber(v)
	case time.Time:
		return lua.LString(v.Format(time.RFC3339))
	case net.IP:
		return lua.LString(v.String())
	case string:
		return lua.LString(v)
	default:
		log.Warningf("Type unsupported: %s %#+v", reflect.TypeOf(m), m)
	}

	return nil
}
