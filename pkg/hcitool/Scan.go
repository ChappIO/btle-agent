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

func Scan(client mqtt.Client) error {
	log.Printf("Scanning for new devices for 1 minute")

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
		// stop after 1 minute
		time.Sleep(10 * time.Second)
		_ = cmd.Process.Signal(os.Interrupt)
	}()

	scan := bufio.NewScanner(out)
	log.Printf("Scan started...")
	for scan.Scan() {
		line := scan.Text()
		if line == "LE Scan..." {
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
		log.Printf("Found %s %s", dev.Name, dev.Mac)
	}

	errScan := bufio.NewScanner(errPipe)
	for errScan.Scan() {
		log.Println("ERROR: " + errScan.Text())
	}

	log.Println("scan complete")
	return nil
}
