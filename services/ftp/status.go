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

// FTP status codes, defined in RFC 959
const (
	StatusInitiating    = 100
	StatusRestartMarker = 110
	StatusReadyMinute   = 120
	StatusAlreadyOpen   = 125
	StatusAboutToSend   = 150

	StatusCommandOK             = 200
	StatusCommandNotImplemented = 202
	StatusSystem                = 211
	StatusDirectory             = 212
	StatusFile                  = 213
	StatusHelp                  = 214
	StatusName                  = 215
	StatusReady                 = 220
	StatusClosing               = 221
	StatusDataConnectionOpen    = 225
	StatusClosingDataConnection = 226
	StatusPassiveMode           = 227
	StatusLongPassiveMode       = 228
	StatusExtendedPassiveMode   = 229
	StatusLoggedIn              = 230
	StatusLoggedOut             = 231
	StatusLogoutAck             = 232
	StatusRequestedFileActionOK = 250
	StatusPathCreated           = 257

	StatusUserOK             = 331
	StatusLoginNeedAccount   = 332
	StatusRequestFilePending = 350

	StatusNotAvailable             = 421
	StatusCanNotOpenDataConnection = 425
	StatusTransfertAborted         = 426
	StatusInvalidCredentials       = 430
	StatusHostUnavailable          = 434
	StatusFileActionIgnored        = 450
	StatusActionAborted            = 451
	Status452                      = 452

	StatusBadCommand              = 500
	StatusBadArguments            = 501
	StatusNotImplemented          = 502
	StatusBadSequence             = 503
	StatusNotImplementedParameter = 504
	StatusNotLoggedIn             = 530
	StatusStorNeedAccount         = 532
	StatusFileUnavailable         = 550
	StatusPageTypeUnknown         = 551
	StatusExceededStorage         = 552
	StatusBadFileName             = 553
)

var statusText = map[int]string{
	// 200
	StatusCommandOK:             "OK.",
	StatusCommandNotImplemented: "Command not implemented, obsolete.",
	StatusSystem:                "System status, or system help reply.",
	StatusDirectory:             "Directory status.",
	StatusFile:                  "File status.",
	StatusHelp:                  "Help message.",
	StatusName:                  "",
	StatusReady:                 "Service ready for new user.",
	StatusClosing:               "Service closing control connection.",
	StatusDataConnectionOpen:    "Data connection open; no transfer in progress.",
	StatusClosingDataConnection: "Closing data connection. Requested file action successful.",
	StatusPassiveMode:           "Entering Passive Mode.",
	StatusLongPassiveMode:       "Entering Long Passive Mode.",
	StatusExtendedPassiveMode:   "Entering Extended Passive Mode.",
	StatusLoggedIn:              "User logged in, proceed.",
	StatusLoggedOut:             "User logged out; service terminated.",
	StatusLogoutAck:             "Logout command noted, will complete when transfer done.",
	StatusRequestedFileActionOK: "Requested file action okay, completed.",
	StatusPathCreated:           "Path created.",

	// 300
	StatusUserOK:             "User name OK, need password.",
	StatusLoginNeedAccount:   "Need account for login.",
	StatusRequestFilePending: "Requested file action pending further information.",

	// 400
	StatusNotAvailable:             "Service not available, closing control connection.",
	StatusCanNotOpenDataConnection: "Can't open data connection.",
	StatusTransfertAborted:         "Connection closed; transfer aborted.",
	StatusInvalidCredentials:       "Invalid username or password.",
	StatusHostUnavailable:          "Requested host unavailable.",
	StatusFileActionIgnored:        "Requested file action not taken.",
	StatusActionAborted:            "Requested action aborted. Local error in processing.",
	Status452:                      "Insufficient storage space in system.",

	// 500
	StatusBadCommand:              "Command not found.",
	StatusBadArguments:            "Syntax error in parameters or arguments.",
	StatusNotImplemented:          "Command not implemented.",
	StatusBadSequence:             "Bad sequence of commands.",
	StatusNotImplementedParameter: "Command not implemented for that parameter.",
	StatusNotLoggedIn:             "Not logged in.",
	StatusStorNeedAccount:         "Need account for storing files.",
	StatusFileUnavailable:         "File unavailable.",
	StatusPageTypeUnknown:         "Page type unknown.",
	StatusExceededStorage:         "Exceeded storage allocation.",
	StatusBadFileName:             "File name not allowed.",
}
