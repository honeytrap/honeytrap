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
package scripter

import (
	"encoding/json"
	"github.com/honeytrap/honeytrap/utils/files"
	"io/ioutil"
	"strings"
	"os"
	"encoding/base64"
	"errors"
)

var basepath = "scripts/"

// fileInfo covers the file info for responses
type fileInfo struct {
	Path string `json:"path"`
	Content string `json:"content"`
}

// response struct is used for JSON responses
type response struct {
	Type string `json:"type"`
	Data interface{} `json:"data"`
}

// HandleRequests handles the request coming from other environments
func HandleRequests(scripters map[string]Scripter, message []byte) ([]byte, error) {
	var js map[string]interface{}
	json.Unmarshal(message, &js)

	switch val, _ := js["action"]; val {
	case "script_reload":
		return handleScriptReload(scripters)
	case "script_put":
		return handleScriptPut(js)
	case "script_delete":
		return handleScriptDelete(js)
	case "script_read":
		return handleScriptRead(js)
	}

	return nil, nil
}

// handleScriptReload handles the reload script web request
func handleScriptReload(scripters map[string]Scripter) ([]byte, error) {
	ReloadAllScripters(scripters)

	return nil, nil
}

// handleScriptRead handles the read script web request
func handleScriptRead(js map[string]interface{}) ([]byte, error) {
	dir, ok := js["dir"].(string)
	if !ok {
		dir = ""
	}

	arrFileInfo, err := readFiles(dir)
	if err != nil {
		return nil, err
	}

	return generateResponse("scripts", arrFileInfo)
}

// handleScriptPut handles the put script web request
func handleScriptPut(js map[string]interface{}) ([]byte, error) {
	path, ok := js["path"].(string)
	if !ok {
		return nil, errors.New("undefined script put path")
	}

	content, ok := js["file"].(string)
	if !ok {
		return nil, errors.New("undefined script content")
	}

	files.Put(basepath + path, content)

	return nil, nil
}

// handleScriptDelete handles the delete script web request
func handleScriptDelete(js map[string]interface{}) ([]byte, error) {
	path, ok := js["path"].(string)
	if !ok {
		return nil, errors.New("undefined script delete path")
	}

	if err := files.Delete(basepath + path); err != nil {
		return nil, err
	}

	return nil, nil
}

// readFiles reads the files in the scripts directory
func readFiles(dir string) ([]fileInfo, error) {
	var arrFileInfo []fileInfo

	dirFiles, err := files.Walker(basepath + dir)
	if err != nil {
		return nil, err
	}

	for _, file := range dirFiles {
		content, err := ioutil.ReadFile(basepath + dir + file)
		if err != nil {
			return nil, err
		}

		arrFileInfo = append(arrFileInfo, fileInfo{Path: strings.Replace(basepath + dir + file, string(os.PathSeparator), "/", -1), Content: base64.StdEncoding.EncodeToString(content)})
	}

	return arrFileInfo, nil
}

// generateResponse generates a JSON response for web requests
func generateResponse(responseType string, data interface{}) ([]byte, error) {
	response := response{ Type: responseType, Data: data }
	return json.Marshal(response)
}
