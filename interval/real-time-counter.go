package main

import (
	"fmt"
	"time"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)


type RealTimeCounter struct {
	session				*mgo.Session
	rabbitChan		*amqp.Channel
	updateChannel	chan string
}

type ZoneOccupancy struct {
	ZoneId		string
	Occupancy	int
}

func InitRealTimeCounter(mongoAddress string, consulAddress string) *RealTimeCounter {
	var realTimeCounter *RealTimeCounter = new(RealTimeCounter)

	fmt.Printf("Connecting to MongoDB at: %v\n", mongoAddress)
	session, err := mgo.Dial(mongoAddress)
	failOnError(err, "Failed to connect to MongoDB\n")
	session.SetMode(mgo.Monotonic, true)

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

	go realTimeCounter.handleZoneChanges()
	return realTimeCounter
}

func (realTimeCounter *RealTimeCounter) ZoneChanged(zoneId string) {
	realTimeCounter.updateChannel <- zoneId
}

func (realTimeCounter *RealTimeCounter) handleZoneChanges() {
	for {
		zoneId := <- realTimeCounter.updateChannel
		occupancy := realTimeCounter.getOccupancy(zoneId)

		// TODO Publish event
		fmt.Println(occupancy)
	}
}

func (realTimeCounter *RealTimeCounter) getOccupancy(zoneId string) *ZoneOccupancy {
	currentTimestamp := time.Now()
	c := realTimeCounter.session.DB("store").C("intervals")
	occupancy, err := c.Find(bson.M{"zoneId": zoneId, "from": bson.M{"$lte": currentTimestamp}, "$or": []bson.M{bson.M{"to": bson.M{"$gte": currentTimestamp}}, bson.M{"to": bson.M{"$exists": false}}}}).Count()
	
	if err != nil {
		panic(err)
	}
	
	zoneOccupancy := new(ZoneOccupancy)
	zoneOccupancy.Occupancy = occupancy
	zoneOccupancy.ZoneId = zoneId

	return zoneOccupancy
}
