package main

import (
	"btle-agent/pkg/config"
	"btle-agent/pkg/hcitool"
	"btle-agent/pkg/miflora"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*config.Broker)
	client := mqtt.NewClient(opts)
	log.Printf("connecting to %s...", opts.Servers[0])
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("connected")
	}
	defer client.Disconnect(250)

	if err := miflora.Init(client); err != nil {
		panic(err)
	}
	if err := hcitool.Init(client); err != nil {
		panic(err)
	}

	// wait for kill signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
