// Derived from javascript at http://www.movable-type.co.uk/scripts/geohash.html
//
// Original copyright states...
// "Geohash encoding/decoding and associated functions (c) Chris Veness 2014 / MIT Licence"

package geohash

import (
	"bytes"
	"errors"
	"strings"
)

// Encode latitude/longitude to geohash, either to specified precision or to automatically evaluated precision.
func Encode(lat, lon float64, precision int) (string, error) {
	var idx = 0 // index into base32 map
	var bit = 0 // each char holds 5 bits
	var evenBit = true
	var latMin = -90.0
	var latMax = 90.0
	var lonMin = -180.0
	var lonMax = 180.0
	if precision < 1 {
		return "", errors.New("invalid precision")
	}
	var geohash bytes.Buffer
	for geohash.Len() < precision {
		if evenBit {
			// bisect E-W longitude
			var lonMid = (lonMin + lonMax) / 2
			if lon > lonMid {
				idx = idx*2 + 1
				lonMin = lonMid
			} else {
				idx = idx * 2
				lonMax = lonMid
			}
		} else {
			// bisect N-S latitude
			var latMid = (latMin + latMax) / 2
			if lat > latMid {
				idx = idx*2 + 1
				latMin = latMid
			} else {
				idx = idx * 2
				latMax = latMid
			}
		}
		evenBit = !evenBit

		bit = bit + 1
		if bit == 5 {
			// 5 bits gives us a character: append it and start over
			b := base32F(idx)
			if b == '?' {
				return "", errors.New("encoding error")
			}
			geohash.WriteByte(b)
			bit = 0
			idx = 0
		}
	}
	return geohash.String(), nil
}

// Decode geohash to latitude/longitude (location is approximate centre of geohash cell, to reasonable precision).
func Decode(geohash string) (lat, lon float64, err error) {
	swLat, swLon, neLat, neLon, err1 := Bounds(geohash) // <-- the hard work
	if err1 != nil {
		return 0, 0, err1
	}
	return (neLat-swLat)/2 + swLat, (neLon-swLon)/2 + swLon, nil
}

// Bounds returns SW/NE latitude/longitude bounds of specified geohash.
func Bounds(geohash string) (swLat, swLon, neLat, neLon float64, err error) {
	geohash = strings.ToLower(geohash)
	var evenBit = true
	var latMin = -90.0
	var latMax = 90.0
	var lonMin = -180.0
	var lonMax = 180.0
	for i := 0; i < len(geohash); i++ {
		var chr = geohash[i]
		var idx = base32R(chr)
		if idx == -1 {
			return 0, 0, 0, 0, errors.New("invalid geohash")
		}
		for n := uint(4); ; n-- {
			var bitN = idx >> n & 1
			if evenBit {
				// longitude
				var lonMid = (lonMin + lonMax) / 2
				if bitN == 1 {
					lonMin = lonMid
				} else {
					lonMax = lonMid
				}
			} else {
				// latitude
				var latMid = (latMin + latMax) / 2
				if bitN == 1 {
					latMin = latMid
				} else {
					latMax = latMid
				}
			}
			evenBit = !evenBit
			if n == 0 {
				break
			}
		}
	}
	return latMin, lonMin, latMax, lonMax, nil
}

func base32R(b byte) int {
	switch b {
	default:
		return -1
	case '0':
		return 0
	case '1':
		return 1
	case '2':
		return 2
	case '3':
		return 3
	case '4':
		return 4
	case '5':
		return 5
	case '6':
		return 6
	case '7':
		return 7
	case '8':
		return 8
	case '9':
		return 9
	case 'b':
		return 10
	case 'c':
		return 11
	case 'd':
		return 12
	case 'e':
		return 13
	case 'f':
		return 14
	case 'g':
		return 15
	case 'h':
		return 16
	case 'j':
		return 17
	case 'k':
		return 18
	case 'm':
		return 19
	case 'n':
		return 20
	case 'p':
		return 21
	case 'q':
		return 22
	case 'r':
		return 23
	case 's':
		return 24
	case 't':
		return 25
	case 'u':
		return 26
	case 'v':
		return 27
	case 'w':
		return 28
	case 'x':
		return 29
	case 'y':
		return 30
	case 'z':
		return 31
	}
}

func base32F(i int) byte {
	switch i {
	default:
		return '?'
	case 0:
		return '0'
	case 1:
		return '1'
	case 2:
		return '2'
	case 3:
		return '3'
	case 4:
		return '4'
	case 5:
		return '5'
	case 6:
		return '6'
	case 7:
		return '7'
	case 8:
		return '8'
	case 9:
		return '9'
	case 10:
		return 'b'
	case 11:
		return 'c'
	case 12:
		return 'd'
	case 13:
		return 'e'
	case 14:
		return 'f'
	case 15:
		return 'g'
	case 16:
		return 'h'
	case 17:
		return 'j'
	case 18:
		return 'k'
	case 19:
		return 'm'
	case 20:
		return 'n'
	case 21:
		return 'p'
	case 22:
		return 'q'
	case 23:
		return 'r'
	case 24:
		return 's'
	case 25:
		return 't'
	case 26:
		return 'u'
	case 27:
		return 'v'
	case 28:
		return 'w'
	case 29:
		return 'x'
	case 30:
		return 'y'
	case 31:
		return 'z'
	}
}
