package config

import (
	"flag"
	"time"
)

var Broker = flag.String("broker", "tcp://127.0.0.1:1883", "The broker url.")
var Adapter = flag.String("adapter", "hci0", "the bluetooth adapter")
var Interval = flag.Duration("iterval", 90 * time.Minute, "the time between starting new scans")

func init() {
	flag.Parse()
}
