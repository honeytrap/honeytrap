package models

import "github.com/honeytrap/honeytrap/pushers/eventcollector/persistence"

type Model interface {
	Create() error
	Get() error
	Find() error
}

var (
	db = persistence.Connect()
)
