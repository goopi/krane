package main

import "bytes"
import "encoding/binary"
import "io/ioutil"
import "time"

const APPLE_PRODUCTION_GATEWAY_URI = "gateway.push.apple.com:2195"
const APPLE_PRODUCTION_FEEDBACK_URI = "feedback.push.apple.com:2196"
const APPLE_DEVELOPMENT_GATEWAY_URI = "gateway.sandbox.push.apple.com:2195"
const APPLE_DEVELOPMENT_FEEDBACK_URI = "feedback.sandbox.push.apple.com:2196"

type Client struct {
	gateway_uri  string
	feedback_uri string
	certificate  []byte
	passphrase   []byte
	timeout      time.Duration
	conn         *Connection
	res          chan interface{}
	quit         chan bool
}

func NewClient(sandbox bool, certificate string, passphrase []byte) *Client {
	c := &Client{
		passphrase: passphrase,
		timeout:    5 * time.Second,
		res:        make(chan interface{}, 50),
		quit:       make(chan bool),
	}

	if sandbox {
		c.gateway_uri = APPLE_DEVELOPMENT_GATEWAY_URI
		c.feedback_uri = APPLE_DEVELOPMENT_FEEDBACK_URI
	} else {
		c.gateway_uri = APPLE_PRODUCTION_GATEWAY_URI
		c.feedback_uri = APPLE_PRODUCTION_FEEDBACK_URI
	}

	if dat, err := ioutil.ReadFile(certificate); err == nil {
		c.certificate = dat
	}

	return c
}

// TODO: send multiple
func (c *Client) Push(n *Notification) error {
	c.conn = NewConnection(c.gateway_uri, c.certificate, c.passphrase)

	err := c.conn.Open()
	if err != nil {
		return err
	}

	defer c.conn.Close()

	payload, err := n.ToBinary()
	if err != nil {
		return err
	}

	err = c.conn.Write(payload)
	if err != nil {
		return err
	}

	// TODO: checking error response

	return nil
}

func (c *Client) UnregisteredDevices() (devices []string, err error) {
	c.conn = NewConnection(c.feedback_uri, c.certificate, c.passphrase)

	err = c.conn.Open()
	if err != nil {
		return
	}

	defer c.conn.Close()

	// set feedback timeout
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))

	go c.feedbackLoop()

	for closed := false; closed != true; {
		select {
		case r := <-c.res:
			if d, ok := r.(*Device); ok {
				devices = append(devices, string(d.ToJSON()))
			}
		case <-c.quit:
			closed = true
		}
	}

	return
}

func (c *Client) feedbackLoop() {
	// Binary format of a feedback tuple (38 bytes)
	b := make([]byte, 38, 38)

	// Timestamp (4 bytes)
	// A timestamp indicating when APNs determined that the app
	// no longer exists on the device.
	var timestamp uint32

	// Token length (2 bytes)
	// The length of the device token as a two-byte integer.
	var tokenLength uint16

	// Device token (32 bytes)
	// The device token in binary format.
	deviceToken := make([]byte, 32, 32)

	for {
		err := c.conn.Read(b)
		if err != nil {
			c.quit <- true
			return
		}

		buf := bytes.NewReader(b)
		binary.Read(buf, binary.BigEndian, &timestamp)
		binary.Read(buf, binary.BigEndian, &tokenLength)
		binary.Read(buf, binary.BigEndian, &deviceToken)

		c.res <- NewDevice(deviceToken, timestamp)
	}
}
