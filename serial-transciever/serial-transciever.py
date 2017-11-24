import time
import pika
import hashlib
import serial
import consul
import json
import sys

def main(args):
    ser = serial.Serial()

    print "Configuring serial"
    if len(args) >= 3:
        print "Baudrate: " + args[2]
        print "Port: " + args[1]
        ser.baudrate = int(args[2])
        ser.port = args[1]
    else:
        print "Default config"
        print "Baudrate: 115200"
        print "Port: COM6"
        ser.baudrate = 115200
        ser.port = 'COM6'

    print "Establishing consul connection"
    if len(args) >= 4:
        print "Address: " + args[3]
        con = consul.Consul(args[3])
    else:
        print "Default config:"
        print "Address: localhost"
        con = consul.Consul()

    (index, response) = con.catalog.service('rabbit')
    jsondump = json.dumps(response[0])
    service = json.loads(jsondump)

    rabbitAddress = service['TaggedAddresses']['wan']
    rabbitPort = service['ServicePort']

    lookupTable = {}

    print "Establishing rabbit connection with: " + rabbitAddress + ':' + str(rabbitPort)
    connection = pika.BlockingConnection(pika.ConnectionParameters(rabbitAddress, rabbitPort))
    channel = connection.channel()
    queue = 'sensor_data'
    channel.queue_declare(queue=queue)

    print "Opening serial port"
    ser.open()

    while True:
        send = False
        a = ser.readline()
        b = a.split(';')
        if len(b) == 3:
            c = b[2].split('\r')
            timestamp = int(time.time())
            deviceID = hashlib.sha256(b[0]).hexdigest()
            rssi = b[1]
            sensorID = hashlib.sha256(c[0]).hexdigest()
            if(not lookupTable.has_key(deviceID)):
                lookupTable[deviceID] = timestamp
                send = True
            elif(lookupTable[deviceID] + 5 < timestamp):
                lookupTable[deviceID] = timestamp
                send = True
            if(send):
                body = "{\"device\":\"" + deviceID + "\",\"sensor\":\"" + sensorID + "\",\"rssi\":" + rssi + ",\"time\":" + str(timestamp) + "}"
                print body
                print ("Unique entries: " + str(len(lookupTable.keys())))
                channel.basic_publish(exchange='',
                      routing_key=queue,
                      body= body)

    ser.close()
    connection.close()

main(sys.argv)