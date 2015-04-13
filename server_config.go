package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sevlyar/go-daemon"
)

var (
	DefaultServerPath = ""
	DefaultStorePath  = ""
	DefaultLogPath    = ""

	DefaultPublicKeyPath  = ""
	DefaultPrivateKeyPath = ""

	SSL = false
)

type ServerConfig struct {
	StorePath  string
	Path       string
	LogPath    string
	PublicKey  string
	PrivateKey string
	Cert       tls.Certificate
	Daemon     *daemon.Context
}

func NewServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{}

	p, err := HomeDir()

	if err != nil {
		return config, err
	}

	Log.Debugf("Home path set to %s.", p)

	DefaultServerPath = path.Join(p, "."+Name)
	DefaultStorePath = path.Join(DefaultServerPath, "store.db")
	DefaultLogPath = path.Join(DefaultServerPath, "logs")

	DefaultPublicKeyPath = path.Join(DefaultServerPath, "envdb.cert.pem")
	DefaultPrivateKeyPath = path.Join(DefaultServerPath, "envdb.key.pem")

	Log.Debugf("Default Server Config Path: %s.", DefaultServerPath)
	Log.Debugf("Default Server Store Path: %s.", DefaultStorePath)
	Log.Debugf("Default Server Log Path: %s.", DefaultLogPath)

	if err := os.MkdirAll(DefaultServerPath, 0777); err != nil {
		return config, err
	}

	if err := os.MkdirAll(DefaultLogPath, 0777); err != nil {
		return config, err
	}

	config.Path = DefaultServerPath
	config.StorePath = DefaultStorePath
	config.LogPath = DefaultLogPath

	if !IsExist(DefaultPrivateKeyPath) || !IsExist(DefaultPublicKeyPath) {
		err := NewKeyPair()

		if err != nil {
			os.Exit(-1)
		}
	}

	cert, err := tls.LoadX509KeyPair(DefaultPublicKeyPath, DefaultPrivateKeyPath)

	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}

	config.Cert = cert

	config.Daemon = &daemon.Context{
		PidFileName: fmt.Sprintf("%s/%s-server.pid", DefaultServerPath, Name),
		PidFilePerm: 0644,
		LogFileName: fmt.Sprintf("%s/%s-server.log", DefaultLogPath, Name),
		LogFilePerm: 0640,
		WorkDir:     DefaultServerPath,
		Umask:       027,
	}

	return config, err
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func NewKeyPair() error {
	var priv interface{}
	var err error

	priv, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		Log.Fatalf("failed to generate private key: %s", err)
		return err
	}

	var notBefore time.Time
	notBefore = time.Now()

	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Envdb"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split("localhost", ",")

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
		return err
	}

	certOut, err := os.Create(DefaultPublicKeyPath)

	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", DefaultPublicKeyPath, err)
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	Log.Debug("%s saved successfully.", DefaultPublicKeyPath)

	keyOut, err := os.OpenFile(DefaultPrivateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		Log.Errorf("failed to open %s for writing:", DefaultPrivateKeyPath, err)
		return err
	}

	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()

	Log.Debugf("%s saved successfully.", DefaultPrivateKeyPath)

	return nil
}
