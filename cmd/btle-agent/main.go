package main

import (
	"btle-agent/pkg/hcitool"
	"flag"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	broker := flag.String("broker", "tcp://127.0.0.1:1883", "The broker url.")
	//adapter := flag.String("adapter", "hci0", "the bluetooth adapter")
	flag.Parse()

	opts := mqtt.NewClientOptions()
	opts.ClientID = "btle-agent"
	opts.AutoReconnect = true
	opts.AddBroker(*broker)
	client := mqtt.NewClient(opts)
	log.Printf("connecting to %s...", opts.Servers[0])
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("connected")
	}
	defer client.Disconnect(250)

	client.Subscribe("agents/btle/scan", 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("%s: %s", message.Topic(), string(message.Payload()))
		err := hcitool.Scan(client)
		log.Println("done?")
		if err != nil {
			log.Printf("SCAN ERROR: %s", err)
		}
	})

	// wait for kill signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
