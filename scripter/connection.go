package scripter

// Connection Wrapper struct
type ConnectionStruct struct {
	Service string
	MyConn  ScrConn
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (w *ConnectionStruct) Handle(message string) (string, error) {
	result := message
	var err error

	result, err = w.MyConn.HandleScripts(w.Service, result)

	if err != nil {
		log.Errorf("Error while handling scripts: %s", err)
	}

	return result, nil
}

//Set a string function for a connection
func (w *ConnectionStruct) SetStringFunction(name string, getString func() string) error {
	return w.MyConn.SetStringFunction(name, getString, w.Service)
}

//Set a string function for a connection
func (w *ConnectionStruct) SetFloatFunction(name string, getFloat func() float64) error {
	return w.MyConn.SetFloatFunction(name, getFloat, w.Service)
}

//Set a string function for a connection
func (w *ConnectionStruct) SetVoidFunction(name string, doVoid func()) error {
	return w.MyConn.SetVoidFunction(name, doVoid, w.Service)
}

//Get a parameter from a connection
func (w *ConnectionStruct) GetParameter(index int) (string, error) {
	return w.MyConn.GetParameter(index, w.Service)
}
