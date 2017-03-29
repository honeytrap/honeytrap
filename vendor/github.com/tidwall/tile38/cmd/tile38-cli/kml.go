package main

import (
	"bytes"
	"fmt"

	"github.com/tidwall/tile38/geojson"
)

type pointT struct {
	name  string
	point geojson.SimplePoint
}

// KML represents a KML object.
type KML struct {
	points []pointT
}

// NewKML returns a new KML object.
func NewKML() *KML {
	return &KML{}
}

// AddPoint adds a point to a KML object.
func (kml *KML) AddPoint(name string, lat, lon float64) {
	kml.points = append(kml.points, pointT{name: name, point: geojson.SimplePoint{X: lon, Y: lat}})
}

// Bytes returns the xml of the KML.
func (kml *KML) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<kml xmlns="http://www.opengis.net/kml/2.2">` + "\n")
	buf.WriteString(`<Document>` + "\n")
	buf.WriteString(`<Style id="yellow"><IconStyle><Icon><href>http://www.google.com/intl/en_us/mapfiles/ms/icons/yellow-dot.png</href></Icon></IconStyle></Style> ` + "\n")
	buf.WriteString(`<Style id="blue"><IconStyle><Icon><href>http://www.google.com/intl/en_us/mapfiles/ms/icons/blue-dot.png</href></Icon></IconStyle></Style> ` + "\n")
	for _, point := range kml.points {
		buf.WriteString(fmt.Sprintf(`<Placemark><styleUrl>#yellow</styleUrl><name>%s</name><Point><coordinates>%f,%f,0</coordinates></Point></Placemark>`+"\n", point.name, point.point.X, point.point.Y))
	}
	buf.WriteString(`</Document></kml>` + "\n")
	return buf.Bytes()
}
