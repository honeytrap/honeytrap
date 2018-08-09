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

// This file contains test utilities, no actual tests.

package cli_test

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/satori/go.uuid"
)

func run(t *testing.T, flags ...string) string {
	return runWithErrHandler(func(e error) {
		t.Fatal(e)
	}, flags...)
}

func runWithConfig(t *testing.T, config string, flags ...string) string {
	return runWithConfigAndErrHandler(t, config, func(e error) { t.Fatal(e) }, flags...)
}

func fnHoneytrapPath() string {
	if os.Getenv("TRAVIS") == "" {
		return path.Join("/honeytrap", "honeytrap")
	} else {
		return path.Join(os.Getenv("HOME"), "honeytrap")
	}
}
var honeytrapPath = fnHoneytrapPath()

func runWithErrHandler(errHandler func(error), flags ...string) string {
	cmd := exec.Command(honeytrapPath, flags...)
	out, err := cmd.Output()
	if err != nil {
		errHandler(err)
	}
	return string(out)
}

func runWithConfigAndErrHandler(t *testing.T, config string, errHandler func(error), flags ...string) string {
	tmpPath := path.Join(os.TempDir(), "honeytrap-testfile-"+uuid.NewV4().String())
	tmp, err := os.Create(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmp.WriteString(config)
	if err != nil {
		t.Fatal(err)
	}
	flags = append(flags, "-c", tmpPath)
	ret := runWithErrHandler(errHandler, flags...)
	os.Remove(tmpPath)
	return ret
}
