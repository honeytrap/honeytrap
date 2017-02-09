package server

import "net"

//Controller defines the generic properties for controller
type Controller interface {
	Name() string
	Device() (string, error)
	// TODO:
	// Maybe we should have the container be responsible for itself, so idle timing etc.
	// that way we can create also a simple honeypot (low interaction) controller only with stream as well
	// then IsIdle, Device, Name, Etc are not important anymore, or we can also just solve this
	// with having an other SSHProxyListener, so that will be solved already then....
	IsIdle() bool
	Dial(string) (net.Conn, error)
	CleanUp() error
}
