package nscanary

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pkg/peek"
	"github.com/honeytrap/honeytrap/pushers"
)

type TLS struct {
	config *tls.Config
}

func NewTLSConf(certFile, keyFile string) TLS {
	certPem, keyPem, err := genCerts(certFile, keyFile)
	if err != nil {
		log.Errorf("failed setting TLS config: %v", err)
		return TLS{}
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Errorf("failed setting TLS config: %v", err)
		return TLS{}
	}

	return TLS{
		config: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
}

func (c TLS) MaybeTLS(conn net.Conn, event pushers.Channel) (net.Conn, error) {
	pconn := peek.NewConn(conn)

	return pconn, nil
}

// Handshake does a tls.Server handshake on conn and returns the resulting tls connection.
// It creates a tls event and pushes it in the events channel.
func (c TLS) Handshake(conn net.Conn, events pushers.Channel) (net.Conn, error) {
	tlsConn := tls.Server(conn, c.config)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	//TODO (jerry): Send tls data (JA3??)
	state := tlsConn.ConnectionState()

	events.Send(event.New(
		CanaryOptions,
		event.Category("tcp"),
		event.Type("tls"),
		event.SourceAddr(tlsConn.RemoteAddr()),
		event.DestinationAddr(tlsConn.LocalAddr()),
		event.Custom("tls-version", state.Version),
		event.Custom("tls-ciphersuite", tls.CipherSuiteName(state.CipherSuite)),
	))

	return tlsConn, nil
}

// genCerts creates a certificate from a certificate and a key file.
// when no files are given it generates a new certificate and key.
// Returns the PEM encoded certificate and key.
func genCerts(certFile, keyFile string) ([]byte, []byte, error) {
	var pemCert, pemKey bytes.Buffer

	//if certificate and key are provided, attempt to use them, otherwise generate self-signed ones
	if certFile != "" && keyFile != "" {
		file, err := os.Open(certFile)
		if err != nil {
			return nil, nil, fmt.Errorf("open(%s): %v", certFile, err)
		}
		io.Copy(&pemCert, file)
		file.Close()

		file, err = os.Open(keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("open(%s): %v", keyFile, err)
		}
		io.Copy(&pemKey, file)
		file.Close()

		return pemCert.Bytes(), pemKey.Bytes(), nil
	}

	rootCert, rootKey, err := genRootCert()
	if err != nil {
		return nil, nil, err
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNo, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 6, 0)
	cert := x509.Certificate{
		SerialNumber:          serialNo,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  false,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		Subject: pkix.Name{
			Organization: []string{"Acme Corp."},
			CommonName:   "Acme Corp.",
		},
		DNSNames: []string{"localhost"},
	}

	der, err := x509.CreateCertificate(rand.Reader, &cert, &rootCert, &key.PublicKey, rootKey)
	if err != nil {
		return nil, nil, err
	}
	if err = pem.Encode(&pemCert, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return nil, nil, err
	}

	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	if err = pem.Encode(&pemKey, &pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}); err != nil {
		return nil, nil, err
	}

	return pemCert.Bytes(), pemKey.Bytes(), nil
}

func genRootCert() (x509.Certificate, *ecdsa.PrivateKey, error) {
	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0)
	serialNo, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return x509.Certificate{}, nil, err
	}

	cert := x509.Certificate{
		SerialNumber:          serialNo,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		Subject: pkix.Name{
			Organization: []string{"Acme Root"},
			CommonName:   "Acme Root CA",
		},
	}

	return cert, key, nil
}
