/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ftp

import (
	"io"
	"os"
)

// DriverFactory - For each client that connects to the server, a new FTPDriver is required.
// Create an implementation if this interface and provide it to FTPServer.
type DriverFactory interface {
	NewDriver() (Driver, error)
}

// Driver - You will create an implementation of this interface that speaks to your
// chosen persistence layer. graval will create a new instance of your
// driver for each client that connects and delegate to it as required.
type Driver interface {
	// Init init
	Init()

	// params  - a file path
	// returns - a time indicating when the requested path was last modified
	//         - an error if the file doesn't exist or the user lacks
	//           permissions
	Stat(string) (os.FileInfo, error)

	// params  - path
	// returns - true if the current user is permitted to change to the
	//           requested path
	ChangeDir(string) error

	// params  - path, function on file or subdir found
	// returns - []os.FileInfo
	//           path
	ListDir(string) []os.FileInfo

	// params  - path
	// returns - true if the directory was deleted
	DeleteDir(string) error

	// params  - path
	// returns - true if the file was deleted
	DeleteFile(string) error

	// params  - from_path, to_path
	// returns - true if the file was renamed
	Rename(string, string) error

	// params  - path
	// returns - true if the new directory was created
	MakeDir(string) error

	// params  - path
	// returns - a string containing the file data to send to the client
	GetFile(string, int64) (int64, io.ReadCloser, error)

	// params  - destination path, an io.Reader containing the file data
	// returns - true if the data was successfully persisted
	PutFile(string, io.Reader, bool) (int64, error)

	// returns - current directory
	CurDir() string
}
