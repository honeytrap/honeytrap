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
package cefsyslog

import (
	"fmt"
	"log/syslog"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"github.com/BurntSushi/toml"

)

var (
	_ = pushers.Register("cefsyslog", New)
)

var (
	log = logging.MustGetLogger("cefsyslog")
)

//[[service.mapping]] and [[service.static]] in toml
type mapping struct {
	CEFfield string
	Logfield string
}

type service struct {
	Name     string
	Type     string
	DynamicMappings []mapping `toml:"mapping"`
	StaticMappings []mapping `toml:"static"`
}

type CefMapping struct {
	Services []service `toml:"service"`
}

//Config information retrieved from main config.toml
type SyslogConfig struct {
	Address string `toml:"address"` // Like localhost:5672"
	Protocol string `toml:"protocol"` // Like "tcp" or "udp"
	Heartbeat string `toml:"heartbeat"` // Like "True" or "False"
}

//When using the CEFsyslog pusher is initialized the current field mappings are retrieved from the mappings.toml
func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	var mappingConfig CefMapping
	if _, err := toml.DecodeFile("./pushers/cefsyslog/mappings.toml", &mappingConfig); err != nil {
		log.Fatal(err)
	}


	c := CefsyslogObject{
		SyslogConfig: SyslogConfig{},
		CefMapping: mappingConfig,
		ch:	      ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	if c.SyslogConfig.Address == "" {
               return nil, fmt.Errorf("Syslog address not set")
        }

	if c.SyslogConfig.Protocol == "" {
                return nil, fmt.Errorf("Syslog protocol not set")
        }

	cefsyslog, err := syslog.Dial(c.SyslogConfig.Protocol, c.SyslogConfig.Address , syslog.LOG_NOTICE|syslog.LOG_LOCAL7, "HoneyTrap")
	if err != nil {
		log.Fatal(err)
	}

	c.SyslogChannel = cefsyslog



	return &c, nil

}

type CefsyslogObject struct {
	CefMapping
	SyslogConfig
	ch          chan map[string]interface{}
	SyslogChannel *syslog.Writer
}

func (b *CefsyslogObject) Send(e event.Event) {
	var cef_header string
	var cef_body_mapped string
	var cef_body_additional string
	var cef_message string
	var millis int64
	var DynamicMapping []mapping
	var StaticMapping []mapping

	// Array of names that are statically mapped and can be ignored during dynamic mapping
	BlackList := []string{"date", "category",  "type", "source-ip", "source-port", "destination-ip", "destination-port" , "payload" }

	millis = e.GetDate("date").UnixNano() / 1000000

	DynamicMapping, StaticMapping = getMapping( e.Get("category"), e.Get("type"), b.CefMapping)

	cef_header = fmt.Sprintf("CEF:0|Honeytrap|Honeytrap|1|%s-%s|%s %s|5|",  escapeCEFHeaderInput(e.Get("category")), escapeCEFHeaderInput(e.Get("type")), escapeCEFHeaderInput(e.Get("category")), escapeCEFHeaderInput(e.Get("type")))

	cef_body_mapped = fmt.Sprintf("rt=%d", millis)
	cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, getCEFKeyValue("dvc", e.Get("destination-ip")))

	switch e.Get("category") {
	case "heartbeat":
		if  b.Heartbeat != "True" {
			return
		}
	default:
		cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, getCEFKeyValue("dpt", e.Get("destination-port")))
		cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, getCEFKeyValue("dst", e.Get("destination-ip")))
		cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, getCEFKeyValue("spt", e.Get("source-port")))
		cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, getCEFKeyValue("src", e.Get("source-ip")))
		
		e.Range(func(key, value interface{}) bool {
			if keyName, ok := key.(string); ok {
				if !stringInSlice(keyName, BlackList) {
					mapped := false
					for _, mapping_item := range DynamicMapping {
						if mapping_item.Logfield == keyName{
							keyPair := getCEFKeyValue(mapping_item.CEFfield, e.Get(keyName))
							if keyPair != ""{
								cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, keyPair)
							}
							mapped = true
						}
					}
					if !mapped {
						keyPair := getCEFKeyValue(keyName ,e.Get(keyName))
						if keyPair != "" {
							cef_body_additional = fmt.Sprintf("%s ad.%s" , cef_body_additional, keyPair)
						}	
					}
				}
			}
			return true
		})

		for _, mapping_item := range StaticMapping {
			keyPair := getCEFKeyValue(mapping_item.CEFfield, mapping_item.Logfield)
			if keyPair != ""{
				cef_body_mapped = fmt.Sprintf("%s %s", cef_body_mapped, keyPair)
			}
		}
		cef_message = fmt.Sprintf("%s%s%s" ,cef_header, cef_body_mapped, cef_body_additional)
		fmt.Fprintf(b.SyslogChannel, cef_message)

	}



}

//getMapping returns the defined mappings from the config based on the service name and type.
func getMapping(serviceName string, serviceType string, serviceConfig CefMapping) ([]mapping, []mapping) {

	for _, ser := range serviceConfig.Services {
		if ser.Name == serviceName && ser.Type == serviceType {
			return ser.DynamicMappings, ser.StaticMappings
		}
	}
	return nil, nil
}

//getCEFKeyValue returns the escaped key value format for the cef body in the %s=%s form.
//If the value is empty the function will return nil
func getCEFKeyValue(key string, value string) string{
	escapedKey := escapeCEFBodyInput(key)
	escapedValue := escapeCEFBodyInput(value)
	if escapedValue != "" {
		return fmt.Sprintf("%s=%s", escapedKey, escapedValue)
	}
	return ""
}

//Escape "|"
func escapeCEFHeaderInput(header string) string{
	return strings.Replace(header, "|", "\\|", -1)
}

//Escape "=", "\"
func escapeCEFBodyInput(body string) string{
	escapedBody := strings.Replace(body, "\\", "\\\\", -1)
	escapedBody = strings.Replace(escapedBody, "=", "\\=", -1)
	return escapedBody
}

//stringInSlice returns true if a string is already present in a slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}



