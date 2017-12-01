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
	"io/ioutil"
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
	zone string `bson:"zone" json:"zone"`
	rssi uint32 `bson:"rssi" json:"rssi"`

}
type Zone struct {
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string `json:"name"`
	Location []float64 `json:"location"`
	Children []bson.ObjectId	 `json:"children" bson:"children,omitempty"`
	Parent bson.ObjectId `json:"parent" bson:"parent,omitempty"`
}
type datastore struct {
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
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename/:from/:to"), getIntervalByZoneNameDuringInterval(instances))
	mux.HandleFunc(pat.Get("/intervals/zones/:zonename/:at"), getIntervalForZoneByTime(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:deviceid"), getIntervalByDeviceID(instances))
	mux.HandleFunc(pat.Get("/intervals/device/:zonename"), getIntervalForDeviceInZone(instances))
	mux.HandleFunc(pat.Get("/"), healthCheck(instances))
	fmt.Printf("Starting Router\n")
    http.ListenAndServe("0.0.0.0:8080", mux)

}
/*func (i *Instances) readFromZones(sensorid string) bson.ObjectId{

	address,_, err := i.Consul.Catalog.Service("sensorlocation")
	resp, err := http.Get(address[0]+"/sensors/"+sensorid)
	if err != nil {
		failOnError(err, "Read from Zones failed")
	}
	defer resp.Body.Close()
	
	var m SensorLocation
	body, err := ioutil.ReadAll(resp.Body)
	errjson :=json.Unmarshal(body,&m)
	if errjson != nil {
		failOnError(err, "Decode from Zones failed")
	}
	return m.zoneid
}
func (i *Instances) readFromZonesMeta(zoneid string) string{
	address,_, err := i.Consul.Catalog.Service("zones")
	resp, err := http.Get(address[0]+"/zones/"+zoneid)
	if err != nil {
		failOnError(err, "Read from Zones failed")
	}
	defer resp.Body.Close()
	
	var m Zone
	body, err := ioutil.ReadAll(resp.Body)
	errjson :=json.Unmarshal(body,&m)
	if errjson != nil {
		failOnError(err, "Decode from Zones failed")
	}
	return m.Name
}
func (i *Instances) readFromDataStore() datastore{
	address,_, err := i.Consul.Catalog.Service("datastore")
	resp, err := http.Get(address[0]+"/sensor")
	if err != nil {
		failOnError(err, "Read from Datastore failed")
	}
	defer resp.Body.Close()
	
	var m datastore
	body, err := ioutil.ReadAll(resp.Body)
	errjson :=json.Unmarshal(body,&m)
	if errjson != nil {
		failOnError(err, "Decode from Datastore failed")
	}
	return m
}*/
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

		name := bson.ObjectIdHex(pat.Param(r, "zonename"))
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
		zonename := bson.ObjectIdHex(pat.Param(r, "zonename"))

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
		zonename := bson.ObjectIdHex(pat.Param(r, "zonename"))

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