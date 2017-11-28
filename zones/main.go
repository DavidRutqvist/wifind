package main

import (  
    "encoding/json"
    "fmt"
    "log"
	"net/http"
	"net/http/httputil"
	"os"
	"bytes"
    "goji.io"
    "goji.io/pat"
    "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
)

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {  
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(json)
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
	  	panic(err)
	}
}

type PostRes struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

type Zone struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string `json:"name"`
	Location []float64 `json:"location"`
	Parent bson.ObjectId `json:"parent" bson:"parent,omitempty"`
}

type Instances struct {
	Session *mgo.Session
	Consul *api.Client
	RabbitConn *amqp.Connection
	RabbitChan *amqp.Channel
	ExchangeTopic string
	RabbitEndpoint string
}

func createInstances(mongoAddress string, consulAddress string, exchangedTopic string) *Instances {
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
	failOnError(err, "Failed to retrive RabbitMQ from Consul\n")

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

	return instances
}

func main() {
	mongoAddress := os.Getenv("MONGO_ADDRESS")
	consulAddress := os.Getenv("CONSUL_ADDRESS")
	exchangedTopic := os.Getenv("EXCHANGE_TOPIC")

	//instances := createInstances(mongoAddress, "srv.wifind.se:8500", "event")
	instances := createInstances(mongoAddress, consulAddress, exchangedTopic)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/zones"), allZones(instances))
	mux.HandleFunc(pat.Post("/zones"), addZone(instances))
	mux.HandleFunc(pat.Put("/zones/:zoneid"), updateZone(instances))
	mux.HandleFunc(pat.Get("/zones/:zoneid"), zoneByZoneID(instances))
	mux.HandleFunc(pat.Get("/zones/:zoneid/children"), getChildrenFromId(instances))
	mux.HandleFunc(pat.Get("/zones"), healthCheck(instances))
	fmt.Printf("Starting Router\n")
    http.ListenAndServe("0.0.0.0:8080", mux)
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

func allZones(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
    return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))
		
		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		session := i.Session.Copy()
        defer session.Close()

        c := session.DB("store").C("zones")

        zones := make([]Zone, 0)
		err = c.Find(bson.M{"parent": bson.M{"$exists": false}}).All(&zones)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
            return
        }
		
		respBody, err = json.MarshalIndent(zones, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
        
        ResponseWithJSON(w, respBody, http.StatusOK)
    }
}

func getChildrenFromId(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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

        c := session.DB("store").C("zones")

		var zones []Zone
		err = c.Find(bson.M{"parent": zoneid}).All(&zones)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if len(zones) == 0 {
			response = PostRes{Success: false, Message: "No children found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

		

        respBody, err = json.MarshalIndent(zones, "", "  ")
        if err != nil {
            log.Fatal(err)
        }

        ResponseWithJSON(w, respBody, http.StatusOK)
    }
}

func addZone(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
    return func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, false)
		fmt.Println(string(requestDump))
		failOnError(err, string(requestDump))

        session := i.Session.Copy()
		defer session.Close()
		
		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte
		var zone Zone
		
        decoder := json.NewDecoder(r.Body)
        err = decoder.Decode(&zone)
        if err != nil {
			response = PostRes{Success: false, Message: "Incorrect body"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusBadRequest)
            return
        }

		c := session.DB("store").C("zones")

        err = c.Insert(zone)
        if err != nil {
            if mgo.IsDup(err) {
				response = PostRes{Success: false, Message: "Zone with this ID already exists"}
				json.NewEncoder(buffer).Encode(response)
				respBody := buffer.Bytes()
				ResponseWithJSON(w, respBody, http.StatusBadRequest)
				return
			}
			panic(err)
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
		
		json.NewEncoder(buffer).Encode(zone)
		body := buffer.Bytes()
		topic := "zones." + zone.Id.Hex() + ".new"

		fmt.Printf("Publishing ZONE_CREATED event to %v\n", topic)
		i.PublishEvent(topic, body)
		
		buffer = new(bytes.Buffer)

		response = PostRes{Success: true, Message: "Zone created"}
		json.NewEncoder(buffer).Encode(response)
		respBody = buffer.Bytes()
		ResponseWithJSON(w, respBody, http.StatusCreated)
    }
}

func zoneByZoneID(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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

        c := session.DB("store").C("zones")

        var zone Zone
        err = c.Find(bson.M{"_id": zoneid}).One(&zone)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if zone.Id == "" {
			response = PostRes{Success: false, Message: "Zone not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

        respBody, err = json.MarshalIndent(zone, "", "  ")
        if err != nil {
            log.Fatal(err)
        }

        ResponseWithJSON(w, respBody, http.StatusOK)
    }
}

func updateZone(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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

        var zone Zone
        decoder := json.NewDecoder(r.Body)
        err = decoder.Decode(&zone)
        if err != nil {
            response = PostRes{Success: false, Message: "Incorrect body"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusBadRequest)
            return
        }

        c := session.DB("store").C("zones")
		zone.Id = zoneid
		err = c.Update(bson.M{"_id": zoneid}, &zone)
		
        if err != nil {
            switch err {
			default:
				panic(err)
				response = PostRes{Success: false, Message: "Database error"}
				json.NewEncoder(buffer).Encode(response)
				respBody := buffer.Bytes()
				ResponseWithJSON(w, respBody, http.StatusInternalServerError)
                return
            case mgo.ErrNotFound:
				response = PostRes{Success: false, Message: "Zone not found"}
				json.NewEncoder(buffer).Encode(response)
				respBody := buffer.Bytes()
				ResponseWithJSON(w, respBody, http.StatusInternalServerError)
                return
            }
        }
		
		err = c.Find(bson.M{"_id": zoneid}).One(&zone)
		json.NewEncoder(buffer).Encode(zone)
		body := buffer.Bytes()
		topic := "zones." + zone.Id.Hex() + ".updated"
		
		fmt.Printf("Publishing ZONE_UPDATED event to %v\n", topic)
		i.PublishEvent(topic, body)

		buffer = new(bytes.Buffer)

		response = PostRes{Success: true, Message: "Zone updated"}
		json.NewEncoder(buffer).Encode(response)
		respBody = buffer.Bytes()
		ResponseWithJSON(w, respBody, http.StatusOK)
    }
}

func (i *Instances) PublishEvent(topic string, body []byte) {
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

	err = channel.Publish(
		i.ExchangeTopic,	// exchange
		topic,				// topic
		false,
		false,
		amqp.Publishing{
			ContentType:	"text/plain",
			Body:			body,
	})
	failOnError(err, "Failed to publish event.")
}