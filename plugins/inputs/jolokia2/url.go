package jolokia2

import (
	"time"
)

type URLConfig struct {
	Name string
	URL  string
	Username  string
	Password  string
	ResponseTimeout time.Duration

	SSLCA              string `toml:"ssl_ca"`
	SSLCert            string `toml:"ssl_cert"`
	SSLKey             string `toml:"ssl_key"`
	InsecureSkipVerify bool
}
