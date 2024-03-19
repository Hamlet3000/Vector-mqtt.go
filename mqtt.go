package main

import (
  "context"
  "flag"
  "fmt"
  "log"
  "time"
  "github.com/fforchino/vector-go-sdk/pkg/vector"
  "github.com/fforchino/vector-go-sdk/pkg/vectorpb"
  "github.com/eclipse/paho.mqtt.golang"
)

func main() {
  var serial = flag.String("serial", "", "Vector's Serial Number")
  flag.Parse()

  // configure mqtt client
  client := configMqtt()

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
    battery := voltageToPercent(voltage)

    // get values into JSON format
    message := fmt.Sprintf("{'robots': [{'name': '%s', 'voltage': %f, 'battery': %d, 'docked': %t, 'time': '%s'}]}", name, voltage, battery, docked, aktTime)

    // send values to publish
    publishMqtt(client, name, message)

    // delay
    time.Sleep(10 * time.Second)
  }
}

// get name vor serialnumber
func getName4Serial(serial string) string {

  name := "Vector-"+serial
  return name
}

// get batterystate in percent from voltage
func voltageToPercent(value float64) int {
    maxValue := 4.37
    minValue := 3.56

    if value >= maxValue {
        return 100
    } else if value <= minValue {
        return 0
    } else {
        percent := ((value - minValue) / (maxValue - minValue)) * 100
        return int(percent + 0.5) // Rundung zur nächsten ganzen Zahl
    }
}

// configure MQTT-Client
func configMqtt() mqtt.Client {

  // config MQTT-Client
  opts := mqtt.NewClientOptions().AddBroker("tcp://BROKER_IP:1883")
  opts.SetClientID("Vector")

  // connect to MQTT-Broker
  client := mqtt.NewClient(opts)
  if token := client.Connect(); token.Wait() && token.Error() != nil {
    log.Fatalf("Error when establishing connection to MQTT-Broker: %v", token.Error())
  }
  return client
}

// Publish values to MQTT-Broker
func publishMqtt(client mqtt.Client, topic string, message string) {

  // publish
  //fmt.Println(message)
  if token := client.Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
    log.Printf("Error publishing battery state: %v", token.Error())
  }
}
