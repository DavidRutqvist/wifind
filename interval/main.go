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
	"time"
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
	Success bool `json:"success"`
	Message string `json:"message"`
}
type SensorLocation struct {
  sensorId string `json:"id" bson:"_id, omitempty"`
  zoneid bson.ObjectId `json:"zoneid" bson:"zoneid"`
  from time.Time `json:"from" bson:"from"`
  to time.Time `json:"to" bson:"to"`
}

type Interval struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	deviceid string `bson:"devicemac" json:"devicemac"`
	sensorid string `bson:"sensormac" json:"sensormac"`
	from time.Time `bson:"from" json:"from"` 
	to time.Time `bson:"to" json:"to"`
	zone bson.ObjectId `bson:"zone" json:"zone"`
	rssi int32 `bson:"rssi" json:"rssi"`

}
type Zone struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string `json:"name"`
	Location []float64 `json:"location"`
	Children []bson.ObjectId	 `json:"children" bson:"children,omitempty"`
	Parent bson.ObjectId `json:"parent" bson:"parent,omitempty"`
}
type Datastore struct {
	device string `json:"device" bson:"device"`
	sensor string `json:"sensor" bson:"sensor"`
	rssi int32 `json:"rssi" bson:"rssi"`
	time time.Time `json:"time" bson:"time"`
}
type Instances struct {
	Session *mgo.Session
	Consul *api.Client
	RabbitConn *amqp.Connection
	RabbitChan *amqp.Channel
	ExchangeTopic string
	RabbitEndpoint string
	SubscriptionTopic string
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
	instances.SubscriptionTopic = SubscriptionTopic

	
	return instances
}

func main() {
	mongoAddress := os.Getenv("MONGO_ADDRESS")
	//consulAddress := os.Getenv("CONSUL_ADDRESS")
	consulAddress := "srv.wifind.se:8500"
	//exchangedTopic := os.Getenv("EXCHANGE_TOPIC")
	exchangedTopic := "event"

	//instances := createInstances(mongoAddress, "srv.wifind.se:8500")
	instances := createInstances(mongoAddress, consulAddress, exchangedTopic, "sensor.*.detected.*")

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/intervals"), allIntervals(instances))
	mux.HandleFunc(pat.Get("/intervals/:at"), getIntervalByTime(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename"), getIntervalByZoneName(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename/:from/:to"), getIntervalByZoneNameDuringInterval(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename/:at"), getIntervalForZoneByTime(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:deviceid"), getIntervalByDeviceID(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:zonename"), getIntervalForDeviceInZone(instances))
	mux.HandleFunc(pat.Get("/"), healthCheck(instances))
	fmt.Printf("Starting Router\n")
	go instances.Recieve()
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
        var allIntervals []Interval
        var intervalsWithin []Interval
        err = c.Find(bson.M{}).All(&allIntervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		for i := 0; i < len(allIntervals); i++ {
			if allIntervals[i].from.Before(tm) && allIntervals[i].to.After(tm){
				intervalsWithin = append(intervalsWithin,allIntervals[i])
			}
		}
        if intervalsWithin[0].zone == "" {
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

        var intervals []Interval
        err = c.Find(bson.M{"_name": name}).All(&intervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if intervals[0].zone == "" {
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

		deviceid := bson.ObjectIdHex(pat.Param(r, "deviceid"))

        c := session.DB("store").C("intervals")

        var intervals []Interval
        err = c.Find(bson.M{"_deviceid": deviceid}).All(&intervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if intervals[0].deviceid == "" {
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

		name := (pat.Param(r, "zonename"))
		deviceid := bson.ObjectIdHex(pat.Param(r, "deviceid"))

        c := session.DB("store").C("intervals")

        var intervals []Interval
        err = c.Find(bson.M{"_name": name, "_deviceid": deviceid}).All(&intervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		
        if intervals[0].zone == "" {
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
		zonename := (pat.Param(r, "zonename"))

        c := session.DB("store").C("intervals")
        var timeIntervals []Interval
        var intervalsWithin []Interval
        err = c.Find(bson.M{"_zonename": zonename, }).All(&timeIntervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		for i := 0; i < len(timeIntervals); i++ {
			if timeIntervals[i].from.Before(tmto) && timeIntervals[i].to.After(tmfrom){
				intervalsWithin = append(intervalsWithin,timeIntervals[i])
			}
		}
        if intervalsWithin[0].zone == "" {
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
		zonename := (pat.Param(r, "zonename"))

        c := session.DB("store").C("intervals")
        var timeIntervals []Interval
        var intervalsWithin []Interval
        err = c.Find(bson.M{"_zonename": zonename, }).All(&timeIntervals)
        if err != nil {
			response = PostRes{Success: false, Message: "Database error"}
			json.NewEncoder(buffer).Encode(response)
			respBody := buffer.Bytes()
			ResponseWithJSON(w, respBody, http.StatusInternalServerError)
			return
        }
		for i := 0; i < len(timeIntervals); i++ {
			if timeIntervals[i].from.Before(tm) && timeIntervals[i].to.After(tm){
				intervalsWithin = append(intervalsWithin,timeIntervals[i])
			}
		}
        if intervalsWithin[0].zone == "" {
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
func (i *Instances) Recieve() {
        conn, err := amqp.Dial(i.RabbitEndpoint) //Localhost?
        failOnError(err, "Failed to connect to RabbitMQ")
        defer conn.Close()

        channel, err := conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer channel.Close()

        session := i.Session.Copy()
			defer session.Close()

        //var interval Interval'
		


		fmt.Println("d")
        err = channel.ExchangeDeclare(
                i.ExchangeTopic, // name
                "topic",      // type
                true,         // durable
                false,        // auto-deleted
                false,        // internal
                false,        // no-wait
                nil,          // arguments
        )
        failOnError(err, "Failed to declare an exchange")

        q, err := channel.QueueDeclare(
                "",    // name
                false, // durable
                false, // delete when usused
                true,  // exclusive
                false, // no-wait
                nil,   // arguments
        )
        failOnError(err, "Failed to declare a queue")

        err = channel.QueueBind(
                q.Name,       // queue name
                i.SubscriptionTopic,     // routing key
                i.ExchangeTopic, // exchange
                false,
                nil)
        failOnError(err, "Failed to bind a queue")
        

        msgs, err := channel.Consume(
                q.Name, // queue
                "",     // consumer
                true,   // auto ack
                false,  // exclusive
                false,  // no local
                false,  // no wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        go func() {
        	var datastore Datastore
			var sensorlocation SensorLocation
            for d := range msgs {
            	fmt.Println("sd")
            	fmt.Println(d)
            	
                decoder := json.NewDecoder(bytes.NewReader(d.Body))
		        err = decoder.Decode(&datastore)
				sensorid := datastore.sensor

				resp, err := http.Get("http://app.wifind.se:9999/api/sensors/"+sensorid)
				sensordecoder := json.NewDecoder(resp.Body)
				err = sensordecoder.Decode(&sensorlocation)
				if err != nil{
					failOnError(err, "didn't get data")

				}
				zoneid := sensorlocation.zoneid
				
				var interval *Interval = new(Interval)
				interval.sensorid = sensorid
				interval.zone = bson.ObjectId(zoneid)
				interval.deviceid = datastore.device
				interval.rssi = datastore.rssi



		        c := session.DB("store").C("intervals")
    			err = c.Insert(interval)
            	}	
        	}()
        	log.Printf(" [*] Waiting for stuffs. To exit press CTRL+C")
}
