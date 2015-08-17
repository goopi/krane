package main

import "github.com/docopt/docopt-go"
import "github.com/howeyc/gopass"
import "fmt"
import "os"
import "strconv"

var Version = "0.0.1"

const Usage = `
  Usage:
    krane push TOKEN -c CERTIFICATE [options]
    krane feedback
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

        client := NewClient(sandbox, cert, pass)

        payload := NewPayload()
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

        notification := NewNotification()
        notification.DeviceToken = token
        notification.AddPayload(payload)

        err := client.Push(notification)

        if err == nil {
            fmt.Printf("\x1b[32;1m%s\x1b[0m\n", "Push notification sent successfully")
        } else {
            exitWithError("Push notification unsuccessful")
        }

        return
    }

    // TODO: implement feedback
    if args["feedback"].(bool) {
    }
}

func exitWithError(msg string) {
    fmt.Printf("\x1b[31;1m%s\x1b[0m\n", msg)
    os.Exit(1)
}

