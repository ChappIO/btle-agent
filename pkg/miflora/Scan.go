package miflora

import (
	"btle-agent/pkg/config"
	"fmt"
	mifloraC "github.com/barnybug/miflora"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"sync"
	"time"
)

var lock = sync.Mutex{}

func Init(client mqtt.Client) error {
	token := client.Subscribe("agents/btle/miflora", 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("%s: %s", message.Topic(), string(message.Payload()))
		err := scan(client, *config.Adapter)
		if err != nil {
			log.Printf("MIFLORA ERROR: %s", err)
		}
	})
	if token.Error() != nil {
		return token.Error()
	}
	go func() {
		for client.IsConnected() {
			err := scan(client, *config.Adapter)
			if err != nil {
				log.Printf("MIFLORA ERROR: %s", err)
			}
			time.Sleep(*config.Interval)
		}
	}()
	return nil
}

func readFirmware(dev *mifloraC.Miflora) (out mifloraC.Firmware, err error) {
	for i := 0; i < 3; i++ {
		out, err = dev.ReadFirmware()
		if err == nil {
			return
		}
		time.Sleep(3 * time.Second)
	}
	return
}

func readSensors(dev *mifloraC.Miflora) (out mifloraC.Sensors, err error) {
	for i := 0; i < 3; i++ {
		out, err = dev.ReadSensors()
		if err == nil {
			return
		}
		time.Sleep(3 * time.Second)
	}
	return
}

func scan(client mqtt.Client, adapter string) error {
	lock.Lock()
	defer lock.Unlock()

	addresses := strings.Split(*config.MifloraAddress, ",")
	for i, address := range addresses {
		addresses[i] = strings.ToUpper(strings.TrimSpace(address))
	}

	log.Printf("scanning %d sensors", len(addresses))

	// process them
	for _ ,address := range addresses {
		log.Printf("fetching firmware info for %s...", address)
		id := strings.ToLower(strings.ReplaceAll(address, ":", ""))
		dev := mifloraC.NewMiflora(address, adapter)
		firmware, err := readFirmware(dev)
		if err != nil {
			log.Printf("error on %s: %s", address, err)
			continue
		}
		log.Printf("fetching sensor data for %s...", address)
		sensors, err := readSensors(dev)
		if err != nil {
			log.Printf("error on %s: %s", address, err)
			continue
		}

		topic := "miflora/" + id

		client.Publish(topic+"/$firmware", 0, true, firmware.Version)
		client.Publish(topic+"/$timestamp", 0, true, fmt.Sprintf("%d", time.Now().UnixNano() / 1000000))
		client.Publish(topic+"/battery", 0, true, fmt.Sprintf("%d", firmware.Battery))
		client.Publish(topic+"/temperature", 0, true, fmt.Sprintf("%f", sensors.Temperature))
		client.Publish(topic+"/conductivity", 0, true, fmt.Sprintf("%d", sensors.Conductivity))
		client.Publish(topic+"/luminosity", 0, true, fmt.Sprintf("%d", sensors.Light))
		client.Publish(topic+"/moisture", 0, true, fmt.Sprintf("%d", sensors.Moisture))
		log.Printf("updated %s", address)
	}

	return nil
}
