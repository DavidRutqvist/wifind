import time
import pika
import hashlib
import serial
import consul
import json

def main():
    ser = serial.Serial()
    ser.baudrate = 115200
    ser.port = 'COM6'

    con = consul.Consul()

    (index, response) = con.catalog.service('rabbit')
    jsondump = json.dumps(response[0])
    service = json.loads(jsondump)

    rabbitAddress = service['TaggedAddresses']['wan']
    rabbitPort = service['ServicePort']

    connection = pika.BlockingConnection(pika.ConnectionParameters(rabbitAddress, rabbitPort))
    channel = connection.channel()
    queue = 'sensorData'
    channel.queue_declare(queue=queue)
    

    ser.open()


    for i in range(0, 1000):
        a = ser.readline()
        b = a.split(';')
        if len(b) == 3:
            c = b[2].split('\r')
            channel.basic_publish(exchange='',
                      routing_key=queue,
                      body= hashlib.sha256(b[0]).hexdigest() + ';' + b[1] + ';' + hashlib.sha256(c[0]).hexdigest() + ';' + str(time.time()))

    ser.close()
    connection.close()

main()