# Krane

Send Apple Push Notifications from the terminal.

## Installation

```
$ go get github.com/goopi/krane
```

## Usage

Send a push

```
$ krane push <token> -c /path/to/certificate.pem -a 'Hello, World!'
```

Query the feedback service

```
$ krane feedback -c /path/to/certificate.pem
```

## License

MIT
