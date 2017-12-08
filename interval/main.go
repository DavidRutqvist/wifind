package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"io/ioutil"
	"strconv"
)

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}
func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf(msg)
		panic(err)
	}
}

type PostRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
type SensorLocation struct {
	SensorId string    `json:"id" bson:"_id,omitempty"`
	Zoneid   string    `json:"zoneid" bson:"zoneid"`
	From     time.Time `json:"from" bson:"from"`
	To       time.Time `json:"to" bson:"to"`
}

type Interval struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Deviceid string        `bson:"deviceId" json:"deviceId"`
	From     time.Time     `bson:"from" json:"from"`
	To       time.Time     `bson:"to" json:"to"`
	Zone     string 	   `bson:"zoneId" json:"zoneId"`
}
type Datastore struct {
	Device string `json:"device" bson:"device"`
	Sensor string `json:"sensor" bson:"sensor"`
	Rssi   int32  `json:"rssi" bson:"rssi"`
	Time   int64  `json:"time" bson:"time"`
}
type Instances struct {
	Session           *mgo.Session
	Consul            *api.Client
	RabbitConn        *amqp.Connection
	RabbitChan        *amqp.Channel
	ExchangeTopic     string
	RabbitEndpoint    string
	SubscriptionTopic string
	EventChannel 	  chan Event	
}

type Event struct {
	body []byte
	topic string
}

func createInstances(mongoAddress string, consulAddress string, exchangedTopic string, SubscriptionTopic string) *Instances {
	var instances *Instances = new(Instances)

	fmt.Printf("Connecting to MongoDB at: %v\n", mongoAddress)
	session, err := mgo.Dial(mongoAddress)
	failOnError(err, "Failed to connect to MongoDB\n")
	session.SetMode(mgo.Monotonic, true)
	instances.Session = session

	fmt.Printf("Connecting to Consul at: %v\n", consulAddress)
	config := api.DefaultConfig()
	config.Address = consulAddress
	consul, err := api.NewClient(config)
	failOnError(err, "Failed to connect to Consul\n")
	instances.Consul = consul

	fmt.Printf("Retrieving RabbitMQ service\n")
	catalog := consul.Catalog()
	services, _, err := catalog.Service("rabbit", "", nil)
	failOnError(err, "Failed to retrieve RabbitMQ from Consul\n")

	service := services[0]
	rabbitAddress := service.Address
	rabbitPort := service.ServicePort
	rabbitEndpoint := fmt.Sprintf("amqp://" + rabbitAddress)
	instances.RabbitEndpoint = rabbitEndpoint
	fmt.Printf("Connecting to RabbitMQ at: %v:%v\n", rabbitAddress, rabbitPort)
	connection, err := amqp.Dial(rabbitEndpoint)
	failOnError(err, "Failed to connect to RabbitMQ\n")

	fmt.Printf("Opening RabbitMQ channel\n")
	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel\n")

	instances.RabbitConn = connection
	instances.RabbitChan = channel
	instances.ExchangeTopic = exchangedTopic
	instances.SubscriptionTopic = SubscriptionTopic
	instances.EventChannel = make(chan Event)

	return instances
}
func CreateInterval(sensorlocation SensorLocation, datastore Datastore) *Interval{
	var interval Interval
	interval.Id = bson.NewObjectId()
	interval.Zone = sensorlocation.Zoneid
	interval.Deviceid = datastore.Device
	interval.From = time.Unix(datastore.Time, 0)
	interval.To = time.Unix(datastore.Time, 0)
	return &interval
}
func main() {
	mongoAddress := os.Getenv("MONGO_ADDRESS")
	//consulAddress := os.Getenv("CONSUL_ADDRESS")
	consulAddress := "http://srv.wifind.se:8500"
	//exchangedTopic := os.Getenv("EXCHANGE_TOPIC")
	exchangedTopic := "event"

	//instances := createInstances(mongoAddress, "srv.wifind.se:8500")
	instances := createInstances(mongoAddress, consulAddress, exchangedTopic, "sensor.*.detected.*")

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/intervals"), allIntervals(instances))
	mux.HandleFunc(pat.Get("/intervals/:at"), getIntervalByTime(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zoneid"), getIntervalByZoneName(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zoneid/:from/:to"), getIntervalByZoneNameDuringInterval(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zoneid/:at"), getIntervalForZoneByTime(instances))
	//mux.HandleFunc(pat.Get("/intervals/device/:deviceid"), getIntervalByDeviceID(instances))
	//mux.HandleFunc(pat.Get("/intervals/device/:zoneid"), getIntervalForDeviceInZone(instances))
	mux.HandleFunc(pat.Get("/"), healthCheck(instances))
	fmt.Printf("Starting Router\n")
	go instances.Recieve()
	go instances.eventBroadcaster()
	http.ListenAndServe("0.0.0.0:8080", mux)

}
func (i *Instances) eventBroadcaster() {
	connection, err := amqp.Dial(i.RabbitEndpoint)
	failOnError(err, "Failed to open a connection\n")
	defer connection.Close()

	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel\n")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		i.ExchangeTopic, 	// name
		"topic",			// topic
		true,				// durable
		false,				// auto-deleted
		false,				// internal
		false,				// no-wait
		nil,				// arguments
	)
	failOnError(err, "Failed to declare an exchange.")
	for event <- i.EventChannel {
		err = channel.Publish(
			i.ExchangeTopic,	// exchange
			event.Topic,				// topic
			false,
			false,
			amqp.Publishing{
				ContentType:	"application/json",
				Body:			event.Body,
			})
	failOnError(err, "Failed to publish event.")	
	}
	

} 

func healthCheck(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		buffer := new(bytes.Buffer)

		response := PostRes{Success: true, Message: "I am alive"}
		json.NewEncoder(buffer).Encode(response)
		respBody := buffer.Bytes()
		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func allIntervals(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		session := i.Session.Copy()
		defer session.Close()

		c := session.DB("store").C("intervals")

		intervals := make([]Interval, 0)
		err = c.Find(bson.M{}).All(&intervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		respBody, err = json.MarshalIndent(intervals, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalByTime(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		at := (pat.Param(r, "at"))
		i, err := strconv.ParseInt(at, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)

		c := session.DB("store").C("intervals")
		var intervalsWithin = make([]Interval, 0)
		err = c.Find(bson.M{"from": bson.M{"$lte": tm}, "$or": []bson.M{bson.M{"to": bson.M{"$gte": tm}}, bson.M{"to": bson.M{"$exists": false}}}}).All(&intervalsWithin)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		respBody, err = json.MarshalIndent(intervalsWithin, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalByZoneName(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		zoneid := bson.ObjectIdHex(pat.Param(r, "zoneid"))

		c := session.DB("store").C("intervals")

		intervals := make([]Interval, 0)
		err = c.Find(bson.M{"zoneId": zoneid}).All(&intervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		respBody, err = json.MarshalIndent(intervals, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalByDeviceID(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		deviceid := pat.Param(r, "deviceid")

		c := session.DB("store").C("intervals")

		var intervals []Interval
		err = c.Find(bson.M{"deviceId": deviceid}).All(&intervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		if intervals[0].Deviceid == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		respBody, err = json.MarshalIndent(intervals, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalForDeviceInZone(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		zoneid := (pat.Param(r, "zoneid"))
		deviceid := pat.Param(r, "deviceid")

		c := session.DB("store").C("intervals")

		var intervals []Interval
		err = c.Find(bson.M{"zoneId": zoneid, "deviceId": deviceid}).All(&intervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		if intervals[0].Zone == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		respBody, err = json.MarshalIndent(intervals, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalByZoneNameDuringInterval(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		from := (pat.Param(r, "from"))
		i, err := strconv.ParseInt(from, 10, 64)
		if err != nil {
			panic(err)
		}
		tmfrom := time.Unix(i, 0)
		to := (pat.Param(r, "to"))
		y, err := strconv.ParseInt(to, 10, 64)
		if err != nil {
			panic(err)
		}
		tmto := time.Unix(y, 0)
		zoneid := (pat.Param(r, "zoneid"))

		c := session.DB("store").C("intervals")
		var timeIntervals []Interval
		var intervalsWithin []Interval
		err = c.Find(bson.M{"zoneId": zoneid}).All(&timeIntervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(timeIntervals); i++ {
			if timeIntervals[i].From.Before(tmto) && timeIntervals[i].To.After(tmfrom) {
				intervalsWithin = append(intervalsWithin, timeIntervals[i])
			}
		}
		if intervalsWithin[0].Zone == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
		respBody, err = json.MarshalIndent(intervalsWithin, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func getIntervalForZoneByTime(i *Instances) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

		session := i.Session.Copy()
		defer session.Close()

		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		at := (pat.Param(r, "at"))
		i, err := strconv.ParseInt(at, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		zoneid := (pat.Param(r, "zoneid"))

		c := session.DB("store").C("intervals")
		var timeIntervals []Interval
		var intervalsWithin []Interval
		err = c.Find(bson.M{"zoneId": zoneid}).All(&timeIntervals)
		if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(timeIntervals); i++ {
			if timeIntervals[i].From.Before(tm) && timeIntervals[i].To.After(tm) {
				intervalsWithin = append(intervalsWithin, timeIntervals[i])
			}
		}
		if intervalsWithin[0].Zone == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
		respBody, err = json.MarshalIndent(intervalsWithin, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
func (i *Instances) Update (datastore Datastore, sensorlocation SensorLocation) *Interval {
	session := i.Session.Copy()
		defer session.Close()
	var interval *Interval
	c := session.DB("store").C("intervals")

	if (&sensorlocation != nil) && time.Unix(datastore.Time, 0).Before(sensorlocation.From) && (&(sensorlocation.To) == nil || time.Unix(datastore.Time, 0).After(sensorlocation.To)) {
		err := c.Find(bson.M{"deviceId": datastore.Device}).One(&interval) //hämta senaste intervall för (mobil) enhet
		if err != nil{
			fmt.Println("NO INTERVAL")
			fmt.Println(err)
			//failOnError(err, "Cant find interval\n")
			interval = CreateInterval(sensorlocation, datastore) //finns inget - skapa nytt
			err = c.Insert(interval)
			failOnError(err, "Failed to insert interval\n")
		} else {
			duration := time.Since(interval.To)
			if interval.Zone != sensorlocation.Zoneid{ //senaste intervall fel zon - skapa nytt
				fmt.Println("Different zone")
				interval = CreateInterval(sensorlocation, datastore)
				err = c.Insert(interval)
				failOnError(err, "Failed to insert interval time\n")
			} else if duration.Minutes() > 5{ //senaste intervall to värdet för länge sedan - skapa nytt
				fmt.Println("Old interval")
				interval = CreateInterval(sensorlocation, datastore)
				err = c.Insert(interval)
				failOnError(err, "Failed to insert interval time\n")
			} else { // inom 5 min, uppdatera
				fmt.Println("----- UPDATE -----")
				interval.To = time.Unix(datastore.Time, 0) // updatera to värdet till nu
				err = c.Update(bson.M{"_id": interval.Id}, interval)
				failOnError(err, "Failed to update interval\n")
				topic := "intervals." + bson.ObjectIdHex(interval.Id) + ".updated"
				
			}
		}
	} else {
		fmt.Println("No zone mapping")
	}
	return interval
}
func (i *Instances) Recieve() {
	conn, err := amqp.Dial(i.RabbitEndpoint)
	failOnError(err, "Failed to connect to RabbitMQ\n")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel\n")
	defer channel.Close()

	session := i.Session.Copy()
	defer session.Close()

	//var interval Interval'

	err = channel.ExchangeDeclare(
		i.ExchangeTopic, // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare an exchange\n")

	q, err := channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue\n")

	fmt.Printf("queue name: %v\nroutingkey: %v\nexchange: %v\n", q.Name, i.SubscriptionTopic, i.ExchangeTopic)

	err = channel.QueueBind(
		q.Name,              // queue name
		i.SubscriptionTopic, // routing key
		i.ExchangeTopic,     // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue\n")

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		true,   // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer\n")

	fmt.Printf("Retrieving sensorlocation service\n")
	catalog := i.Consul.Catalog()
	services, _, err := catalog.Service("sensorlocation", "", nil)
	failOnError(err, "Failed to retrive sensorlocation from Consul\n")

	service := "http://" + services[0].ServiceAddress + ":" + strconv.Itoa(services[0].ServicePort)

	fmt.Println(service)

	forever := make(chan bool)
	go func() {
		var datastore Datastore
		var sensorlocation SensorLocation

		for d := range msgs {

			err := json.Unmarshal(d.Body, &datastore)
			failOnError(err, "Failed to unmarshal event.\n")

			

			resp, err := http.Get(service + "/sensors/" + datastore.Sensor)
			failOnError(err, "Failed to get sensorlocation.\n")
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&sensorlocation)
			failOnError(err, "didn't get data\n")

			_ = i.Update(datastore, sensorlocation)
			failOnError(err, "Interval update fail")
			/*fmt.Printf("%s\n", i.Zone)
			fmt.Printf("%s\n", i.Deviceid)
			fmt.Printf("%s\n", i.From)
			fmt.Printf("%v\n", i.To)
			fmt.Printf("%v\n", i.Rssi)
//*/
			/*
				hämta senaste intervall för (mobil) enhet

				finns inget - skapa nytt

				senaste intervall fel zon - skapa nytt

				senaste intervall to värdet för länge sedan - skapa nytt

				updatera to värdet till nu

				Om något värde saknas från sensorlocation eller datastore SKIT I ALLT!

				//*/
		}
	}()
	log.Printf(" [*] Waiting for stuffs. To exit press CTRL+C")
	<-forever

}

/*func (i *Interval) Print() {
	fmt.Printf("%v\n%v\n%v\n%v\n", i.Sensorid, i.Zone.String(), i.Deviceid, i.Id.String())
}*/
