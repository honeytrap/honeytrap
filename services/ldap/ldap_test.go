package ldap

var (
	abandonRequest = []byte{
		0x30, 0x06, // start sequence
		0x02, 0x01, 0x06, // message ID (6)
		0x50, 0x01, 0x05, // the abandon request
		// There is no response for this one
	}
)
