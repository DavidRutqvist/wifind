import random
from random import randint
import datetime
import hashlib
import json
import urllib.request

def hash_sensor(sensorID):
    return hashlib.sha256(sensorID.encode()).hexdigest()
    
def hash_device(deviceID):
    return hashlib.sha256(deviceID.encode()).hexdigest()

def main(amountSensors, amountDevices, amountConnections):
  for i in range(0,amountConnections):
    rand_sensor = randint(0,amountSensors-1)
    rand_device = randint(amountSensors+1,amountSensors+amountDevices)
    rand_RSSI = randint(-90, -30)
    device_hash = hash_device(str(rand_device))
    sensor_hash = hash_sensor(str(rand_sensor))
    payload = {'device': device_hash, 'sensor' : sensor_hash, 'rssi' : rand_RSSI}
    print(payload)

    #myurl = "http://www.testmycode.com"
    #req = urllib.request.Request(myurl)
    #req.add_header('Content-Type', 'application/json; charset=utf-8')
    #jsondata = json.dumps(payload)
    #jsondataasbytes = jsondata.encode('utf-8')   # needs to be bytes
    #req.add_header('Content-Length', len(jsondataasbytes))
    #print (jsondataasbytes)
    #response = urllib.request.urlopen(req, jsondataasbytes) 

main(4, 10, 40) #sensors, devices, connections
