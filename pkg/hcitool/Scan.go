package hcitool

import (
	"bufio"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Device struct {
	Mac  string
	Name string
}

func Init(client mqtt.Client) error {
	token := client.Subscribe("agents/btle/scan", 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("%s: %s", message.Topic(), string(message.Payload()))
		err := scan(client)
		if err != nil {
			log.Printf("SCAN ERROR: %s", err)
		}
	})
	return token.Error()
}

func scan(client mqtt.Client) error {
	log.Printf("Scanning for new devices for 30 seconds...")

	cmd := exec.Command("hcitool", "lescan")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	defer cmd.Process.Signal(os.Interrupt)

	go func() {
		time.Sleep(30 * time.Second)
		_ = cmd.Process.Signal(os.Interrupt)
	}()

	scan := bufio.NewScanner(out)
	log.Printf("Scan started...")
	for scan.Scan() {
		line := scan.Text()
		if line == "LE Scan ..." {
			continue
		}
		parts := strings.SplitN(scan.Text(), " ", 2)
		addr := parts[0]
		name := parts[1]

		if name == "(unknown)" {
			// we don't know what this is yet...
			continue
		}
		dev := Device{
			Mac:  addr,
			Name: name,
		}

		topic := fmt.Sprintf("bluetoothle/%s", strings.ReplaceAll(strings.ToLower(dev.Mac), ":", ""))
		client.Publish(topic+"/$name", 2, true, dev.Name)
		client.Publish(topic+"/$timestamp", 2, true, fmt.Sprintf("%d", time.Now().Unix()))
		client.Publish(topic+"/$mac", 2, true, dev.Mac)
		log.Printf("found and reported %s %s", dev.Name, dev.Mac)
	}

	errScan := bufio.NewScanner(errPipe)
	for errScan.Scan() {
		log.Println("ERROR: " + errScan.Text())
	}

	log.Println("scan complete")
	return nil
}
