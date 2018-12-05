package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"time"
)

func NewCertificate(certFile, keyFile, commonName string, keyUsage x509.ExtKeyUsage) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{keyUsage},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0666); err != nil {
		return err
	}

	if err := ioutil.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0600); err != nil {
		return err
	}

	return nil
}

func ReadCertificate(certFile, keyFile string) (tls.Certificate, error) {
	var c tls.Certificate

	if certFile != "" {
		b, err := ioutil.ReadFile(certFile)
		if err != nil {
			return tls.Certificate{}, err
		}

		var cert *x509.Certificate
		for cert == nil {
			var block *pem.Block
			block, b = pem.Decode(b)
			switch {
			case block == nil:
				return tls.Certificate{}, errors.New("certificate not found")
			case block.Type == "CERTIFICATE":
				cert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return tls.Certificate{}, err
				}
			}
		}
		c.Certificate = [][]byte{cert.Raw}
		c.Leaf = cert
	}

	if keyFile != "" {
		b, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return tls.Certificate{}, err
		}

		var key *rsa.PrivateKey
		for key == nil {
			var block *pem.Block
			block, b = pem.Decode(b)
			switch {
			case block == nil:
				return tls.Certificate{}, errors.New("private key not found")
			case block.Type == "RSA PRIVATE KEY":
				key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					return tls.Certificate{}, err
				}
			}
		}
		c.PrivateKey = key
	}

	return c, nil
}
