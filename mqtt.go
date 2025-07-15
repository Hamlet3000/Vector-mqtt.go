package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type RobotData struct {
	Name    string  `json:"name"`
	Voltage float64 `json:"voltage"`
	Docked  bool    `json:"docked"`
	Time    string  `json:"time"`
}

type Payload struct {
	Robots []RobotData `json:"robots"`
}

func main() {
	flag.Parse()
	serials := flag.Args()
	if len(serials) == 0 {
		log.Fatal("Bitte mindestens eine Seriennummer angeben")
	}

	// MQTT-Client konfigurieren
	client := configMqtt("tcp://homeassistant:1883", "TODO_USERNAME", "TODO_PASSWORD")

	// Für jede Seriennummer eine eigene Goroutine starten
	for _, serial := range serials {
		go monitorRobot(serial, client)
	}

	// Hauptprozess blockieren, damit Goroutinen weiterlaufen
	select {}
}

// Überwacht einen Roboter und veröffentlicht regelmäßig MQTT-Daten
func monitorRobot(serial string, client mqtt.Client) {
	robot, err := vector.NewEP(serial)
	if err != nil {
		log.Printf("Fehler beim Verbinden mit Roboter %s: %v", serial, err)
		return
	}

	for {
		batteryState, err := robot.Conn.BatteryState(
			context.Background(),
			&vectorpb.BatteryStateRequest{},
		)
		if err != nil {
			log.Printf("Fehler beim Abfragen des Akkustands von %s: %v", serial, err)
			continue
		}

		name := getName4Serial(serial)
		t := time.Now()
		aktTime := t.Format("2006-01-02T15:04:05.000-0700")
		voltage := batteryState.BatteryVolts
		docked := batteryState.IsOnChargerPlatform

		robotData := RobotData{
			Name:    name,
			Voltage: float64(voltage),
			Docked:  docked,
			Time:    aktTime,
		}

		publishMqtt(client, name, robotData)

		time.Sleep(10 * time.Second)
	}
}

// Gibt den Gerätenamen anhand der Seriennummer zurück
func getName4Serial(serial string) string {
	switch serial {
	case "0dd1d1d6":
		return "Vector-G5D7"
	case "004043e9":
		return "Vector-F2U7"
	default:
		return "Vector-" + serial
	}
}

// Konfiguriert den MQTT-Client
func configMqtt(brokerAddress, username, password string) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(brokerAddress)
	opts.SetClientID("Vector")
	opts.SetUsername(username)
	opts.SetPassword(password)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Fehler beim Aufbau der MQTT-Verbindung: %v", token.Error())
	}
	return client
}

// Veröffentlicht Roboter-Daten auf MQTT
func publishMqtt(client mqtt.Client, name string, data RobotData) {
	base := name + "/"

	fmt.Println("MQTT senden für:", name)
	fmt.Println(base+"voltage =", data.Voltage)
	fmt.Println(base+"docked  =", data.Docked)
	fmt.Println(base+"time    =", data.Time)

	client.Publish(base+"voltage", 0, true, fmt.Sprintf("%.2f", data.Voltage)).Wait()
	client.Publish(base+"docked", 0, true, fmt.Sprintf("%v", data.Docked)).Wait()
	client.Publish(base+"time", 0, true, data.Time).Wait()
}

