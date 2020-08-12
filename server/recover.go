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

package server

/*
// RecoverHandler defines a function which calls the provided ServeFunc
// within a protective recover() function.
func RecoverHandler(serveFn ServeFunc) error {
	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Errorf("Error: %s", err)
			log.Debugf("Stack of %d bytes: %s\n", count, string(trace))
			return
		}
	}()

	if err := serveFn(); err != nil {
		log.Error("Error: ", err)
		return err
	}

	return nil
}

*/
