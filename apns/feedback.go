package apns

import "encoding/hex"
import "encoding/json"

type Device struct {
	Token     string `json:"token"`
	Timestamp uint32 `json:"timestamp"`
}

func NewDevice(token []byte, timestamp uint32) *Device {
	d := &Device{
		Token:     hex.EncodeToString(token),
		Timestamp: timestamp,
	}

	return d
}

func (d *Device) ToJSON() []byte {
	j, _ := json.MarshalIndent(d, "", "  ")
	return j
}
