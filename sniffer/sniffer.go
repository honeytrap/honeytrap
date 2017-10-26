// +build ignore

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
package sniffer

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:sniffer")

// contains configuration fields.
var (
	timeout     = 10 * time.Millisecond
	promisc     = false
	offline     = false
	ErrNoSource = errors.New("No gopacket.Source")
)

// Sniffer defines a struct which handles network data capturing from containers.
type Sniffer struct {
	filter      string
	stopChan    chan bool
	stoppedChan chan bytes.Buffer
}

// New returns a new Sniffer instance.
func New(filter string) *Sniffer {
	return &Sniffer{
		filter:      filter,
		stopChan:    make(chan bool),
		stoppedChan: make(chan bytes.Buffer),
	}
}

// Start initializes and begins network data collection.
func (c *Sniffer) Start(device string) error {
	return c.serve(device)
}

// Stop returns a reader which provides access to captured data.
func (c *Sniffer) Stop() (io.ReadCloser, error) {
	log.Debug("Sniffer stopping")

	c.stopChan <- true

	log.Debug("Sniffer signalled")

	// wait for sniffer to stop
	buff := <-c.stoppedChan

	log.Debug("Sniffer stopped %d", buff.Len())

	return ioutil.NopCloser(&buff), nil
}

//serve begins collecting data packets and writing into the buffer
func (c *Sniffer) serve(device string) error {
	handle, err := pcap.OpenLive(device, 65616, promisc, timeout)
	if err != nil {
		return err
	}

	/*
		err = c.handle.SetBPFFilter(c.filter)
		if err != nil {
			return err
		}
	*/

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	if source == nil {
		return ErrNoSource
	}

	go func() {
		log.Info("Packet recorder started (%s)", device)

		// dont buffer in memory
		buffer := bytes.Buffer{}

		w := gzip.NewWriter(&buffer)

		defer func() {
			w.Close()

			c.stoppedChan <- buffer

			handle.Close()
			log.Debug("Packet recorder stopped.")
		}()

		gow := pcapgo.NewWriter(w)
		if err := gow.WriteFileHeader(65616, handle.LinkType()); err != nil {
			log.Error("pcapgo.WriterHeader: ", err)
			return
		}

		for {
			select {
			case <-c.stopChan:
				log.Debug("Got stop signal")
				return
			case packet := <-source.Packets():
				// TODO: should be pushed to channel, then in channel it can be filterer
				if err := gow.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
					log.Errorf("error writing packet: %+q", err)
				}
			}
		}

	}()

	return nil
}
