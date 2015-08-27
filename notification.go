package main

import "bytes"
import "encoding/binary"
import "encoding/hex"
import "encoding/json"
import "errors"
import "math/rand"
import "time"

// Notification Payload
// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/ApplePushService.html#//apple_ref/doc/uid/TP40008194-CH100-SW1

// The maximum size allowed for a notification payload is 2 kilobytes.
const MaximumPayloadSize = 2048

// The "aps" dictionary
type Payload struct {
	Alert            interface{} `json:"alert,omitempty"`
	Badge            int         `json:"badge,omitempty"`
	Sound            string      `json:"sound,omitempty"`
	ContentAvailable int         `json:"content-available,omitempty"`
}

func NewPayload() *Payload {
	return new(Payload)
}

// The "alert" dictionary
type AlertDictionary struct {
	Title           string      `json:"title,omitempty"`
	Body            string      `json:"body,omitempty"`
	TitleLocKey     string      `json:"title-loc-key,omitempty"`
	TitleLocArgs    []string    `json:"title-loc-args,omitempty"`
	ActionLocKey    string      `json:"action-loc-key,omitempty"`
	LocKey          string      `json:"loc-key,omitempty"`
	LocArgs         []string    `json:"loc-args,omitempty"`
	LaunchImage     string      `json:"launch-image,omitempty"`
}
// TODO: custom notification actions

func NewAlertDictionary() *AlertDictionary {
	return new(AlertDictionary)
}

// Binary Interface and Notification Format
// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW4

type Notification struct {
	DeviceToken     string
	Payload         map[string]interface{}
	Identifier      int32
	ExpirationDate  uint32
	Priority        uint8
}

func NewNotification() (n *Notification) {
	n = new(Notification)

	src := rand.NewSource(time.Now().UnixNano())
	n.Identifier = rand.New(src).Int31()

	n.Payload = make(map[string]interface{})

	// sent immediately
	n.Priority = 10

	return
}

func (n *Notification) AddPayload(p *Payload) {
	// The badge field is optional (omitempty), but we should send it
	// if the value is 0. In order to achieve that, we set the value
	// to -1, otherwise it will be omitted.
	if p.Badge == 0 {
		p.Badge = -1
	}

	n.SetPayloadValue("aps", p)
}

func (n *Notification) SetPayloadValue(key string, value interface{}) {
	n.Payload[key] = value
}

func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n.Payload)
}

func (n *Notification) ToBinary() ([]byte, error) {
	token, err := hex.DecodeString(n.DeviceToken)
	if err != nil {
		return nil, err
	}

	payload, err := n.ToJSON()
	if err != nil {
		return nil, err
	}

	if len(payload) > MaximumPayloadSize {
		return nil, errors.New("Invalid payload size")
	}

	frameDataBuffer := n.frameDataItems(token, payload)

	buffer := bytes.NewBuffer([]byte{})

	// Command: populate with the number 2 (1 byte)
	binary.Write(buffer, binary.BigEndian, uint8(2))

	// Frame length: the size of the frame data (4 bytes)
	binary.Write(buffer, binary.BigEndian, uint32(frameDataBuffer.Len()))

	// Frame data: the frame contains the body, structured as a series of items (variable length)
	binary.Write(buffer, binary.BigEndian, frameDataBuffer.Bytes())

	return buffer.Bytes(), nil
}

func (n *Notification) frameDataItems(token, payload []byte) *bytes.Buffer {
	buffer := new(bytes.Buffer)

	// Frame data
	// The frame data is made up of a series of items. Each item is
	// made up of the following, in order:
	//    * Item ID (1 byte)
	//    * Item data length (2 bytes)
	//    * Item data (variable length)

	// #1 Device token (32 bytes)
	// The device token in binary form, as was registered by the device.
	binary.Write(buffer, binary.BigEndian, uint8(1))
	binary.Write(buffer, binary.BigEndian, uint16(32))
	binary.Write(buffer, binary.BigEndian, token)

	// #2 Payload (variable length, less than or equal to 2 kilobytes)
	// The JSON-formatted payload.
	binary.Write(buffer, binary.BigEndian, uint8(2))
	binary.Write(buffer, binary.BigEndian, uint16(len(payload)))
	binary.Write(buffer, binary.BigEndian, payload)

	// #3 Notification identifier (4 bytes)
	// An arbitrary, opaque value that identifies this notification.
	binary.Write(buffer, binary.BigEndian, uint8(3))
	binary.Write(buffer, binary.BigEndian, uint16(4))
	binary.Write(buffer, binary.BigEndian, n.Identifier)

	// #4 Expiration date (4 bytes)
	// A UNIX epoch date expressed in seconds (UTC) that identifies when
	// the notification is no longer valid and can be discarded.
	binary.Write(buffer, binary.BigEndian, uint8(4))
	binary.Write(buffer, binary.BigEndian, uint16(4))
	binary.Write(buffer, binary.BigEndian, n.ExpirationDate)

	// #5 Priority (1 byte)
	// The notification’s priority.
	binary.Write(buffer, binary.BigEndian, uint8(5))
	binary.Write(buffer, binary.BigEndian, uint16(1))
	binary.Write(buffer, binary.BigEndian, n.Priority)

	return buffer
}

// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW12
var ErrorResponseCodes = map[uint8]string{
	0:   "No errors encountered",
	1:   "Processing error",
	2:   "Missing device token",
	3:   "Missing topic",
	4:   "Missing payload",
	5:   "Invalid token size",
	6:   "Invalid topic size",
	7:   "Invalid payload size",
	8:   "Invalid token",
	10:  "Shutdown",
	255: "Unknown error",
}

func ErrorForCode(code uint8) error {
	msg, ok := ErrorResponseCodes[code]
	if !ok {
		msg = "Unknown error"
	}

	return errors.New(msg)
}
