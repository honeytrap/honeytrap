package scripter

//ConnectionStruct
type ConnectionStruct struct {
	Service string
	Conn    ScrConn
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (w *ConnectionStruct) Handle(message string) (string, error) {
	result, err := w.Conn.Handle(w.Service, message)

	if err != nil {
		log.Errorf("Error while handling scripts: %s", err)
	}

	if result != nil {
		return result.Content, nil
	}

	return "", nil
}

//SetStringFunction sets a string function for a connection
func (w *ConnectionStruct) SetStringFunction(name string, getString func() string) error {
	return w.Conn.SetStringFunction(name, getString, w.Service)
}

//SetFloatFunction sets a string function for a connection
func (w *ConnectionStruct) SetFloatFunction(name string, getFloat func() float64) error {
	return w.Conn.SetFloatFunction(name, getFloat, w.Service)
}

//SetVoidFunction sets a string function for a connection
func (w *ConnectionStruct) SetVoidFunction(name string, doVoid func()) error {
	return w.Conn.SetVoidFunction(name, doVoid, w.Service)
}

//GetParameters gets a parameter from a connection
func (w *ConnectionStruct) GetParameters(params []string) (map[string]string, error) {
	return w.Conn.GetParameters(params, w.Service)
}
