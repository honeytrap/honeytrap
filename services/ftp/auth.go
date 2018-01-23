package ftp

type Auth interface {
	CheckPasswd(string, string) (bool, error)
}

type FtpUser struct {
}

func (u *FtpUser) CheckPasswd(name, password string) (bool, error) {
	login := false

	if name == "anonymous" {
		login = true
	} else if name == "admin" && password == "god" {
		login = true
	}

	return login, nil
}
