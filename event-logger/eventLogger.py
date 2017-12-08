import time
import pika
import hashlib
import serial
import consul
import json
import sys

def main(args):
    

    print "Establishing consul connection"
    print "Address: " + args[1]
    con = consul.Consul(args[1])

    (index, response) = con.catalog.service('rabbit')
    jsondump = json.dumps(response[0])
    service = json.loads(jsondump)

    rabbitAddress = service['TaggedAddresses']['wan']
    rabbitPort = service['ServicePort']

    lookupTable = {}

    print "Establishing rabbit connection with: " + rabbitAddress + ':' + str(rabbitPort)
    connection = pika.BlockingConnection(pika.ConnectionParameters(rabbitAddress, rabbitPort))
    channel = connection.channel()
    channel.exchange_declare(exchange='event',
     exchange_type='topic',
     durable=True)

    result = channel.queue_declare(exclusive=True)

    queue_name = result.method.queue

    channel.queue_bind(exchange='event',
                       queue=queue_name,
                       routing_key=args[2])


    channel.basic_consume(callback,
                      queue=queue_name,
                      no_ack=True)

    channel.start_consuming()

def callback(ch, method, properties, body):
    print("%r\n%r\n\n" % (method.routing_key, body))
    
main(sys.argv) 