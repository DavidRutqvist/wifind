package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type RealTimeCounter struct {
	session       *mgo.Session
	rabbitChan    *amqp.Channel
	updateChannel chan string
	exchangeTopic string
}

type ZoneOccupancy struct {
	ZoneId    string `json:"zoneID"`
	Occupancy int    `json:"occupancy"`
}

func InitRealTimeCounter(mongoAddress string, consulAddress string, exchangeTopic string) *RealTimeCounter {
	var realTimeCounter *RealTimeCounter = new(RealTimeCounter)

	fmt.Printf("Connecting to MongoDB at: %v\n", mongoAddress)
	session, err := mgo.Dial(mongoAddress)
	failOnError(err, "Failed to connect to MongoDB\n")
	session.SetMode(mgo.Monotonic, true)
	realTimeCounter.session = session

	fmt.Printf("Connecting to Consul at: %v\n", consulAddress)
	config := api.DefaultConfig()
	config.Address = consulAddress
	consul, err := api.NewClient(config)
	failOnError(err, "Failed to connect to Consul\n")

	fmt.Printf("Retrieving RabbitMQ service\n")
	catalog := consul.Catalog()
	services, _, err := catalog.Service("rabbit", "", nil)
	failOnError(err, "Failed to retrieve RabbitMQ from Consul\n")

	service := services[0]
	rabbitAddress := service.Address
	rabbitPort := service.ServicePort
	rabbitEndpoint := fmt.Sprintf("amqp://" + rabbitAddress)
	fmt.Printf("Connecting to RabbitMQ at: %v:%v\n", rabbitAddress, rabbitPort)
	connection, err := amqp.Dial(rabbitEndpoint)
	failOnError(err, "Failed to connect to RabbitMQ\n")

	fmt.Printf("Opening RabbitMQ channel\n")
	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel\n")

	realTimeCounter.rabbitChan = channel
	realTimeCounter.updateChannel = make(chan string)
	realTimeCounter.exchangeTopic = exchangeTopic

	go realTimeCounter.handleZoneChanges()
	return realTimeCounter
}

func (realTimeCounter *RealTimeCounter) ZoneChanged(zoneId string) {
	realTimeCounter.updateChannel <- zoneId
}

func (realTimeCounter *RealTimeCounter) handleZoneChanges() {
	for {
		zoneId := <-realTimeCounter.updateChannel
		occupancy := realTimeCounter.getOccupancy(zoneId)

		// TODO Publish event
		event := createZoneOccupancyEvent(occupancy, "UPDATED")

		err := realTimeCounter.rabbitChan.Publish(
			realTimeCounter.exchangeTopic, // exchange
			event.Topic,                   // topic
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        event.Body,
			})
		failOnError(err, "Failed to publish event.")

		fmt.Println(occupancy)
	}
}

func (realTimeCounter *RealTimeCounter) getOccupancy(zoneId string) *ZoneOccupancy {
	currentTimestamp := time.Now().UTC()
	expireTimestamp := currentTimestamp.Add(time.Duration(-5) * time.Minute)
	devices := make([]string, 0)
	c := realTimeCounter.session.DB("store").C("intervals")
	err := c.Find(bson.M{"zoneId": zoneId, "from": bson.M{"$lte": currentTimestamp}, "to": bson.M{"$gte": expireTimestamp}}).Distinct("deviceId", &devices)

	if err != nil {
		panic(err)
	}

	zoneOccupancy := new(ZoneOccupancy)
	zoneOccupancy.Occupancy = len(devices)
	zoneOccupancy.ZoneId = zoneId

	return zoneOccupancy
}

func createZoneOccupancyEvent(zoneOccupancy *ZoneOccupancy, eventType string) Event {
	var event Event

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(zoneOccupancy)

	body := buffer.Bytes()
	topic := "OCCUPANCY." + zoneOccupancy.ZoneId + "." + eventType

	event.Body = body
	event.Topic = topic
	return event
}
