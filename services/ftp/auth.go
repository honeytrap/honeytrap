package ftp

type Auth interface {
	CheckPasswd(string, string) (bool, error)
}

type FtpUser struct {
	users map[string]string
}

func (u *FtpUser) CheckPasswd(name, password string) (bool, error) {
	login := false

	if pw, ok := u.users[name]; ok && pw == password {
		login = true
	}

	return login, nil
}
