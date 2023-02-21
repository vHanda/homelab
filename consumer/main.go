package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"context"
	"database/sql"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
)

type MessageHandler struct {
	db *sql.DB
}

func NewMessageHandler() *MessageHandler {
	var (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "PASS"
		dbname   = "homelab"
	)

	var err error
	if val := os.Getenv("POSTGRES_HOST"); val != "" {
		host = val
	}
	if val := os.Getenv("POSTGRES_PORT"); val != "" {
		port, err = strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
	}
	if val := os.Getenv("POSTGRES_USER"); val != "" {
		user = val
	}
	if val := os.Getenv("POSTGRES_PASSWORD"); val != "" {
		password = val
	}
	if val := os.Getenv("POSTGRES_DB"); val != "" {
		dbname = val
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	return &MessageHandler{db: db}
}

func (h *MessageHandler) Close() {
	h.db.Close()
}

func (h *MessageHandler) handle(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	if topic != mqtt_topic {
		return
	}

	var sensor1Data Sensor1Data
	err := json.Unmarshal(msg.Payload(), &sensor1Data)
	if err != nil {
		log.Println(err)
		return
	}

	ctx := context.Background()
	q := &Queries{db: h.db}

	err = q.AddSensorData(ctx, AddSensorDataParams{
		Time:        time.Unix(sensor1Data.Time, 0),
		Temperature: sensor1Data.Temperature,
		Humidity:    sensor1Data.Humidity,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

type Sensor1Data struct {
	Time        int64   `json:"time"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

const mqtt_topic = "sensor1"

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Println("Received unknown message on topic: " + msg.Topic())
}
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Println("Lost connection to MQTT broker")
}

func main() {
	var err error

	message_handler := NewMessageHandler()
	log.Println("Connected to Postges")

	var mqtt_host = "localhost"
	var mqtt_port = 1883
	var mqtt_client = "go_mqtt_client"
	var mqtt_user = "go_mqttt"
	var mqtt_password = "go_mqtt_password"

	if mqttHost := os.Getenv("MQTT_HOST"); mqttHost != "" {
		mqtt_host = mqttHost
	}
	if mqttPort := os.Getenv("MQTT_PORT"); mqttPort != "" {
		mqtt_port, err = strconv.Atoi(mqttPort)
		if err != nil {
			panic(err)
		}
	}
	if mqttClient := os.Getenv("MQTT_CLIENT"); mqttClient != "" {
		mqtt_client = mqttClient
	}
	if mqttUser := os.Getenv("MQTT_USER"); mqttUser != "" {
		mqtt_user = mqttUser
	}
	if mqttPassword := os.Getenv("MQTT_PASSWORD"); mqttPassword != "" {
		mqtt_password = mqttPassword
	}

	mqtt.ERROR = log.New(os.Stdout, "[MQTT ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[MQTT CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[MQTT WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[MQTT DEBUG] ", 0)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqtt_host, mqtt_port))
	opts.SetClientID(mqtt_client)
	opts.SetUsername(mqtt_user)
	opts.SetPassword(mqtt_password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetOrderMatters(false)

	opts.ConnectRetry = true
	opts.AutoReconnect = true

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("mqtt connection established")

		t := c.Subscribe(mqtt_topic, 1, message_handler.handle)
		// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
		// in other handlers does cause problems its best to just assume we should not block
		go func() {
			_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
			if t.Error() != nil {
				fmt.Printf("ERROR SUBSCRIBING: %s\n", t.Error())
			} else {
				fmt.Println("subscribed to: ", mqtt_topic)
			}
		}()
	}
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	client.AddRoute(mqtt_topic, message_handler.handle)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig

	client.Disconnect(250)
	log.Println("Shutdown complete")
}
