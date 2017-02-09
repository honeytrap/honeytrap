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

var (
	timeout     = 10 * time.Millisecond
	promisc     = false
	offline     = false
	ErrNoSource = errors.New("No gopacket.Source")
)

type Sniffer struct {
	filter      string
	stopChan    chan bool
	stoppedChan chan bytes.Buffer
}

func New(filter string) *Sniffer {
	return &Sniffer{
		filter:      filter,
		stopChan:    make(chan bool),
		stoppedChan: make(chan bytes.Buffer),
	}
}

func (c *Sniffer) Start(device string) error {
	return c.serve(device)
}

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
					log.Error("Error writing packet: %s", err.Error())
				}
			}
		}

	}()

	return nil
}
