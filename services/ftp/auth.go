package ftp

type Auth interface {
	CheckPasswd(string, string) (bool, error)
}

type User struct {
	users map[string]string
}

func (u *User) CheckPasswd(name, password string) (bool, error) {
	login := false

	if pw, ok := u.users[name]; ok && pw == password {
		login = true
	}

	return login, nil
}
