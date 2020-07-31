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
	//TODO (jerry): Hardcoded ip address!!!
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

// MaybeTLS checks for a tls signature and does a tls handshake if it is tls.
// return a tls.Conn or the given connection.
func (c TLS) MaybeTLS(conn net.Conn, events pushers.Channel) (net.Conn, error) {
	var signature [3]byte

	pconn := peek.NewConn(conn)
	if _, err := pconn.Peek(signature[:]); err != nil {
		return nil, err
	}

	fmt.Printf("signature = %x\n", signature)

	if signature[0] == 0x16 && signature[1] == 0x03 && signature[2] <= 0x03 {
		return c.Handshake(pconn, events)
	}

	return pconn, nil
}

var tlsVersion = map[uint16]string{
	tls.VersionTLS10: "1.0",
	tls.VersionTLS11: "1.1",
	tls.VersionTLS12: "1.2",
	tls.VersionTLS13: "1.3",
	tls.VersionSSL30: "SSL 3.0",
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
		event.Custom("tls-version", tlsVersion[state.Version]),
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

		log.Debugf("loaded certificate %s, key %s", certFile, keyFile)

		return pemCert.Bytes(), pemKey.Bytes(), nil
	}

	caCert, caKey, err := genRootCert()
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
		SerialNumber: serialNo,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		IsCA:         false,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	var certPrivKeyPEM bytes.Buffer
	pem.Encode(&certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

// returns the PEM encoded ca-certificate and private key.
func genRootCert() (*x509.Certificate, *ecdsa.PrivateKey, error) {
	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0)
	serialNo, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	ca := x509.Certificate{
		SerialNumber:          serialNo,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
	}

	/*
		//generate certificate and PEM encode.
		caBytes, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &key.PublicKey, key)
		if err != nil {
			return nil, nil, err
		}

		var caPEM bytes.Buffer
		pem.Encode(&caPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		})

		keyBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, nil, err
		}

		var caPrivKeyPEM bytes.Buffer
		pem.Encode(&caPrivKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		})
	*/

	return &ca, key, nil
}
