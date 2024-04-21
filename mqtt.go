package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "time"

    "github.com/fforchino/vector-go-sdk/pkg/vector"
    "github.com/fforchino/vector-go-sdk/pkg/vectorpb"
    "github.com/eclipse/paho.mqtt.golang"
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
    var serial = flag.String("serial", "", "Vector's Serial Number")
    flag.Parse()

    // configure mqtt clients
    client := configMqtt("tcp://192.168.0.7:1883", "", "")

    // connect to robot
    robot, err := vector.NewEP(*serial)
    if err != nil {
        log.Fatal(err)
    }

    for {
        batteryState, err := robot.Conn.BatteryState(
            context.Background(),
            &vectorpb.BatteryStateRequest{},
        )
        if err != nil {
            log.Fatal(err)
        }

        name := getName4Serial(*serial)
        t := time.Now()
        aktTime := t.Format("2006-01-02T15:04:05.000-0700")
        voltage := batteryState.BatteryVolts
        docked := batteryState.IsOnChargerPlatform

        // Prepare robot data
        robotData := RobotData{
            Name:    name,
            Voltage: float64(voltage),
            Docked:  docked,
            Time:    aktTime,
        }

        // Prepare payload
        payload := Payload{
            Robots: []RobotData{robotData},
        }

        // Publish data to MQTT brokers
        publishMqtt(client, name, payload)


        // Delay
        time.Sleep(10 * time.Second)
    }
}

// get name for serial number
func getName4Serial(serial string) string {
    switch serial {
    case "0dd1d1d6":
        return "Vector-D1H8"
    case "004043e9":
        return "Vector-F2U7"
    default:
        return "Vector-" + serial
    }
}

// configure MQTT-Client
func configMqtt(brokerAddress, username, password string) mqtt.Client {
    // config MQTT-Client
    opts := mqtt.NewClientOptions().AddBroker(brokerAddress)
    opts.SetClientID("Vector")

    // Set username and password
    opts.SetUsername(username)
    opts.SetPassword(password)

    // connect to MQTT-Broker
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("Error when establishing connection to MQTT-Broker: %v", token.Error())
    }
    return client
}

// Publish values to MQTT-Broker
func publishMqtt(client mqtt.Client, topic string, data interface{}) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Printf("Error while coding JSON Data: %v", err)
        return
    }

    // Print JSON payload to console
    fmt.Println(topic, string(jsonData))

    // publish
    if token := client.Publish(topic, 0, false, jsonData); token.Wait() && token.Error() != nil {
        log.Printf("Error while publishing Data: %v", token.Error())
    }
}

