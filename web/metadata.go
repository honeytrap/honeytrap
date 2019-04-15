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
package web

import (
	"encoding/json"
	"time"
)

type Metadata struct {
	Start         time.Time
	Version       string `json:"version"`
	ReleaseTag    string `json:"release_tag"`
	CommitID      string `json:"commitid"`
	ShortCommitID string `json:"shortcommitid"`
}

func (metadata Metadata) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	m["start"] = metadata.Start
	m["version"] = metadata.Version
	m["commitid"] = metadata.CommitID
	m["shortcommitid"] = metadata.ShortCommitID
	m["release_tag"] = metadata.ReleaseTag
	return json.Marshal(m)
}
