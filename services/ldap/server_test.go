package ldap

import "testing"

func TestIsLogin(t *testing.T) {
	s := Server{
		login: "tester",
	}

	got := s.isLogin()

	if got != true {
		t.Errorf("isLogin (login name: %s) Got %v, Want true", s.login, got)
	}

	s.login = ""
	got = s.isLogin()

	if got != false {
		t.Errorf("isLogin (login name: '') Got %v, Want false", got)
	}
}

func TestDSEGet(t *testing.T) {
	str := "HT"

	d := &DSE{
		VendorName: []string{str},
	}

	got := d.Get()

	if name, ok := got.Attrs["vendorName"]; ok {
		if name[0] != str {
			t.Errorf("*DSE.Get Got %s, Want %s", name, str)
		}
	} else {
		t.Errorf("*DSE.Get Key: vendorName not in AttributeMap")
	}
}
