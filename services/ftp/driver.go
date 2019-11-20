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
package ftp

import (
	"io"
	"os"
)

// For each client that connects to the server, a new FTPDriver is required.
// Create an implementation if this interface and provide it to FTPServer.
type DriverFactory interface {
	NewDriver() (Driver, error)
}

// You will create an implementation of this interface that speaks to your
// chosen models layer. graval will create a new instance of your
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
