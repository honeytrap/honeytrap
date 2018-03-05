package ftp

type Auth interface {
	CheckPasswd(string, string) (bool, error)
}

type FtpUser map[string]string

/*
type FtpUser struct {
	users map[string]string
}
*/
func (u FtpUser) CheckPasswd(name, password string) (bool, error) {
	login := false

	if pw, ok := u[name]; ok && pw == password {
		login = true
	}

	return login, nil
}
