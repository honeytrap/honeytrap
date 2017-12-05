package services

import "bytes"

type ippMessage struct {
	versionMajor int8
	versionMinor int8
	statusCode   int16
	requestId    int32
	nameLenght   int16 //number of octets in name
	name         string
	valueLenght  int16 //number of octets in value
	value        string
}

func ippHandler(body *bytes.Buffer) {
}
