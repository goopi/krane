package main

import "github.com/docopt/docopt-go"
import "github.com/goopi/krane/apns"
import "github.com/howeyc/gopass"
import "fmt"
import "os"
import "strconv"

var Version = "0.0.2"

const Usage = `
  Usage:
    krane push TOKEN -c CERTIFICATE [-abs] [-d] [-p]
    krane feedback -c CERTIFICATE [-d] [-p]
    krane -h | --help
    krane -v | --version

  Sends an Apple Push Notification.

  Options:
    -a, --alert ALERT               Body of the alert to send.
    -b, --badge NUMBER              Badge number.
    -s, --sound SOUND               Sound to play.
    -d, --develop                   Sandbox environment.
    -c, --certificate CERTIFICATE   Path to certificate (.pem) file.
    -p, --passphrase                Certificate passphrase.
    -h, --help                      This message.
    -v, --version                   Output version.

`

func main() {
	args, _ := docopt.Parse(Usage, nil, true, Version, false)

	if args["push"].(bool) {
		token := args["TOKEN"].(string)
		cert := args["--certificate"].(string)
		sandbox := args["--develop"].(bool)
		passphrase := args["--passphrase"].(bool)

		alert, ok := args["--alert"].(string)
		if !ok {
			exitWithError("Enter your alert message")
		}

		if _, err := os.Stat(cert); os.IsNotExist(err) {
			exitWithError("Could not find certificate file")
		}

		var pass []byte
		if passphrase {
			fmt.Print("Password: ")
			pass = gopass.GetPasswdMasked()
		}

		client := apns.NewClient(sandbox, cert, pass)

		payload := apns.NewPayload()
		payload.Alert = alert

		if args["--badge"] != nil {
			badge, err := strconv.Atoi(args["--badge"].(string))
			if err != nil {
				exitWithError("Invalid badge")
			}

			payload.Badge = badge
		}

		if sound, ok := args["--sound"].(string); ok {
			payload.Sound = sound
		}

		n := apns.NewNotification()
		n.DeviceToken = token
		n.AddPayload(payload)

		notifications := make([]*apns.Notification, 0)
		notifications = append(notifications, n)

		err := client.Push(notifications)

		if err != nil {
			exitWithError("Push notifications unsuccessful")
		}

		sent := 0
		unsent := 0

		for _, n := range notifications {
			if n.Sent {
				sent++;
			} else {
				unsent++;
			}
		}

		if sent > 0 {
			msg := fmt.Sprintf("%d push notifications sent successfully", sent)
			successMessage(msg)
		}

		if unsent > 0 {
			msg := fmt.Sprintf("%d push notifications unsuccessful", unsent)
			errorMessage(msg)
		}
	}

	if args["feedback"].(bool) {
		cert := args["--certificate"].(string)
		sandbox := args["--develop"].(bool)
		passphrase := args["--passphrase"].(bool)

		if _, err := os.Stat(cert); os.IsNotExist(err) {
			exitWithError("Could not find certificate file")
		}

		var pass []byte
		if passphrase {
			fmt.Print("Password: ")
			pass = gopass.GetPasswdMasked()
		}

		client := apns.NewClient(sandbox, cert, pass)

		devices, err := client.UnregisteredDevices()
		if err != nil {
			exitWithError("Error getting feedback")
		}

		if len(devices) > 0 {
			fmt.Println(devices)
		} else {
			successMessage("No feedback available")
		}
	}
}

func successMessage(msg string) {
	fmt.Printf("\x1b[32;1m%s\x1b[0m\n", msg)
}

func errorMessage(msg string) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", msg)
}

func exitWithError(msg string) {
	errorMessage(msg)
	os.Exit(1)
}
