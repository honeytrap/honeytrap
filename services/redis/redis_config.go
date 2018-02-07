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
	"reflect"
	"strconv"
	"strings"
)

var Redisdef = reflect.ValueOf(&RedisDefault).Elem()

func cleanVersion(v string) []int {
	n := strings.Split(v, ".")
	a := []int{}

	for _, m := range n {
		i, _ := strconv.Atoi(m)
		a = append(a, i)
	}
	return a
}

func compareVersions(persoV, defaultV []int) bool {
	for i := 0; i < len(defaultV); i++ {
		if persoV[i] > defaultV[i] {
			return true
			break
		} else if persoV[i] < defaultV[i] {
			return false
			break
		}
	}
	return true
}

func splitDefault(totalvalue string) []string {
	return strings.Split(totalvalue, ";")
}

func takePersoVersion(w string) []int {
	if w != "" {
		return cleanVersion(w)
	} else {
		return cleanVersion(splitDefault(RedisDefault.ServerVersion.Redis_version)[2])
	}

}

func (s *redisService) configureRedisService() RedisServiceConfiguration {

	redis_config_perso := s.RedisServiceConfig
	redisperso := reflect.ValueOf(redis_config_perso)

	redis_version_perso := takePersoVersion(redis_config_perso.Redis_version)

	for i := 0; i < Redisdef.NumField(); i++ {
		f := Redisdef.Field(i)

		for j := 0; j < f.NumField(); j++ {

			z := f.Field(j)

			defaultVersion := cleanVersion(splitDefault(z.Interface().(string))[0])

			tt := Redisdef.Field(i).Field(j)

			if compareVersions(redis_version_perso, defaultVersion) {
				if value_to_add := redisperso.Field(i).Field(j).Interface().(string); value_to_add != "" {
					tt.SetString(value_to_add)
				} else {
					defaultValue := splitDefault(string(tt.Interface().(string)))[2]
					tt.SetString(defaultValue)
				}
			} else {
				tt.SetString("__")
			}
		}
	}
	return RedisDefault
}
