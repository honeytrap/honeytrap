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

package honeytrap_test

import (
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/satori/go.uuid"
)

func newTmpFilename() string {
	return path.Join(os.TempDir(), "honeytrap-tests-"+uuid.NewV4().String())
}

// Generates a TOML config for a port linked to a single service.
func serviceWithPort(svc string, port string) string {
	svcId := uuid.NewV4().String()
	return `[[port]]
	port="` + port + `"
	services=["` + svcId + `"]

	[service.` + svcId + `]
	type="` + svc + `"
	`
}

var honeytrapBinary = path.Join("/honeytrap", "honeytrap")

func runWithConfig(config string, flags ...string) (confPath string, p *os.Process) {
	config += `[listener]
type="socket"

[[logging]]
output = "stdout" # This is required, otherwise Honeytrap won't start
level = "debug"`
	confPath = newTmpFilename()
	tmp, err := os.Create(confPath)
	if err != nil {
		panic(err)
	}
	_, err = tmp.WriteString(config)
	if err != nil {
		panic(err)
	}
	// Argv[0] is the current process.
	flags = append([]string{honeytrapBinary}, flags...)
	flags = append(flags, "-c", confPath)

	p, err = os.StartProcess(honeytrapBinary, flags, &os.ProcAttr{
		Files: []*os.File{
			nil,
			os.Stdout,
			os.Stderr,
		},
	})
	if err != nil {
		panic(err)
	}

	// Allow Honeytrap to load stuff
	time.Sleep(1 * time.Second)
	return
}

var isNmapAvailable = exec.Command("nmap", "--version").Run() == nil

// Return the nmap identification (eg. "Apache httpd")
func nmapIdentify(t *testing.T, portNum string) string {
	if !isNmapAvailable {
		t.Skipf("Nmap is not installed (`command -v nmap` failed)")
	}
	nmapOutput := newTmpFilename()
	defer os.Remove(nmapOutput)
	cmd := exec.Command("nmap",
		"-Pn",             // Skip ping check
		"-sV",             // Detect service type
		"-oX", nmapOutput, // Write XML output to file
		"-p"+portNum,
		"127.0.0.1",
	)
	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}
	f, err := os.Open(nmapOutput)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	doc, err := xmlquery.Parse(f)
	if err != nil {
		panic(err)
	}
	state := xmlquery.FindOne(doc, "/nmaprun/host/ports/port[1]/state").SelectAttr("state")
	if state != "open" {
		t.Errorf("Expected 'open' state, found '%s'", state)
	}
	return xmlquery.FindOne(doc, "/nmaprun/host/ports/port[1]/service").SelectAttr("product")
}

func mustWait(p *os.Process) *os.ProcessState {
	ret, err := p.Wait()
	if err != nil {
		panic(err)
	}
	return ret
}
