package ftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"github.com/honeytrap/honeytrap/storage"
)

func Storage() (*ftpStorage, error) {
	if s, err := storage.Namespace("ftp"); err != nil {
		return nil, err
	} else {
		return &ftpStorage{
			s,
		}, nil
	}
}

type ftpStorage struct {
	storage.Storage
}

//Returns a TLS Certificate or nil on error
func (s *ftpStorage) Certificate() *tls.Certificate {

	keyname := "private-key"

	priv_b, err := s.Get(keyname)
	if err != nil {
		priv_b, err = generateKey()
		if err != nil {
			return nil
		}
		if err := s.Set(keyname, priv_b); err != nil {
			log.Errorf("Could not persist %s: %s", keyname, err.Error())
		}
	}

	cert, err := generateCert(priv_b)
	if err != nil {
		log.Errorf("Could not generate a Certificate. %s", err)
		return nil
	}

	return cert
}

func generateKey() ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		return nil, cerr
	}

	data := x509.MarshalPKCS1PrivateKey(priv)

	return data, nil
}

func generateCert(priv_b []byte) (*tls.Certificate, error) {
	log.Debug("START generateCert()")

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
		Issuer: pkix.Name{
			Country:            []string{""},
			Organization:       []string{""},
			OrganizationalUnit: []string{""},
			Locality:           []string{""},
			Province:           []string{""},
			StreetAddress:      []string{""},
			PostalCode:         []string{""},
			SerialNumber:       fmt.Sprintf("%d", 0),
			//CommonName:         s.Banner,
		},
		SignatureAlgorithm: x509.SHA512WithRSA,
		//PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		IsCA:        false,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := x509.ParsePKCS1PrivateKey(priv_b)
	if err != nil {
		log.Errorf("Could not parse private key: %s", err.Error())
		return nil, err
	}

	pub := priv.Public()

	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}, nil
}
