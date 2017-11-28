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
		fmt.Printf(msg)
	  	panic(err)
	}
}
type PostRes struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

type Interval struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	DeviceMAC string `json:"devicemac"`
	SensorMAC string `json:"sensormac"`
	from uint64 `json:"from"` 
	to uint64 `json:"to"`
	zone string `json:zone`

}
type Zone struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string `json:"name"`
	Location []float64 `json:"location"`
	Children []bson.ObjectId	 `json:"children" bson:"children,omitempty"`
	Parent bson.ObjectId `json:"parent" bson:"parent,omitempty"`
}
type Instances struct {
	Session *mgo.Session
	Consul *api.Client
	RabbitConn *amqp.Connection
	RabbitChan *amqp.Channel
}
func createInstances(mongoAddress string, consulAddress string) *Instances {
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
	fmt.Printf("Connecting to RabbitMQ at: %v:%v\n", rabbitAddress, rabbitPort)
	connection, err := amqp.Dial(rabbitEndpoint)
	failOnError(err, "Failed to connect to RabbitMQ\n")
	
	fmt.Printf("Opening RabbitMQ channel\n")
	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel\n")

	instances.RabbitConn = connection
	instances.RabbitChan = channel

	
	return instances
}

func main() {
	mongoAddress := os.Getenv("MONGO_ADDRESS")
	consulAddress := os.Getenv("CONSUL_ADDRESS")

	//instances := createInstances(mongoAddress, "srv.wifind.se:8500")
	instances := createInstances(mongoAddress, consulAddress)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/intervals"), allIntervals(instances))
	mux.HandleFunc(pat.Get("/intervals/:at"), getIntervalByTime(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename"), getIntervalByZoneName(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename/:from/:to"), getIntervalByZoneIDDuringInterval(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:deviceid"), getIntervalByDeviceID(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:zonename"), getIntervalForDeviceInZone(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:at"), getIntervalForZoneByTime(instances))
	mux.HandleFunc(pat.Get("/"), healthCheck(instances))
	fmt.Printf("Starting Router\n")
    http.ListenAndServe("0.0.0.0:8080", mux)

}
func readFromZones(){
	resp, err := http.Get("http.se.com.muuuuuu.nuuuuuuuujust 	")
	if err != nil {
		failOnError(err, "Read from Zones failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
}
func readFromZonesMeta(){
	resp, err := http.Get("http://example.com/")
	if err != nil {
		failOnError(err, "Read from Zone Metadata failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
}
func readFromDataStore(){
	resp, err := http.Get("http://example.com/")
	if err != nil {
		failOnError(err, "Read from Datastore failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
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

func allIntevals(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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

		at := bson.ObjectIdHex(pat.Param(r, "at"))

        c := session.DB("store").C("intervals")
//IF TIMESTAMP.FROM << REQUESTED && TIMESTAMP.TO >> REQUESTED = intervallet överlappar den sökta tiden
        var interval Interval
        err = c.Find(bson.M{"_name": name}).One(&interval)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if interval.name == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}
//IF TIMESTAMP.FROM << REQUESTED && TIMESTAMP.TO >> REQUESTED = intervallet överlappar den sökta tiden
        respBody, err = json.MarshalIndent(interval, "", "  ")
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

		name := bson.ObjectIdHex(pat.Param(r, "zonename"))

        c := session.DB("store").C("intervals")

        var interval Interval
        err = c.Find(bson.M{"_name": name}).One(&interval)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if interval.name == "" {
			response = PostRes{Success: false, Message: "Interval not found"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
		}

        respBody, err = json.MarshalIndent(interval, "", "  ")
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
		buffer := new(bytes.Buffer)
		var response PostRes
		var respBody []byte

		session := i.Session.Copy()
        defer session.Close()

        c := session.DB("store").C("intervals")
	}
}
func getIntervalByZoneIDDuringInterval(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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
	}
}
func getIntervalForZoneByTime(i *Instances) func(w http.ResponseWriter, r *http.Request) {  
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
	}
}