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

package cli_test

import (
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	out := run(t, "--help")
	// The output should contain an explanation of a few flags.
	// --help and --version are future-proof.
	shouldContain := []string{"--help", "--version"}
	for _, s := range shouldContain {
		if !strings.Contains(out, s) {
			t.Errorf("Help text doesn't contain '%s'", s)
		}
	}
}

func TestListServices(t *testing.T) {
	out := runWithConfig(t, "", "--list-services")
	/* The output should contain a list of services; http and telnet are some
	 * basic ones.
	 */
	shouldContain := []string{"http", "telnet"}
	for _, s := range shouldContain {
		if !strings.Contains(out, s) {
			t.Errorf("Services list doesn't contain '%s'", s)
		}
	}
}

func TestListChannels(t *testing.T) {
	out := runWithConfig(t, "", "--list-channels")
	/* The output should contain a list of channels; console and file are some
	 * basic ones.
	 */
	shouldContain := []string{"console", "file"}
	for _, s := range shouldContain {
		if !strings.Contains(out, s) {
			t.Errorf("Channels list doesn't contain '%s'", s)
		}
	}
}

func TestListListeners(t *testing.T) {
	out := runWithConfig(t, "", "--list-listeners")
	/* The output should contain a list of listeners; socket and agent are some
	 * basic ones.
	 */
	shouldContain := []string{"socket", "agent"}
	for _, s := range shouldContain {
		if !strings.Contains(out, s) {
			t.Errorf("Listeners list doesn't contain '%s'", s)
		}
	}
}
