package main

import "crypto/tls"
import "crypto/x509"
import "encoding/pem"
import "errors"
import "net"
import "time"

type Connection struct {
	uri         string
	certificate []byte
	passphrase  []byte
	conn        net.Conn
	client      *tls.Conn
}

func NewConnection(uri string, certificate []byte, passphrase []byte) *Connection {
	c := &Connection{
		uri:         uri,
		certificate: certificate,
		passphrase:  passphrase,
	}

	return c
}

func (c *Connection) newCertificate() (*tls.Certificate, error) {
	var certPEMBlock *pem.Block
	var keyPEMBlock *pem.Block
	var rest []byte

	rest = c.certificate
	for {
		certPEMBlock, rest = pem.Decode(rest)
		if certPEMBlock == nil || certPEMBlock.Type == "CERTIFICATE" {
			break
		}
	}

	if certPEMBlock == nil {
		return nil, errors.New("Failed to parse certificate data")
	}

	rest = c.certificate
	for {
		keyPEMBlock, rest = pem.Decode(rest)
		if keyPEMBlock == nil || keyPEMBlock.Type == "RSA PRIVATE KEY" {
			break
		}
	}

	if keyPEMBlock == nil {
		return nil, errors.New("Failed to parse key data")
	}

	if x509.IsEncryptedPEMBlock(keyPEMBlock) {
		der, err := x509.DecryptPEMBlock(keyPEMBlock, c.passphrase)

		if err != nil {
			return nil, errors.New("Failed to decrypt: wrong passphrase")
		}

		keyPEMBlock = &pem.Block{
			Type:    keyPEMBlock.Type,
			Headers: keyPEMBlock.Headers,
			Bytes:   der,
		}
	}

	certBytes := pem.EncodeToMemory(certPEMBlock)
	keyBytes := pem.EncodeToMemory(keyPEMBlock)

	cert, err := tls.X509KeyPair(certBytes, keyBytes)

	return &cert, err
}

func (c *Connection) Open() error {
	cert, err := c.newCertificate()
	if err != nil {
		return err
	}

	host, _, _ := net.SplitHostPort(c.uri)
	conf := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ServerName:   host,
	}

	conn, err := net.Dial("tcp", c.uri)
	if err != nil {
		return err
	}

	tlsConn := tls.Client(conn, conf)

	err = tlsConn.Handshake()
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = tlsConn

	return nil
}

func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Connection) Read(b []byte) error {
	n, err := c.client.Read(b)

	if n == 0 {
		return err
	}

	return nil
}

func (c *Connection) Write(b []byte) error {
	_, err := c.client.Write(b)
	return err
}

func (c *Connection) Close() {
	c.conn.Close()
	c.client.Close()
}
