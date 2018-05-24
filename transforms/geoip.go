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
package transforms

import (
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/honeytrap/honeytrap/event"
	"github.com/op/go-logging"
	"github.com/oschwald/geoip2-golang"
)

var (
	_          = Register("geoip", Geoip())
	geoLiteURL = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz"
)

func downloadGeoLiteDb(dest string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", geoLiteURL, nil)
	if err != nil {
		return err
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	gzf, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzf.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, gzf)
	if err != nil {
		return err
	}

	return nil
}

// Geoip adds
func Geoip() TransformFunc {
	log := logging.MustGetLogger("transforms/geoip")
	dbPath := path.Join(os.TempDir(), "GeoLite2-Country.mmdb")
	if err := downloadGeoLiteDb(dbPath); err != nil {
		log.Error(err.Error())
		return func(e event.Event, send func(event.Event)) { send(e) }
	}
	db, err := geoip2.Open(dbPath)
	if err != nil {
		log.Error(err.Error())
		return func(e event.Event, send func(event.Event)) { send(e) }
	}
	return func(e event.Event, send func(event.Event)) {
		if !e.Has("source-ip") {
			send(e)
			return
		}
		ipstr := e.Get("source-ip")
		ip := net.ParseIP(ipstr)
		country, err := db.Country(ip)
		if err != nil {
			log.Error(err.Error())
			send(e)
			return
		}
		e.Store("source-ip-country", country.Country.IsoCode)
		send(e)
	}
}
