package main

import "io/ioutil"

const APPLE_PRODUCTION_GATEWAY_URI = "gateway.push.apple.com:2195"
const APPLE_PRODUCTION_FEEDBACK_URI = "feedback.push.apple.com:2196"
const APPLE_DEVELOPMENT_GATEWAY_URI = "gateway.sandbox.push.apple.com:2195"
const APPLE_DEVELOPMENT_FEEDBACK_URI = "feedback.sandbox.push.apple.com:2196"

type Client struct {
	gateway_uri string
	feedback_uri string
	certificate []byte
	passphrase []byte
	timeout float32
}

func NewClient(sandbox bool, certificate string, passphrase []byte) *Client {
	c := &Client{
		passphrase: passphrase,
		timeout: 0.5,
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
	conn := NewConnection(c.gateway_uri, c.certificate, c.passphrase)

	err := conn.Open()
	if err != nil {
		return err
	}

	defer conn.Close()

	payload, err := n.ToBinary()
	if err != nil {
		return err
	}

	err = conn.Write(payload)
	if err != nil {
		return err
	}

	// TODO: checking error response

	return nil
}
