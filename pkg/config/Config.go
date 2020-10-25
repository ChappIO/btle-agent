package config

import "flag"

var Broker = flag.String("broker", "tcp://127.0.0.1:1883", "The broker url.")
var Adapter = flag.String("adapter", "hci0", "the bluetooth adapter")

func init() {
	flag.Parse()
}
