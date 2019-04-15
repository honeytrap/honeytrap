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
package elasticsearch

import (
	"context"
	"errors"
	"syscall"

	"net/http"
	"time"
)

type Retrier struct {
}

func (r Retrier) Retry(ctx context.Context, retry int, req *http.Request, resp *http.Response, err error) (time.Duration, bool, error) {
	if err == syscall.ECONNREFUSED {
		log.Error("Elasticsearch or network down. Reconnecting in 60 seconds.")
		return time.Duration(time.Second * 60), true, errors.New("Elasticsearch or network down")
	}

	log.Error("Error connecting to Elasticsearch: %s", err.Error())
	return time.Duration(time.Second * 10), true, err
}
