package ftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/honeytrap/honeytrap/storage"
)

func getStorage() (*ftpStorage, error) {
	s, err := storage.Namespace("ftp")
	if err != nil {
		return nil, err
	}
	return &ftpStorage{
		s,
	}, nil
}

type ftpStorage struct {
	storage.Storage
}

func (s *ftpStorage) FileSystem() (base, serviceroot string) {
	b, err := s.Get("base")
	if err != nil {
		return "", ""
	}
	base = string(b)

	sr, err := s.Get("fs_root")
	if err != nil {
		return "", ""
	}
	serviceroot = string(sr)

	return
}

//Returns a TLS Certificate
func (s *ftpStorage) Certificate() (*tls.Certificate, error) {

	keyname := "pemkey"
	certname := "pemcert"

	pemkey, err := s.Get(keyname)
	if err != nil {
		pemkey, err = generateKey()
		if err != nil {
			return nil, err
		}
		if err = s.Set(keyname, pemkey); err != nil {
			log.Errorf("Could not persist %s: %s", keyname, err.Error())
		}
	}

	pemcert, err := s.Get(certname)
	if err != nil {
		pemcert, err = generateCert(pemkey)
		if err != nil {
			return nil, err
		}
		if err = s.Set(certname, pemcert); err != nil {
			log.Errorf("Could not persist %s: %s", certname, err.Error())
		}
	}

	tlscert, err := tls.X509KeyPair(pemcert, pemkey)
	if err != nil {
		return nil, err
	}

	return &tlscert, nil
}

//Returns a PEM encoded RSA private key
func generateKey() ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		return nil, cerr
	}

	pemdata := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return pemdata, nil
}

func generateCert(pempriv []byte) ([]byte, error) {

	snLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, snLimit)

	if err != nil {
		log.Debug("Could not generate certificate serial number")
	}

	ca := &x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:            []string{""},
			Organization:       []string{""},
			OrganizationalUnit: []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		//IsCA:        false,
		//ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		//KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	block, _ := pem.Decode(pempriv)

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Could not parse private key: %s", err.Error())
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, priv.Public(), priv)
	if err != nil {
		return nil, err
	}

	certpem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return certpem, nil
}
