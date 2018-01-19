package ftp

type Auth interface {
	CheckPasswd(string, string) (bool, error)
}
